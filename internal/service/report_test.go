package service

import (
	"context"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/ports"
)

// Testa delegacao do ReportService para o repositorio.
func TestReportServiceDelegates(t *testing.T) {
	repo := &reportRepoFake{
		revenue: ports.RevenueSummary{TotalCents: 1000},
		statuses: []ports.StudentStatusSummary{
			{Status: "active", Total: 2},
		},
		delinquents: []ports.DelinquentSubscription{{SubscriptionID: "sub-1"}},
		upcoming:    []ports.DueSubscription{{SubscriptionID: "sub-2"}},
	}
	service := NewReportService(repo)

	if got, err := service.RevenueByPeriod(context.Background(), time.Now(), time.Now()); err != nil || got.TotalCents != 1000 {
		t.Fatalf("unexpected revenue response: %#v err=%v", got, err)
	}
	if got, err := service.StudentsByStatus(context.Background()); err != nil || len(got) != 1 {
		t.Fatalf("unexpected students status response: %#v err=%v", got, err)
	}
	if got, err := service.DelinquentSubscriptions(context.Background(), time.Now()); err != nil || len(got) != 1 {
		t.Fatalf("unexpected delinquents response: %#v err=%v", got, err)
	}
	if got, err := service.UpcomingDue(context.Background(), time.Now(), time.Now()); err != nil || len(got) != 1 {
		t.Fatalf("unexpected upcoming response: %#v err=%v", got, err)
	}
}
