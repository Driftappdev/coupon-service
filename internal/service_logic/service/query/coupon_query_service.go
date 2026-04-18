package query

import (
	"context"

	"coupon-service/internal/interface/repository"
	serviceport "coupon-service/internal/interface/service/query"
	"coupon-service/internal/model"
)

type CouponQueryService struct {
	repo repository.CouponRepository
}

func NewCouponQueryService(
	repo repository.CouponRepository,
) *CouponQueryService {
	return &CouponQueryService{
		repo: repo,
	}
}

var _ serviceport.CouponQueryService = (*CouponQueryService)(nil)

func (s *CouponQueryService) ListActiveCoupons(
	ctx context.Context,
) ([]model.Coupon, error) {
	return s.repo.FindAllActive(ctx)
}
