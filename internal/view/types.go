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
	Name         string
	DisplayName  string
	Role         string
}

type PlanItem struct {
	Name         string
	DurationDays int
	Price        string
	Description  string
}

type PlansPageData struct {
	Items []PlanItem
}

type StudentsPreviewData struct {
	Items []string
}

type LoginData struct {
	Email string
	Error string
}
