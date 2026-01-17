package postgres

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/adapter/postgres/sqlc"
	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository struct {
	queries *sqlc.Queries
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{queries: sqlc.New(pool)}
}

func (r *StudentRepository) Create(ctx context.Context, student domain.Student) (domain.Student, error) {
	params := sqlc.CreateStudentParams{
		FullName:  student.FullName,
		BirthDate: dateTo(student.BirthDate),
		Gender:    textTo(student.Gender),
		Phone:     textTo(student.Phone),
		Email:     textTo(student.Email),
		Cpf:       textTo(student.CPF),
		Address:   textTo(student.Address),
		Notes:     textTo(student.Notes),
		PhotoUrl:  textTo(student.PhotoURL),
		Status:    sqlc.StudentStatus(student.Status),
	}

	created, err := r.queries.CreateStudent(ctx, params)
	if err != nil {
		return domain.Student{}, err
	}

	return mapStudent(created), nil
}

func (r *StudentRepository) Update(ctx context.Context, student domain.Student) (domain.Student, error) {
	id, err := stringToUUID(student.ID)
	if err != nil || !id.Valid {
		return domain.Student{}, err
	}

	params := sqlc.UpdateStudentParams{
		ID:        id,
		FullName:  student.FullName,
		BirthDate: dateTo(student.BirthDate),
		Gender:    textTo(student.Gender),
		Phone:     textTo(student.Phone),
		Email:     textTo(student.Email),
		Cpf:       textTo(student.CPF),
		Address:   textTo(student.Address),
		Notes:     textTo(student.Notes),
		PhotoUrl:  textTo(student.PhotoURL),
		Status:    sqlc.StudentStatus(student.Status),
	}

	updated, err := r.queries.UpdateStudent(ctx, params)
	if err != nil {
		return domain.Student{}, err
	}

	return mapStudent(updated), nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id string) (domain.Student, error) {
	uuidValue, err := stringToUUID(id)
	if err != nil || !uuidValue.Valid {
		return domain.Student{}, err
	}

	student, err := r.queries.GetStudent(ctx, uuidValue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Student{}, ports.ErrNotFound
		}
		return domain.Student{}, err
	}

	return mapStudent(student), nil
}

func (r *StudentRepository) Search(ctx context.Context, filter ports.StudentFilter) ([]domain.Student, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	statuses := make([]sqlc.StudentStatus, 0, len(filter.Statuses))
	for _, status := range filter.Statuses {
		if status != "" {
			statuses = append(statuses, sqlc.StudentStatus(status))
		}
	}

	params := sqlc.SearchStudentsParams{
		Column1: filter.Query,
		Column2: statuses,
		Limit:   int32(limit),
		Offset:  int32(filter.Offset),
	}

	rows, err := r.queries.SearchStudents(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Student, 0, len(rows))
	for _, student := range rows {
		result = append(result, mapStudent(student))
	}

	return result, nil
}

func mapStudent(student sqlc.Student) domain.Student {
	return domain.Student{
		ID:        uuidToString(student.ID),
		FullName:  student.FullName,
		BirthDate: dateFrom(student.BirthDate),
		Gender:    textFrom(student.Gender),
		Phone:     textFrom(student.Phone),
		Email:     textFrom(student.Email),
		CPF:       textFrom(student.Cpf),
		Address:   textFrom(student.Address),
		Notes:     textFrom(student.Notes),
		PhotoURL:  textFrom(student.PhotoUrl),
		Status:    domain.StudentStatus(student.Status),
		CreatedAt: timeFrom(student.CreatedAt),
		UpdatedAt: timeFrom(student.UpdatedAt),
	}
}
