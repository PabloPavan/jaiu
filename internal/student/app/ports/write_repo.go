package ports

import (
	"context"

	"github.com/PabloPavan/jaiu/internal/student/domain"
)

type StudentWriteRepository interface {
	Create(ctx context.Context, student domain.Student) (domain.Student, error)
	Update(ctx context.Context, student domain.Student) (domain.Student, error)
	FindByID(ctx context.Context, id string) (domain.Student, error)
}
