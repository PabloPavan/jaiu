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

type StudentsPreviewData struct {
	Items []string
}

type LoginData struct {
	Email string
	Error string
}
