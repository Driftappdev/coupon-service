package route

import (
	"context"
	publichandler "coupon-service/internal/adapter/http/claim-query"
	servicecore "coupon-service/internal/servicecore"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	publicH *publichandler.CouponHTTPHandler,
) {
	r.Post("/internal/admin/control", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		secret := strings.TrimSpace(os.Getenv("ADMIN_CONTROL_SHARED_SECRET"))
		if secret != "" && req.Header.Get("X-Admin-Secret") != secret {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unauthorized"})
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_json"})
			return
		}
		action, _ := body["action"].(string)
		if action == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "action_required"})
			return
		}
		_ = context.Background()
		_ = time.Now()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accepted": true,
			"service":  "coupon-service",
			"action":   action,
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(servicecore.HealthController("coupon-service", "v1"))
		})

		r.Route("/coupons", func(r chi.Router) {
			r.Get("/", publicH.ListCoupons)
			r.Post("/{code}/claim", publicH.ClaimCoupon)
		})

	})
}
