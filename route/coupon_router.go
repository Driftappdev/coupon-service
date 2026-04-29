package route

import (
	adminhandler "coupon-service/internal/adapter/http/admin_service"
	publichandler "coupon-service/internal/adapter/http/claim-query"
	servicecore "coupon-service/internal/servicecore"
	adminmiddleware "coupon-service/pkg/wrapper/adminmiddleware"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	r chi.Router,
	publicH *publichandler.CouponHTTPHandler,
	adminH *adminhandler.CouponAdminHTTPHandler,
) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(servicecore.HealthController("coupon-service", "v1"))
		})

		r.Route("/coupons", func(r chi.Router) {
			r.Get("/", publicH.ListCoupons)
			r.Post("/{code}/claim", publicH.ClaimCoupon)
		})

		r.Route("/admin/coupons", func(r chi.Router) {
			r.Use(adminmiddleware.ChiRequireRoles("SUPER_ADMIN", "ADMIN", "EDITOR"))
			r.Post("/", adminH.CreateCoupon)
			r.Put("/{code}", adminH.UpdateCoupon)
			r.Delete("/{code}", adminH.DeactivateCoupon)
		})
	})
}
