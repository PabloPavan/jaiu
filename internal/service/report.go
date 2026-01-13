package service

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/ports"
)

type ReportService struct {
	repo ports.ReportRepository
}

func NewReportService(repo ports.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) RevenueByPeriod(ctx context.Context, start, end time.Time) (ports.RevenueSummary, error) {
	return s.repo.RevenueByPeriod(ctx, start, end)
}

func (s *ReportService) StudentsByStatus(ctx context.Context) ([]ports.StudentStatusSummary, error) {
	return s.repo.StudentsByStatus(ctx)
}

func (s *ReportService) DelinquentSubscriptions(ctx context.Context, now time.Time) ([]ports.DelinquentSubscription, error) {
	return s.repo.DelinquentSubscriptions(ctx, now)
}

func (s *ReportService) UpcomingDue(ctx context.Context, start, end time.Time) ([]ports.DueSubscription, error) {
	return s.repo.UpcomingDue(ctx, start, end)
}
