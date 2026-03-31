package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/wb-go/wbf/logger"

	"github.com/ponchik327/Shortener/internal/domain"
)

const (
	_maxCodeGenAttempts = 5
)

// linkRepository — интерфейс репозитория, используемый LinkService.
type linkRepository interface {
	InsertLink(ctx context.Context, link *domain.Link) (*domain.Link, error)
	GetByShortCode(ctx context.Context, code string) (*domain.Link, error)
	ShortCodeExists(ctx context.Context, code string) (bool, error)
}

// linkCache — интерфейс кэша, используемый LinkService.
type linkCache interface {
	GetByCode(ctx context.Context, code string) (linkID int64, longURL string, err error)
	Set(ctx context.Context, code string, linkID int64, longURL string) error
}

// LinkService реализует бизнес-логику сокращения URL.
type LinkService struct {
	repo       linkRepository
	cache      linkCache
	log        logger.Logger
	codeLength int
}

// NewLinkService создаёт новый LinkService.
func NewLinkService(
	repo linkRepository,
	cache linkCache,
	log logger.Logger,
	codeLength int,
) *LinkService {
	return &LinkService{
		repo:       repo,
		cache:      cache,
		log:        log,
		codeLength: codeLength,
	}
}

// CreateLink создаёт новую короткую ссылку. Если customCode не пустой — используется он;
// в противном случае генерируется случайный base62-код.
func (s *LinkService) CreateLink(ctx context.Context, longURL, customCode string) (*domain.Link, error) {
	code, err := s.resolveCode(ctx, customCode)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.InsertLink(ctx, &domain.Link{ShortCode: code, LongURL: longURL})
	if err != nil {
		return nil, fmt.Errorf("insert link: %w", err)
	}

	s.populateCache(ctx, created)

	return created, nil
}

// GetLink возвращает ссылку по короткому коду, используя кэш при наличии данных.
func (s *LinkService) GetLink(ctx context.Context, code string) (*domain.Link, error) {
	linkID, longURL, err := s.cache.GetByCode(ctx, code)
	switch {
	case err == nil:
		return &domain.Link{ID: linkID, ShortCode: code, LongURL: longURL}, nil
	case errors.Is(err, domain.ErrCacheMiss):
	default:
		s.log.Warn("cache get failed, falling back to db", "code", code, "error", err)
	}

	link, err := s.repo.GetByShortCode(ctx, code)
	if err != nil {
		return nil, err
	}

	s.populateCache(ctx, link)

	return link, nil
}

// resolveCode проверяет доступность пользовательского кода или генерирует уникальный случайный.
func (s *LinkService) resolveCode(ctx context.Context, customCode string) (string, error) {
	if customCode != "" {
		return s.validateCustomCode(ctx, customCode)
	}

	return s.generateUniqueCode(ctx)
}

func (s *LinkService) validateCustomCode(ctx context.Context, code string) (string, error) {
	exists, err := s.repo.ShortCodeExists(ctx, code)
	if err != nil {
		return "", fmt.Errorf("check custom code availability: %w", err)
	}

	if exists {
		return "", domain.ErrCodeTaken
	}

	return code, nil
}

func (s *LinkService) generateUniqueCode(ctx context.Context) (string, error) {
	for range _maxCodeGenAttempts {
		code, err := generateCode(s.codeLength)
		if err != nil {
			return "", fmt.Errorf("generate code: %w", err)
		}

		exists, err := s.repo.ShortCodeExists(ctx, code)
		if err != nil {
			return "", fmt.Errorf("check generated code availability: %w", err)
		}

		if !exists {
			return code, nil
		}
	}

	return "", domain.ErrGenerationFailed
}

// populateCache записывает ссылку в кэш. Ошибки логируются и не прерывают вызывающий код.
func (s *LinkService) populateCache(ctx context.Context, link *domain.Link) {
	if err := s.cache.Set(ctx, link.ShortCode, link.ID, link.LongURL); err != nil {
		s.log.Warn("cache set failed", "code", link.ShortCode, "error", err)
	}
}
