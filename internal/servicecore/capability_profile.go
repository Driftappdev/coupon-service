package servicecore

// CapabilityProfile documents enterprise package targets for this service.
type CapabilityProfile struct {
	ServiceName         string
	InboundHTTP         bool
	InboundGRPC         bool
	InboundEvent        bool
	OutboundGRPCClient  bool
	OutboundEvent       bool
	HasDBOrMigration    bool
	RecommendedPackages []string
}

func EnterpriseCapabilityProfile() CapabilityProfile {
	return CapabilityProfile{
		ServiceName:         "coupon-service",
		InboundHTTP:         true,
		InboundGRPC:         false,
		InboundEvent:        true,
		OutboundGRPCClient:  false,
		OutboundEvent:       true,
		HasDBOrMigration:    true,
		RecommendedPackages: []string{"engine-core/messaging", "engine-core/messaging/idempotency", "engine-core/persistence", "engine-core/runtime/app", "middleware/auth", "middleware/metrics", "middleware/ratelimit", "middleware/recovery", "middleware/requestid", "middleware/retry", "middleware/timeout", "middleware/tracing", "middleware/validation", "tx/uow + inbox/outbox/dlq"},
	}
}
