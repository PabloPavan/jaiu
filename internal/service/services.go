package service

import "github.com/PabloPavan/jaiu/internal/ports"

type Services struct {
	Students      *StudentService
	Plans         *PlanService
	Subscriptions *SubscriptionService
	Payments      *PaymentService
	Reports       *ReportService
	Auth          *AuthService
}

type Dependencies struct {
	Students       ports.StudentRepository
	Plans          ports.PlanRepository
	Subscriptions  ports.SubscriptionRepository
	Payments       ports.PaymentRepository
	BillingPeriods ports.BillingPeriodRepository
	Balances       ports.SubscriptionBalanceRepository
	Allocations    ports.PaymentAllocationRepository
	PaymentTx      ports.PaymentTxRunner
	Reports        ports.ReportRepository
	Users          ports.UserRepository
	Audit          ports.AuditRepository
}

func New(deps Dependencies) *Services {
	return &Services{
		Students:      NewStudentService(deps.Students, deps.Subscriptions, deps.Audit),
		Plans:         NewPlanService(deps.Plans, deps.Subscriptions, deps.Audit),
		Subscriptions: NewSubscriptionService(deps.Subscriptions, deps.Plans, deps.Students, deps.Audit),
		Payments:      NewPaymentService(deps.Payments, deps.Subscriptions, deps.Plans, deps.BillingPeriods, deps.Balances, deps.Allocations, deps.Audit, deps.PaymentTx),
		Reports:       NewReportService(deps.Reports),
		Auth:          NewAuthService(deps.Users, deps.Audit),
	}
}
