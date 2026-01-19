package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type auditRepoFake struct {
	events []domain.AuditEvent
	err    error
}

func (f *auditRepoFake) Record(ctx context.Context, event domain.AuditEvent) error {
	f.events = append(f.events, event)
	return f.err
}

type userRepoFake struct {
	users     map[string]domain.User
	createErr error
	findErr   error
}

func (f *userRepoFake) Create(ctx context.Context, user domain.User) (domain.User, error) {
	if f.createErr != nil {
		return domain.User{}, f.createErr
	}
	if f.users == nil {
		f.users = make(map[string]domain.User)
	}
	if user.ID == "" {
		user.ID = fmt.Sprintf("user-%d", len(f.users)+1)
	}
	f.users[strings.ToLower(user.Email)] = user
	return user, nil
}

func (f *userRepoFake) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	if f.findErr != nil {
		return domain.User{}, f.findErr
	}
	if f.users == nil {
		return domain.User{}, ports.ErrNotFound
	}
	user, ok := f.users[strings.ToLower(email)]
	if !ok {
		return domain.User{}, ports.ErrNotFound
	}
	return user, nil
}

type studentRepoFake struct {
	students   map[string]domain.Student
	createErr  error
	updateErr  error
	findErr    error
	searchErr  error
	lastFilter ports.StudentFilter
}

func (f *studentRepoFake) Create(ctx context.Context, student domain.Student) (domain.Student, error) {
	if f.createErr != nil {
		return domain.Student{}, f.createErr
	}
	if f.students == nil {
		f.students = make(map[string]domain.Student)
	}
	if student.ID == "" {
		student.ID = fmt.Sprintf("student-%d", len(f.students)+1)
	}
	f.students[student.ID] = student
	return student, nil
}

func (f *studentRepoFake) Update(ctx context.Context, student domain.Student) (domain.Student, error) {
	if f.updateErr != nil {
		return domain.Student{}, f.updateErr
	}
	if f.students == nil {
		f.students = make(map[string]domain.Student)
	}
	if student.ID == "" {
		return domain.Student{}, errors.New("missing student id")
	}
	f.students[student.ID] = student
	return student, nil
}

func (f *studentRepoFake) FindByID(ctx context.Context, id string) (domain.Student, error) {
	if f.findErr != nil {
		return domain.Student{}, f.findErr
	}
	if f.students == nil {
		return domain.Student{}, ports.ErrNotFound
	}
	student, ok := f.students[id]
	if !ok {
		return domain.Student{}, ports.ErrNotFound
	}
	return student, nil
}

func (f *studentRepoFake) Search(ctx context.Context, filter ports.StudentFilter) ([]domain.Student, error) {
	f.lastFilter = filter
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	if f.students == nil {
		return nil, nil
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	statuses := make(map[domain.StudentStatus]struct{})
	for _, status := range filter.Statuses {
		statuses[status] = struct{}{}
	}
	results := make([]domain.Student, 0, len(f.students))
	for _, student := range f.students {
		if query != "" && !strings.Contains(strings.ToLower(student.FullName), query) {
			continue
		}
		if len(statuses) > 0 {
			if _, ok := statuses[student.Status]; !ok {
				continue
			}
		}
		results = append(results, student)
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}
	return results, nil
}

func (f *studentRepoFake) Count(ctx context.Context, filter ports.StudentFilter) (int, error) {
	countFilter := filter
	countFilter.Limit = 0
	countFilter.Offset = 0
	students, err := f.Search(ctx, countFilter)
	if err != nil {
		return 0, err
	}
	return len(students), nil
}

type planRepoFake struct {
	plans     map[string]domain.Plan
	createErr error
	updateErr error
	findErr   error
	listErr   error
}

func (f *planRepoFake) Create(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	if f.createErr != nil {
		return domain.Plan{}, f.createErr
	}
	if f.plans == nil {
		f.plans = make(map[string]domain.Plan)
	}
	if plan.ID == "" {
		plan.ID = fmt.Sprintf("plan-%d", len(f.plans)+1)
	}
	f.plans[plan.ID] = plan
	return plan, nil
}

func (f *planRepoFake) Update(ctx context.Context, plan domain.Plan) (domain.Plan, error) {
	if f.updateErr != nil {
		return domain.Plan{}, f.updateErr
	}
	if f.plans == nil {
		f.plans = make(map[string]domain.Plan)
	}
	if plan.ID == "" {
		return domain.Plan{}, errors.New("missing plan id")
	}
	f.plans[plan.ID] = plan
	return plan, nil
}

func (f *planRepoFake) FindByID(ctx context.Context, id string) (domain.Plan, error) {
	if f.findErr != nil {
		return domain.Plan{}, f.findErr
	}
	if f.plans == nil {
		return domain.Plan{}, ports.ErrNotFound
	}
	plan, ok := f.plans[id]
	if !ok {
		return domain.Plan{}, ports.ErrNotFound
	}
	return plan, nil
}

func (f *planRepoFake) ListActive(ctx context.Context) ([]domain.Plan, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.plans == nil {
		return nil, nil
	}
	results := make([]domain.Plan, 0, len(f.plans))
	for _, plan := range f.plans {
		if plan.Active {
			results = append(results, plan)
		}
	}
	return results, nil
}

type subscriptionRepoFake struct {
	subscriptions  map[string]domain.Subscription
	createErr      error
	updateErr      error
	findErr        error
	listStudentErr error
	listPlanErr    error
	listDueErr     error
	listAutoErr    error
}

func (f *subscriptionRepoFake) Create(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	if f.createErr != nil {
		return domain.Subscription{}, f.createErr
	}
	if f.subscriptions == nil {
		f.subscriptions = make(map[string]domain.Subscription)
	}
	if sub.ID == "" {
		sub.ID = fmt.Sprintf("subscription-%d", len(f.subscriptions)+1)
	}
	f.subscriptions[sub.ID] = sub
	return sub, nil
}

func (f *subscriptionRepoFake) Update(ctx context.Context, sub domain.Subscription) (domain.Subscription, error) {
	if f.updateErr != nil {
		return domain.Subscription{}, f.updateErr
	}
	if f.subscriptions == nil {
		f.subscriptions = make(map[string]domain.Subscription)
	}
	if sub.ID == "" {
		return domain.Subscription{}, errors.New("missing subscription id")
	}
	f.subscriptions[sub.ID] = sub
	return sub, nil
}

func (f *subscriptionRepoFake) FindByID(ctx context.Context, id string) (domain.Subscription, error) {
	if f.findErr != nil {
		return domain.Subscription{}, f.findErr
	}
	if f.subscriptions == nil {
		return domain.Subscription{}, ports.ErrNotFound
	}
	sub, ok := f.subscriptions[id]
	if !ok {
		return domain.Subscription{}, ports.ErrNotFound
	}
	return sub, nil
}

func (f *subscriptionRepoFake) ListByStudent(ctx context.Context, studentID string) ([]domain.Subscription, error) {
	if f.listStudentErr != nil {
		return nil, f.listStudentErr
	}
	return f.filter(func(sub domain.Subscription) bool {
		return sub.StudentID == studentID
	}), nil
}

func (f *subscriptionRepoFake) ListByPlan(ctx context.Context, planID string) ([]domain.Subscription, error) {
	if f.listPlanErr != nil {
		return nil, f.listPlanErr
	}
	return f.filter(func(sub domain.Subscription) bool {
		return sub.PlanID == planID
	}), nil
}

func (f *subscriptionRepoFake) ListDueBetween(ctx context.Context, start, end time.Time) ([]domain.Subscription, error) {
	if f.listDueErr != nil {
		return nil, f.listDueErr
	}
	return f.filter(func(sub domain.Subscription) bool {
		return !sub.EndDate.Before(start) && !sub.EndDate.After(end)
	}), nil
}

func (f *subscriptionRepoFake) ListAutoRenew(ctx context.Context) ([]domain.Subscription, error) {
	if f.listAutoErr != nil {
		return nil, f.listAutoErr
	}
	return f.filter(func(sub domain.Subscription) bool {
		return sub.AutoRenew
	}), nil
}

func (f *subscriptionRepoFake) filter(fn func(domain.Subscription) bool) []domain.Subscription {
	if f.subscriptions == nil {
		return nil
	}
	results := make([]domain.Subscription, 0, len(f.subscriptions))
	for _, sub := range f.subscriptions {
		if fn(sub) {
			results = append(results, sub)
		}
	}
	return results
}

type billingPeriodRepoFake struct {
	periods   map[string]domain.BillingPeriod
	createErr error
	updateErr error
	listErr   error
	openErr   error
}

func (f *billingPeriodRepoFake) Create(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error) {
	if f.createErr != nil {
		return domain.BillingPeriod{}, f.createErr
	}
	if f.periods == nil {
		f.periods = make(map[string]domain.BillingPeriod)
	}
	if period.ID == "" {
		period.ID = fmt.Sprintf("period-%d", len(f.periods)+1)
	}
	f.periods[period.ID] = period
	return period, nil
}

func (f *billingPeriodRepoFake) Update(ctx context.Context, period domain.BillingPeriod) (domain.BillingPeriod, error) {
	if f.updateErr != nil {
		return domain.BillingPeriod{}, f.updateErr
	}
	if f.periods == nil {
		f.periods = make(map[string]domain.BillingPeriod)
	}
	if period.ID == "" {
		return domain.BillingPeriod{}, errors.New("missing period id")
	}
	f.periods[period.ID] = period
	return period, nil
}

func (f *billingPeriodRepoFake) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.filter(func(period domain.BillingPeriod) bool {
		return period.SubscriptionID == subscriptionID
	}), nil
}

func (f *billingPeriodRepoFake) ListOpenBySubscription(ctx context.Context, subscriptionID string) ([]domain.BillingPeriod, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	return f.filter(func(period domain.BillingPeriod) bool {
		if period.SubscriptionID != subscriptionID {
			return false
		}
		return period.Status != domain.BillingPaid
	}), nil
}

func (f *billingPeriodRepoFake) MarkOverdue(ctx context.Context, subscriptionID string, now time.Time) error {
	return nil
}

func (f *billingPeriodRepoFake) filter(fn func(domain.BillingPeriod) bool) []domain.BillingPeriod {
	if f.periods == nil {
		return nil
	}
	results := make([]domain.BillingPeriod, 0, len(f.periods))
	for _, period := range f.periods {
		if fn(period) {
			results = append(results, period)
		}
	}
	return results
}

type paymentAllocationRepoFake struct {
	allocations map[string][]domain.PaymentAllocation
	createErr   error
	listErr     error
	deleteErr   error
}

func (f *paymentAllocationRepoFake) Create(ctx context.Context, allocation domain.PaymentAllocation) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.allocations == nil {
		f.allocations = make(map[string][]domain.PaymentAllocation)
	}
	f.allocations[allocation.PaymentID] = append(f.allocations[allocation.PaymentID], allocation)
	return nil
}

func (f *paymentAllocationRepoFake) ListByPayment(ctx context.Context, paymentID string) ([]domain.PaymentAllocation, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.allocations == nil {
		return nil, nil
	}
	return append([]domain.PaymentAllocation(nil), f.allocations[paymentID]...), nil
}

func (f *paymentAllocationRepoFake) DeleteByPayment(ctx context.Context, paymentID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.allocations, paymentID)
	return nil
}

type balanceRepoFake struct {
	balances map[string]domain.SubscriptionBalance
	getErr   error
	setErr   error
	addErr   error
}

func (f *balanceRepoFake) Get(ctx context.Context, subscriptionID string) (domain.SubscriptionBalance, error) {
	if f.getErr != nil {
		return domain.SubscriptionBalance{}, f.getErr
	}
	if f.balances == nil {
		return domain.SubscriptionBalance{}, ports.ErrNotFound
	}
	balance, ok := f.balances[subscriptionID]
	if !ok {
		return domain.SubscriptionBalance{}, ports.ErrNotFound
	}
	return balance, nil
}

func (f *balanceRepoFake) Set(ctx context.Context, balance domain.SubscriptionBalance) (domain.SubscriptionBalance, error) {
	if f.setErr != nil {
		return domain.SubscriptionBalance{}, f.setErr
	}
	if f.balances == nil {
		f.balances = make(map[string]domain.SubscriptionBalance)
	}
	f.balances[balance.SubscriptionID] = balance
	return balance, nil
}

func (f *balanceRepoFake) Add(ctx context.Context, subscriptionID string, delta int64) (domain.SubscriptionBalance, error) {
	if f.addErr != nil {
		return domain.SubscriptionBalance{}, f.addErr
	}
	if f.balances == nil {
		f.balances = make(map[string]domain.SubscriptionBalance)
	}
	balance := f.balances[subscriptionID]
	balance.SubscriptionID = subscriptionID
	balance.CreditCents += delta
	f.balances[subscriptionID] = balance
	return balance, nil
}

type paymentRepoFake struct {
	payments      map[string]domain.Payment
	byIdempotency map[string]string
	createErr     error
	updateErr     error
	findErr       error
	findIdemErr   error
	listSubErr    error
	listPeriodErr error
	updateCalls   int
}

func (f *paymentRepoFake) Create(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if f.createErr != nil {
		return domain.Payment{}, f.createErr
	}
	if f.payments == nil {
		f.payments = make(map[string]domain.Payment)
	}
	if f.byIdempotency == nil {
		f.byIdempotency = make(map[string]string)
	}
	if payment.IdempotencyKey != "" {
		if _, exists := f.byIdempotency[payment.IdempotencyKey]; exists {
			return domain.Payment{}, ports.ErrConflict
		}
	}
	if payment.ID == "" {
		payment.ID = fmt.Sprintf("payment-%d", len(f.payments)+1)
	}
	f.payments[payment.ID] = payment
	if payment.IdempotencyKey != "" {
		f.byIdempotency[payment.IdempotencyKey] = payment.ID
	}
	return payment, nil
}

func (f *paymentRepoFake) Update(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	f.updateCalls++
	if f.updateErr != nil {
		return domain.Payment{}, f.updateErr
	}
	if f.payments == nil {
		f.payments = make(map[string]domain.Payment)
	}
	f.payments[payment.ID] = payment
	return payment, nil
}

func (f *paymentRepoFake) FindByID(ctx context.Context, id string) (domain.Payment, error) {
	if f.findErr != nil {
		return domain.Payment{}, f.findErr
	}
	if f.payments == nil {
		return domain.Payment{}, ports.ErrNotFound
	}
	payment, ok := f.payments[id]
	if !ok {
		return domain.Payment{}, ports.ErrNotFound
	}
	return payment, nil
}

func (f *paymentRepoFake) FindByIdempotencyKey(ctx context.Context, key string) (domain.Payment, error) {
	if f.findIdemErr != nil {
		return domain.Payment{}, f.findIdemErr
	}
	if f.byIdempotency == nil {
		return domain.Payment{}, ports.ErrNotFound
	}
	id, ok := f.byIdempotency[key]
	if !ok {
		return domain.Payment{}, ports.ErrNotFound
	}
	return f.payments[id], nil
}

func (f *paymentRepoFake) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.Payment, error) {
	if f.listSubErr != nil {
		return nil, f.listSubErr
	}
	if f.payments == nil {
		return nil, nil
	}
	results := make([]domain.Payment, 0, len(f.payments))
	for _, payment := range f.payments {
		if payment.SubscriptionID == subscriptionID {
			results = append(results, payment)
		}
	}
	return results, nil
}

func (f *paymentRepoFake) ListByPeriod(ctx context.Context, start, end time.Time) ([]domain.Payment, error) {
	if f.listPeriodErr != nil {
		return nil, f.listPeriodErr
	}
	if f.payments == nil {
		return nil, nil
	}
	results := make([]domain.Payment, 0, len(f.payments))
	for _, payment := range f.payments {
		if payment.PaidAt.IsZero() {
			continue
		}
		if !payment.PaidAt.Before(start) && !payment.PaidAt.After(end) {
			results = append(results, payment)
		}
	}
	return results, nil
}

type reportRepoFake struct {
	revenue        ports.RevenueSummary
	revenueErr     error
	statuses       []ports.StudentStatusSummary
	statusesErr    error
	delinquents    []ports.DelinquentSubscription
	delinquentsErr error
	upcoming       []ports.DueSubscription
	upcomingErr    error
}

func (f *reportRepoFake) RevenueByPeriod(ctx context.Context, start, end time.Time) (ports.RevenueSummary, error) {
	return f.revenue, f.revenueErr
}

func (f *reportRepoFake) StudentsByStatus(ctx context.Context) ([]ports.StudentStatusSummary, error) {
	return f.statuses, f.statusesErr
}

func (f *reportRepoFake) DelinquentSubscriptions(ctx context.Context, now time.Time) ([]ports.DelinquentSubscription, error) {
	return f.delinquents, f.delinquentsErr
}

func (f *reportRepoFake) UpcomingDue(ctx context.Context, start, end time.Time) ([]ports.DueSubscription, error) {
	return f.upcoming, f.upcomingErr
}

type paymentTxRunnerFake struct {
	deps ports.PaymentDependencies
	err  error
}

func (f *paymentTxRunnerFake) RunSerializable(ctx context.Context, fn func(context.Context, ports.PaymentDependencies) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(ctx, f.deps)
}
