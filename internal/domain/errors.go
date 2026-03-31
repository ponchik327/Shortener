package domain

import "errors"

// Sentinel-ошибки, используемые в слоях сервиса и хендлеров.
var (
	ErrNotFound         = errors.New("link not found")
	ErrCodeTaken        = errors.New("short code already taken")
	ErrGenerationFailed = errors.New("failed to generate unique short code")
	ErrCacheMiss        = errors.New("cache miss")
)
