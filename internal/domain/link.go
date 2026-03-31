package domain

import "time"

// Link представляет связку короткого кода с оригинальным URL.
type Link struct {
	ID        int64
	ShortCode string
	LongURL   string
	CreatedAt time.Time
}
