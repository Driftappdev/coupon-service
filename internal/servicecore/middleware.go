package servicecore

import (
	"net/http"

	coremid "coupon-service/internal/servicecore/middleware"
)

func DefaultHTTPMiddlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		coremid.Context(),
		coremid.RateLimit(),
		coremid.Audit(),
		coremid.Permission(),
		coremid.IdempotencyHTTP(),
		coremid.Validation(),
	}
}
