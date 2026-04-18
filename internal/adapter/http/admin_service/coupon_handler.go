package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	dto "coupon-service/internal/dto/http/admin"
	httpport "coupon-service/internal/interface/http/admin"
	"coupon-service/internal/model"
	adminservice "coupon-service/internal/service_logic/service/admin"
)

type CouponAdminHTTPHandler struct {
	handler httpport.CouponPort
}

func NewCouponAdminHTTPHandler(
	handler httpport.CouponPort,
) *CouponAdminHTTPHandler {
	return &CouponAdminHTTPHandler{handler: handler}
}

//////////////////////////////////////////////////////
// Utility
//////////////////////////////////////////////////////

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}

func parseDiscountType(raw string) (model.DiscountType, error) {
	dt := model.DiscountType(raw)

	switch dt {
	case model.DiscountPercent, model.DiscountFixed:
		return dt, nil
	default:
		return "", errors.New("invalid discount type")
	}
}

//////////////////////////////////////////////////////
// POST /admin/coupons
//////////////////////////////////////////////////////

func (h *CouponAdminHTTPHandler) CreateCoupon(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()

	var req dto.CreateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	discountType, err := parseDiscountType(req.DiscountType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	coupon := model.Coupon{
		Code:          req.Code,
		DiscountType:  discountType,
		DiscountValue: req.DiscountValue,
		MinOrder:      req.MinOrder,
		MaxDiscount:   req.MaxDiscount,
		MaxUsage:      req.MaxUsage,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
	}

	err = h.handler.CreateCoupon(r.Context(), coupon)
	if err != nil {
		switch {
		case errors.Is(err, adminservice.ErrInvalidCoupon):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, adminservice.ErrCouponAlreadyExist):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "coupon created",
	})
}

//////////////////////////////////////////////////////
// PUT /admin/coupons/{code}
//////////////////////////////////////////////////////

func (h *CouponAdminHTTPHandler) UpdateCoupon(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close()

	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "coupon code required")
		return
	}

	var req dto.UpdateCouponRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	discountType, err := parseDiscountType(req.DiscountType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	coupon := model.Coupon{
		Code:          code,
		DiscountType:  discountType,
		DiscountValue: req.DiscountValue,
		MinOrder:      req.MinOrder,
		MaxDiscount:   req.MaxDiscount,
		MaxUsage:      req.MaxUsage,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
		Active:        req.Active,
	}

	err = h.handler.UpdateCoupon(r.Context(), coupon)
	if err != nil {
		switch {
		case errors.Is(err, adminservice.ErrInvalidCoupon):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, adminservice.ErrCouponNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "coupon updated",
	})
}

//////////////////////////////////////////////////////
// DELETE /admin/coupons/{code}
//////////////////////////////////////////////////////

func (h *CouponAdminHTTPHandler) DeactivateCoupon(
	w http.ResponseWriter,
	r *http.Request,
) {
	code := chi.URLParam(r, "code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "coupon code required")
		return
	}

	err := h.handler.DeactivateCoupon(r.Context(), code)
	if err != nil {
		switch {
		case errors.Is(err, adminservice.ErrInvalidCoupon):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, adminservice.ErrCouponNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
