package ports

import (
	"context"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
)

type StudentRepository interface {
	Create(ctx context.Context, student domain.Student) (domain.Student, error)
	Update(ctx context.Context, student domain.Student) (domain.Student, error)
	FindByID(ctx context.Context, id string) (domain.Student, error)
	Search(ctx context.Context, filter StudentFilter) ([]domain.Student, error)
}

type StudentFilter struct {
	Query    string
	Statuses []domain.StudentStatus
	Limit    int
	Offset   int
}

type PlanRepository interface {
	Create(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	Update(ctx context.Context, plan domain.Plan) (domain.Plan, error)
	FindByID(ctx context.Context, id string) (domain.Plan, error)
	ListActive(ctx context.Context) ([]domain.Plan, error)
}

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	Update(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	FindByID(ctx context.Context, id string) (domain.Subscription, error)
	ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error)
	ListByPlan(ctx context.Context, planID string) ([]domain.Subscription, error)
	ListDueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error)
	ListAutoRenew(ctx context.Context) ([]domain.Subscription, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	Update(ctx context.Context, payment domain.Payment) (domain.Payment, error)
	FindByID(ctx context.Context, id string) (domain.Payment, error)
	FindByIdempotencyKey(ctx context.Context, key string) (domain.Payment, error)
	ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error)
	ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error)
}

type BillingPeriodRepository interface {
	Create(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error)
	Update(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error)
	ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error)
	ListOpenBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error)
	MarkOverdue(ctx context.Context, subscriptionID string, now time.Time) error
}

type PaymentAllocationRepository interface {
	Create(ctx context.Context, allocation domain.PaymentAllocation) error
	ListByPayment(ctx context.Context, paymentID string) ([]domain.PaymentAllocation, error)
	DeleteByPayment(ctx context.Context, paymentID string) error
}

type SubscriptionBalanceRepository interface {
	Get(ctx context.Context, subscriptionID string) (domain.SubscriptionBalance, error)
	Set(ctx context.Context, balance domain.SubscriptionBalance) (domain.SubscriptionBalance, error)
	Add(ctx context.Context, subscriptionID string, delta int64) (domain.SubscriptionBalance, error)
}

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

type ReportRepository interface {
	RevenueByPeriod(ctx context.Context, start, end time.Time) (RevenueSummary, error)
	StudentsByStatus(ctx context.Context) ([]StudentStatusSummary, error)
	DelinquentSubscriptions(ctx context.Context, now time.Time) ([]DelinquentSubscription, error)
	UpcomingDue(ctx context.Context, start, end time.Time) ([]DueSubscription, error)
}

type AuditRepository interface {
	Record(ctx context.Context, event domain.AuditEvent) error
}

type RevenueSummary struct {
	Start      time.Time
	End        time.Time
	TotalCents int64
}

type StudentStatusSummary struct {
	Status domain.StudentStatus
	Total  int64
}

type DelinquentSubscription struct {
	SubscriptionID string
	StudentID      string
	PlanID         string
	EndDate        time.Time
	DaysOverdue    int
}

type DueSubscription struct {
	SubscriptionID string
	StudentID      string
	PlanID         string
	EndDate        time.Time
}
