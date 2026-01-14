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
	Students      ports.StudentRepository
	Plans         ports.PlanRepository
	Subscriptions ports.SubscriptionRepository
	Payments      ports.PaymentRepository
	Reports       ports.ReportRepository
	Users         ports.UserRepository
}

func New(deps Dependencies) *Services {
	return &Services{
		Students:      NewStudentService(deps.Students),
		Plans:         NewPlanService(deps.Plans),
		Subscriptions: NewSubscriptionService(deps.Subscriptions, deps.Plans, deps.Students),
		Payments:      NewPaymentService(deps.Payments),
		Reports:       NewReportService(deps.Reports),
		Auth:          NewAuthService(deps.Users),
	}
}
