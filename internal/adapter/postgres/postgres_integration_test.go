//go:build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	fixtureStudentID      = "11111111-1111-1111-1111-111111111111"
	fixtureStudentTwoID   = "22222222-2222-2222-2222-222222222222"
	fixturePlanID         = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	fixtureSubscriptionID = "33333333-3333-3333-3333-333333333333"
	fixturePaymentID      = "55555555-5555-5555-5555-555555555555"
	fixturePeriodPaidID   = "66666666-6666-6666-6666-666666666666"
	fixturePeriodOpenID   = "88888888-8888-8888-8888-888888888888"
	fixtureUserEmail      = "admin@example.com"
)

func setupIntegration(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if os.Getenv("JAIU_INTEGRATION") != "true" {
		t.Skip("JAIU_INTEGRATION is not set to true")
	}

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := NewPool(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	t.Cleanup(pool.Close)

	resetDatabase(t, pool)
	loadFixtures(t, pool)
	return pool
}

func resetDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			payment_allocations,
			billing_periods,
			subscription_balances,
			payments,
			subscriptions,
			students,
			plans,
			users,
			audit_events,
			imagekit_outbox
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate database: %v", err)
	}
}

func loadFixtures(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	path := filepath.Join("..", "..", "..", "db", "fixtures", "test.sql")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := pool.Exec(ctx, string(content)); err != nil {
		t.Fatalf("load fixtures: %v", err)
	}
}

// Testa CRUD e busca de alunos com banco real.
func TestStudentRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewStudentRepository(pool)
	ctx := context.Background()

	student, err := repo.FindByID(ctx, fixtureStudentID)
	if err != nil {
		t.Fatalf("find student: %v", err)
	}
	if student.FullName != "Alice Example" {
		t.Fatalf("expected Alice Example, got %q", student.FullName)
	}

	results, err := repo.Search(ctx, ports.StudentFilter{
		Query:    "Alice",
		Statuses: []domain.StudentStatus{domain.StudentActive},
	})
	if err != nil {
		t.Fatalf("search students: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	created, err := repo.Create(ctx, domain.Student{
		FullName: "New Student",
		Status:   domain.StudentActive,
	})
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created student id")
	}

	created.Status = domain.StudentInactive
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("update student: %v", err)
	}
	if updated.Status != domain.StudentInactive {
		t.Fatalf("expected inactive status, got %q", updated.Status)
	}
}

// Testa CRUD e listagem de planos ativos.
func TestPlanRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewPlanRepository(pool)
	ctx := context.Background()

	plan, err := repo.FindByID(ctx, fixturePlanID)
	if err != nil {
		t.Fatalf("find plan: %v", err)
	}
	if !plan.Active {
		t.Fatal("expected active plan")
	}

	plans, err := repo.ListActive(ctx)
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 active plan, got %d", len(plans))
	}

	created, err := repo.Create(ctx, domain.Plan{
		Name:         "Premium",
		DurationDays: 30,
		PriceCents:   5000,
		Active:       true,
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	created.Active = false
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("update plan: %v", err)
	}
	if updated.Active {
		t.Fatal("expected plan to be inactive")
	}
}

// Testa consultas de assinaturas usando fixtures e criacao manual.
func TestSubscriptionRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewSubscriptionRepository(pool)
	ctx := context.Background()

	subscription, err := repo.FindByID(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("find subscription: %v", err)
	}
	if subscription.StudentID != fixtureStudentID {
		t.Fatalf("expected student %q, got %q", fixtureStudentID, subscription.StudentID)
	}

	byStudent, err := repo.ListByStudent(ctx, fixtureStudentID)
	if err != nil {
		t.Fatalf("list by student: %v", err)
	}
	if len(byStudent) != 1 {
		t.Fatalf("expected 1 subscription for student, got %d", len(byStudent))
	}

	byPlan, err := repo.ListByPlan(ctx, fixturePlanID)
	if err != nil {
		t.Fatalf("list by plan: %v", err)
	}
	if len(byPlan) != 1 {
		t.Fatalf("expected 1 subscription for plan, got %d", len(byPlan))
	}

	autoRenew, err := repo.ListAutoRenew(ctx)
	if err != nil {
		t.Fatalf("list auto renew: %v", err)
	}
	if len(autoRenew) != 1 {
		t.Fatalf("expected 1 auto renew subscription, got %d", len(autoRenew))
	}

	dueBetween, err := repo.ListDueBetween(ctx, time.Date(2024, 1, 30, 0, 0, 0, 0, time.UTC), time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("list due between: %v", err)
	}
	if len(dueBetween) != 1 {
		t.Fatalf("expected 1 due subscription, got %d", len(dueBetween))
	}

	created, err := repo.Create(ctx, domain.Subscription{
		StudentID:  fixtureStudentID,
		PlanID:     fixturePlanID,
		StartDate:  time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		EndDate:    time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		Status:     domain.SubscriptionActive,
		PriceCents: 1000,
		PaymentDay: 1,
		AutoRenew:  false,
	})
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created subscription id")
	}
}

// Testa CRUD e consultas de pagamentos.
func TestPaymentRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewPaymentRepository(pool)
	ctx := context.Background()

	payment, err := repo.FindByID(ctx, fixturePaymentID)
	if err != nil {
		t.Fatalf("find payment: %v", err)
	}
	if payment.IdempotencyKey != "idem-1" {
		t.Fatalf("expected idempotency key idem-1, got %q", payment.IdempotencyKey)
	}

	byKey, err := repo.FindByIdempotencyKey(ctx, "idem-1")
	if err != nil {
		t.Fatalf("find by idempotency: %v", err)
	}
	if byKey.ID != fixturePaymentID {
		t.Fatalf("expected payment %q, got %q", fixturePaymentID, byKey.ID)
	}

	bySub, err := repo.ListBySubscription(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("list by subscription: %v", err)
	}
	if len(bySub) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(bySub))
	}

	byPeriod, err := repo.ListByPeriod(ctx, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("list by period: %v", err)
	}
	if len(byPeriod) != 1 {
		t.Fatalf("expected 1 payment in period, got %d", len(byPeriod))
	}

	created, err := repo.Create(ctx, domain.Payment{
		SubscriptionID: fixtureSubscriptionID,
		PaidAt:         time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC),
		AmountCents:    1000,
		Method:         domain.PaymentPix,
		Status:         domain.PaymentConfirmed,
		Kind:           domain.PaymentFull,
		IdempotencyKey: "idem-2",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	created.Reference = "ref-updated"
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("update payment: %v", err)
	}
	if updated.Reference != "ref-updated" {
		t.Fatalf("expected updated reference, got %q", updated.Reference)
	}
}

// Testa billing periods com listagem, criacao, update e overdue.
func TestBillingPeriodRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewBillingPeriodRepository(pool)
	ctx := context.Background()

	periods, err := repo.ListBySubscription(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("list periods: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}

	open, err := repo.ListOpenBySubscription(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("list open periods: %v", err)
	}
	if len(open) != 1 {
		t.Fatalf("expected 1 open period, got %d", len(open))
	}
	if open[0].ID != fixturePeriodOpenID {
		t.Fatalf("expected open period %q, got %q", fixturePeriodOpenID, open[0].ID)
	}
	if open[0].ID == fixturePeriodPaidID {
		t.Fatalf("expected paid period %q to be excluded", fixturePeriodPaidID)
	}

	created, err := repo.Create(ctx, domain.BillingPeriod{
		SubscriptionID:  fixtureSubscriptionID,
		PeriodStart:     time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:       time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC),
		AmountDueCents:  1000,
		AmountPaidCents: 0,
		Status:          domain.BillingOpen,
	})
	if err != nil {
		t.Fatalf("create period: %v", err)
	}
	created.AmountPaidCents = 1000
	created.Status = domain.BillingPaid
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("update period: %v", err)
	}
	if updated.Status != domain.BillingPaid {
		t.Fatalf("expected paid status, got %q", updated.Status)
	}

	if err := repo.MarkOverdue(ctx, fixtureSubscriptionID, time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("mark overdue: %v", err)
	}
	openAfter, err := repo.ListOpenBySubscription(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("list open periods after overdue: %v", err)
	}
	if len(openAfter) != 1 {
		t.Fatalf("expected 1 open/overdue period, got %d", len(openAfter))
	}
	if openAfter[0].Status != domain.BillingOverdue {
		t.Fatalf("expected overdue status, got %q", openAfter[0].Status)
	}
}

// Testa saldo da assinatura com get, set e add.
func TestSubscriptionBalanceRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewSubscriptionBalanceRepository(pool)
	ctx := context.Background()

	balance, err := repo.Get(ctx, fixtureSubscriptionID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if balance.CreditCents != 0 {
		t.Fatalf("expected 0 balance, got %d", balance.CreditCents)
	}

	updated, err := repo.Add(ctx, fixtureSubscriptionID, 200)
	if err != nil {
		t.Fatalf("add balance: %v", err)
	}
	if updated.CreditCents != 200 {
		t.Fatalf("expected 200 credit, got %d", updated.CreditCents)
	}
}

// Testa alocacoes de pagamento.
func TestPaymentAllocationRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewPaymentAllocationRepository(pool)
	ctx := context.Background()

	allocations, err := repo.ListByPayment(ctx, fixturePaymentID)
	if err != nil {
		t.Fatalf("list allocations: %v", err)
	}
	if len(allocations) != 1 {
		t.Fatalf("expected 1 allocation, got %d", len(allocations))
	}

	if err := repo.DeleteByPayment(ctx, fixturePaymentID); err != nil {
		t.Fatalf("delete allocations: %v", err)
	}

	allocations, err = repo.ListByPayment(ctx, fixturePaymentID)
	if err != nil {
		t.Fatalf("list allocations after delete: %v", err)
	}
	if len(allocations) != 0 {
		t.Fatalf("expected 0 allocations, got %d", len(allocations))
	}

	if err := repo.Create(ctx, domain.PaymentAllocation{
		PaymentID:       fixturePaymentID,
		BillingPeriodID: fixturePeriodOpenID,
		AmountCents:     500,
	}); err != nil {
		t.Fatalf("create allocation: %v", err)
	}
}

// Testa CRUD de usuarios e find por email.
func TestUserRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	user, err := repo.FindByEmail(ctx, fixtureUserEmail)
	if err != nil {
		t.Fatalf("find by email: %v", err)
	}
	if user.Email != fixtureUserEmail {
		t.Fatalf("expected %q, got %q", fixtureUserEmail, user.Email)
	}

	created, err := repo.Create(ctx, domain.User{
		Name:         "User Two",
		Email:        "user2@example.com",
		PasswordHash: "hash2",
		Role:         domain.RoleOperator,
		Active:       true,
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created user id")
	}
}

// Testa gravacao de eventos de auditoria no banco.
func TestAuditRepositoryIntegration(t *testing.T) {
	pool := setupIntegration(t)
	repo := NewAuditRepository(pool)
	ctx := context.Background()

	event := domain.AuditEvent{
		Action:     "integration.test",
		EntityType: "student",
		EntityID:   fixtureStudentTwoID,
		ActorID:    fixtureStudentID,
		ActorRole:  "admin",
		Metadata:   map[string]any{"source": "integration"},
		IP:         "127.0.0.1",
		UserAgent:  "test",
	}
	if err := repo.Record(ctx, event); err != nil {
		t.Fatalf("record audit: %v", err)
	}

	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_events WHERE action = $1", "integration.test").Scan(&count)
	if err != nil {
		t.Fatalf("count audit events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 audit event, got %d", count)
	}
}
