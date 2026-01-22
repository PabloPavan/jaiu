package ports

import (
	"context"

	domain "github.com/PabloPavan/jaiu/internal/student/domain"
)

type StudentReadRepository interface {
	Search(ctx context.Context, filter StudentFilter) ([]domain.Student, error)
	Count(ctx context.Context, filter StudentFilter) (int, error)
}

type StudentFilter struct {
	Query    string
	Statuses []domain.StudentStatus
	Limit    int
	Offset   int
}
