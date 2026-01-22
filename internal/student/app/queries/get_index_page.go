package queries

import (
	"time"

	"github.com/PabloPavan/jaiu/internal/student/app/ports"
)

type GetStudentsGrid struct {
	FillName   string
	Status     string
	PlanName   string
	End_date   time.Time
	Created_at time.Time
	Paid_at    time.Time
}

type GetIndexPageStudent struct {
	Repo ports.StudentReadRepository
	// Audit p.AuditRepository
}

func (g *GetIndexPageStudent) Execute() (GetStudentsGrid, error) {
	return GetStudentsGrid{}, nil
}
