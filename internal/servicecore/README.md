# Service Core Layer

This folder is structured to mirror the `matching-service/internal/servicecore` pattern.
Use this layer for service-local middleware, health primitives, and cross-cutting policies.
Shared libraries should be imported through `pkg/wrapper` where appropriate.
