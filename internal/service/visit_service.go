package service

import (
	"context"
	"fmt"
	"time"

	"github.com/wb-go/wbf/logger"

	"github.com/ponchik327/Shortener/internal/domain"
)

const (
	_asyncVisitTimeout = 5 * time.Second
)

// visitRepository — интерфейс репозитория, используемый VisitService.
type visitRepository interface {
	InsertVisit(ctx context.Context, visit *domain.Visit) error
	GetAnalytics(ctx context.Context, linkID int64) (*domain.Analytics, error)
}

// linkService — интерфейс разрешения ссылок, используемый VisitService.
type linkService interface {
	GetLink(ctx context.Context, code string) (*domain.Link, error)
}

// VisitService отвечает за запись переходов и получение аналитики.
type VisitService struct {
	visitRepo visitRepository
	linkSvc   linkService
	log       logger.Logger
}

// NewVisitService создаёт новый VisitService.
func NewVisitService(
	visitRepo visitRepository,
	linkSvc linkService,
	log logger.Logger,
) *VisitService {
	return &VisitService{
		visitRepo: visitRepo,
		linkSvc:   linkSvc,
		log:       log,
	}
}

// RecordVisitAsync записывает переход в фоновой горутине, чтобы не задерживать редирект.
func (s *VisitService) RecordVisitAsync(linkID int64, userAgent string) {
	go func() {
		visitCtx, cancel := context.WithTimeout(context.Background(), _asyncVisitTimeout)
		defer cancel()

		visit := &domain.Visit{
			LinkID:    linkID,
			UserAgent: userAgent,
			VisitedAt: time.Now(),
		}

		if err := s.visitRepo.InsertVisit(visitCtx, visit); err != nil {
			s.log.Error("record visit async failed", "link_id", linkID, "error", err)
		}
	}()
}

// GetAnalytics возвращает агрегированную статистику переходов для указанного короткого кода.
func (s *VisitService) GetAnalytics(ctx context.Context, code string) (*domain.Analytics, error) {
	link, err := s.linkSvc.GetLink(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("resolve link for analytics: %w", err)
	}

	analytics, err := s.visitRepo.GetAnalytics(ctx, link.ID)
	if err != nil {
		return nil, fmt.Errorf("get analytics for link %d: %w", link.ID, err)
	}

	return analytics, nil
}
