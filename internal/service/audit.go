package service

import (
	"context"
	"errors"

	"github.com/PabloPavan/jaiu/internal/domain"
	"github.com/PabloPavan/jaiu/internal/ports"
)

type AuditService struct {
	repo ports.AuditRepository
}

func NewAuditService(repo ports.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Record(ctx context.Context, event domain.AuditEvent) error {
	if s == nil || s.repo == nil {
		return errors.New("audit repository unavailable")
	}
	return s.repo.Record(ctx, event)
}
