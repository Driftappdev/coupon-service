package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"coupon-service/config"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	grpcadapter "coupon-service/internal/adapter/grpc"
	adminhttp "coupon-service/internal/adapter/http/admin_service"
	publichttp "coupon-service/internal/adapter/http/claim-query"
	natsadmin "coupon-service/internal/adapter/nats_admin"
	natsadapter "coupon-service/internal/adapter/nats_claim"
	redisadapter "coupon-service/internal/adapter/ratelimit"
	repoadapter "coupon-service/internal/adapter/repository"
	outboxworker "coupon-service/internal/adapter/worker"

	grpcinfra "coupon-service/integration/grpc"
	httpinfra "coupon-service/integration/http"
	natsinfra "coupon-service/integration/nats"
	postgresinfra "coupon-service/integration/postgres"

	couponevent "coupon-service/internal/interface/coupon_event"
	adminhandler "coupon-service/internal/service_logic/handler/admin"
	grpchandler "coupon-service/internal/service_logic/handler/grpc"
	publichandler "coupon-service/internal/service_logic/handler/public"
	adminsvc "coupon-service/internal/service_logic/service/admin"
	calculatorsvc "coupon-service/internal/service_logic/service/calculator"
	claimsvc "coupon-service/internal/service_logic/service/claim"
	sagasvc "coupon-service/internal/service_logic/service/claim_saga"
	querysvc "coupon-service/internal/service_logic/service/query"
	servicecore "coupon-service/internal/servicecore"
	loggerwrapper "coupon-service/pkg/wrapper/logger"
	metricswrapper "coupon-service/pkg/wrapper/metrics"
	tracingwrapper "coupon-service/pkg/wrapper/tracing"

	pb "coupon-service/proto/pb/order_service"
	route "coupon-service/route"
	gotracing "github.com/driftappdev/libpackage/gotracing"
)

func main() {
	appLogger := loggerwrapper.New("coupon-service")
	gotracing.SetGlobalProvider(tracingwrapper.NewProvider("coupon-service"))

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	cfg := config.Load()
	features := config.LoadFeatureFlags()

	var (
		db          *sql.DB
		redisClient *redis.Client
		httpServer  *httpinfra.Server
		grpcServer  *grpcinfra.Server
	)

	if features.EnableDatabase {
		database, err := postgresinfra.NewPostgres(cfg.Database.DSN)
		if err != nil {
			appLogger.Fatalf("database init failed: %v", err)
		}
		db = database
	}

	if cfg.Redis.Addr != "" {
		rc := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := rc.Ping(ctx).Err(); err != nil {
			appLogger.Warnf("redis unavailable, saga rate limit disabled: %v", err)
			_ = rc.Close()
		} else {
			redisClient = rc
		}
	}

	var (
		couponRepo      *repoadapter.CouponRepository
		usageRepo       *repoadapter.UsageRepository
		outboxRepo      *repoadapter.OutboxRepository
		idempotencyRepo *repoadapter.IdempotencyRepository
	)
	if db != nil {
		couponRepo = repoadapter.NewCouponRepository(db)
		usageRepo = repoadapter.NewUsageRepository(db)
		outboxRepo = repoadapter.NewOutboxRepository(db)
		idempotencyRepo = repoadapter.NewIdempotencyRepository(db)
	}

	var calculatorService *calculatorsvc.CouponCalculatorService
	if features.EnableCalculator && couponRepo != nil {
		calculatorService = calculatorsvc.NewCouponCalculatorService(couponRepo)
	}

	var queryService *querysvc.CouponQueryService
	var claimService *claimsvc.CouponClaimService
	if features.EnableClaim && couponRepo != nil {
		queryService = querysvc.NewCouponQueryService(couponRepo)
		claimService = claimsvc.NewCouponClaimService(
			couponRepo,
			usageRepo,
			outboxRepo,
			idempotencyRepo,
		)
	}

	var adminService *adminsvc.CouponManagementService
	if features.EnableAdmin && couponRepo != nil {
		adminService = adminsvc.NewCouponManagementService(couponRepo, outboxRepo)
	}

	var publicFlow *publichandler.CouponPublicHandler
	if queryService != nil && claimService != nil {
		publicFlow = publichandler.NewCouponPublicHandler(queryService, claimService)
	}

	var adminFlow *adminhandler.CouponAdminHandler
	if adminService != nil {
		adminFlow = adminhandler.NewCouponAdminHandler(adminService)
	}

	var grpcFlow *grpchandler.CouponHandler
	if calculatorService != nil {
		grpcFlow = grpchandler.NewCouponHandler(calculatorService)
	}

	var publisher couponevent.CouponEventPublisher
	var natsJS nats.JetStreamContext
	if features.EnableNATSProducer || features.EnableNATSConsumer {
		nc, err := natsinfra.NewConnection(natsinfra.Config{
			URL:           cfg.NATS.URL,
			MaxReconnect:  10,
			ReconnectWait: 2 * time.Second,
			ClientName:    "coupon-service",
		})
		if err != nil {
			appLogger.Fatalf("nats connection failed: %v", err)
		}
		defer nc.Close()

		js, err := natsinfra.SetupJetStream(nc, natsinfra.StreamConfig{
			Name:     cfg.NATS.Stream,
			Subjects: []string{cfg.NATS.Subject},
			Replicas: 1,
		})
		if err != nil {
			appLogger.Fatalf("nats jetstream setup failed: %v", err)
		}
		natsJS = js

		if features.EnableNATSConsumer {
			if _, err := natsinfra.SetupJetStream(nc, natsinfra.StreamConfig{
				Name:     cfg.NATS.AdminStream,
				Subjects: []string{cfg.NATS.AdminSubject},
				Replicas: 1,
			}); err != nil {
				appLogger.Fatalf("nats admin stream setup failed: %v", err)
			}
		}

		if features.EnableNATSProducer {
			publisher = natsadapter.NewCouponEventPublisher(js, cfg.NATS.Subject)
		}
	}

	if features.EnableDatabase && outboxRepo != nil && publisher != nil {
		worker := outboxworker.NewWorker(
			db,
			outboxRepo,
			publisher,
			2*time.Second,
		)
		go worker.Start(ctx)
	}

	if features.EnableClaim &&
		couponRepo != nil &&
		usageRepo != nil &&
		idempotencyRepo != nil &&
		outboxRepo != nil &&
		publisher != nil &&
		redisClient != nil {
		sagaRepo := repoadapter.NewSagaRepository(db)
		rateLimiter := redisadapter.NewCouponClaimRateLimiter(
			redisClient,
			redisadapter.DefaultWindow,
			redisadapter.DefaultMaxRequests,
		)
		claimStep := sagasvc.NewClaimStep(couponRepo, usageRepo, idempotencyRepo)
		reserveStep := sagasvc.NewReserveStep(publisher)
		confirmStep := sagasvc.NewConfirmStep(couponRepo, outboxRepo)
		orchestrator := sagasvc.NewOrchestrator(
			sagaRepo,
			rateLimiter,
			[]sagasvc.SagaStep{claimStep, reserveStep, confirmStep},
		)
		recoveryWorker := sagasvc.NewRecoveryWorker(
			sagaRepo,
			orchestrator,
			30*time.Second,
			2*time.Minute,
		)
		go recoveryWorker.Start(ctx)
	}

	if features.EnableNATSConsumer && natsJS != nil && adminService != nil {
		adminConsumer := natsadmin.NewAdminCouponConsumer(
			natsJS,
			cfg.NATS.AdminStream,
			cfg.NATS.AdminSubject,
			cfg.NATS.AdminDurable,
			adminService,
		)
		go func() {
			if err := adminConsumer.Start(ctx); err != nil {
				appLogger.Errorf("admin nats consumer error: %v", err)
			}
		}()
	}

	if features.EnableHTTP && publicFlow != nil && adminFlow != nil {
		router := chi.NewRouter()
		router.Get("/metrics/app", metricswrapper.Handler())
		publicHandler := publichttp.NewCouponHTTPHandler(publicFlow)
		adminHandler := adminhttp.NewCouponAdminHTTPHandler(adminFlow)
		route.RegisterRoutes(router, publicHandler, adminHandler)

		var handler http.Handler = router
		middlewares := servicecore.DefaultHTTPMiddlewares()
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}

		httpServer = httpinfra.NewServer(normalizeAddress(cfg.HTTP.Port), handler)
		go func() {
			if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
				appLogger.Errorf("http server error: %v", err)
			}
		}()
	}

	if features.EnableCalculator && grpcFlow != nil {
		grpcHandler := grpcadapter.NewCouponGRPCHandler(grpcFlow)
		grpcServer = grpcinfra.NewServer(cfg.GRPC.Port)
		grpcServer.Register(func(s *grpc.Server) {
			pb.RegisterCouponServiceServer(s, grpcHandler)
		})
		go func() {
			if err := grpcServer.Start(); err != nil {
				appLogger.Errorf("grpc server error: %v", err)
			}
		}()
	}

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if grpcServer != nil {
		grpcServer.Stop(shutdownCtx)
	}

	if httpServer != nil {
		if err := httpServer.Stop(shutdownCtx); err != nil {
			appLogger.Errorf("http shutdown error: %v", err)
		}
	}

	if db != nil {
		if err := db.Close(); err != nil {
			appLogger.Errorf("db close error: %v", err)
		}
	}
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			appLogger.Errorf("redis close error: %v", err)
		}
	}
}

func normalizeAddress(address string) string {
	if strings.HasPrefix(address, ":") {
		return address
	}

	return ":" + address
}
