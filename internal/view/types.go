package view

import (
	"time"

	"github.com/a-h/templ"
)

type Page struct {
	Title       string
	Content     templ.Component
	CurrentUser *UserInfo
	Now         time.Time
}

type UserInfo struct {
	Name        string
	DisplayName string
	Role        string
}

type PlanItem struct {
	ID           string
	Name         string
	DurationDays int
	Price        string
	Description  string
}

type PlansPageData struct {
	Items []PlanItem
}

type PlanFormData struct {
	Title        string
	Action       string
	SubmitLabel  string
	DeleteAction string
	ShowDelete   bool
	Name         string
	DurationDays string
	Price        string
	Description  string
	Active       bool
	Error        string
}

type StudentItem struct {
	ID          string
	FullName    string
	BirthDate   string
	Phone       string
	Email       string
	Status      string
	StatusLabel string
	StatusClass string
}

type StudentsPageData struct {
	Query  string
	Status string
	Items  []StudentItem
}

type StudentFormData struct {
	Title        string
	Action       string
	SubmitLabel  string
	DeleteAction string
	ShowDelete   bool
	FullName     string
	BirthDate    string
	Gender       string
	Phone        string
	Email        string
	CPF          string
	Address      string
	Notes        string
	PhotoURL     string
	Status       string
	Error        string
}

type StudentOption struct {
	ID   string
	Name string
}

type PlanOption struct {
	ID   string
	Name string
}

type SubscriptionItem struct {
	ID          string
	StudentName string
	PlanName    string
	StartDate   string
	EndDate     string
	Status      string
	StatusLabel string
	StatusClass string
	Price       string
}

type SubscriptionsPageData struct {
	StudentID string
	Status    string
	Students  []StudentOption
	Items     []SubscriptionItem
}

type SubscriptionFormData struct {
	Title          string
	Action         string
	SubmitLabel    string
	DeleteAction   string
	ShowDelete     bool
	DisableSelects bool
	StudentID      string
	PlanID         string
	StartDate      string
	EndDate        string
	Status         string
	Price          string
	Students       []StudentOption
	Plans          []PlanOption
	Error          string
}

type SubscriptionOption struct {
	ID    string
	Label string
}

type PaymentItem struct {
	ID                string
	SubscriptionLabel string
	PaidAt            string
	Amount            string
	MethodLabel       string
	Reference         string
	Notes             string
	Status            string
	StatusLabel       string
	StatusClass       string
}

type PaymentsPageData struct {
	SubscriptionID string
	Status         string
	Subscriptions  []SubscriptionOption
	Items          []PaymentItem
}

type PaymentFormData struct {
	Title          string
	Action         string
	SubmitLabel    string
	DeleteAction   string
	ShowDelete     bool
	SubscriptionID string
	PaidAt         string
	Amount         string
	Method         string
	Reference      string
	Notes          string
	Status         string
	Subscriptions  []SubscriptionOption
	Error          string
}

type StudentsPreviewData struct {
	Items []StudentItem
}

type LoginData struct {
	Email string
	Error string
}
