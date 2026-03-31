package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/ponchik327/Shortener/internal/domain"
)

type analyticsProvider interface {
	GetAnalytics(ctx context.Context, code string) (*domain.Analytics, error)
}

// Analytics обрабатывает GET /analytics/{short_url}.
// Возвращает JSON с общим числом переходов и разбивкой по дням, месяцам и User-Agent.
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("short_url")

	analytics, err := h.analyticsProvider.GetAnalytics(r.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "short link not found")
			return
		}

		h.log.Error("get analytics", "code", code, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	h.writeJSON(w, http.StatusOK, analytics)
}
