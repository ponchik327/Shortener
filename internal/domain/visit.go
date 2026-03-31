package domain

import "time"

// Visit представляет единичное событие перехода по короткой ссылке.
type Visit struct {
	ID        int64
	LinkID    int64
	UserAgent string
	VisitedAt time.Time
}

// Analytics содержит агрегированную статистику переходов по короткой ссылке.
type Analytics struct {
	TotalVisits int64
	ByDay       []DayCount
	ByMonth     []MonthCount
	ByUserAgent []UserAgentCount
}

// DayCount содержит количество переходов, агрегированное по календарному дню.
type DayCount struct {
	Day   string // формат: "2006-01-02"
	Count int64
}

// MonthCount содержит количество переходов, агрегированное по календарному месяцу.
type MonthCount struct {
	Month string // формат: "2006-01"
	Count int64
}

// UserAgentCount содержит количество переходов, агрегированное по строке User-Agent.
type UserAgentCount struct {
	UserAgent string
	Count     int64
}
