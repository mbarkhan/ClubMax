package service

import (
	"context"
	"github.com/clubmax/clubmax/internal/domain"
	"github.com/clubmax/clubmax/internal/repository"
	"strings"
)

type AdminService struct{ repo repository.AdminRepository }

func NewAdminService(repo repository.AdminRepository) *AdminService { return &AdminService{repo: repo} }

func (s *AdminService) Overview(ctx context.Context) (domain.AdminOverview, error) {
	return s.repo.AdminOverview(ctx)
}

func (s *AdminService) Gyms(ctx context.Context, search string, limit, offset int) ([]domain.AdminGym, error) {
	return s.repo.ListAdminGyms(ctx, strings.TrimSpace(search), limit, offset)
}

func (s *AdminService) Payments(ctx context.Context, search string, limit, offset int) ([]domain.AdminPayment, error) {
	return s.repo.ListAdminPayments(ctx, strings.TrimSpace(search), limit, offset)
}
