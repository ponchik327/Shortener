package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/ponchik327/Shortener/internal/domain"
)

type linkGetter interface {
	GetLink(ctx context.Context, code string) (*domain.Link, error)
}

type visitRecorder interface {
	RecordVisitAsync(linkID int64, userAgent string)
}

// Redirect обрабатывает GET /s/{short_url}.
// Разрешает короткий код в оригинальный URL, асинхронно записывает переход и выполняет редирект.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("short_url")

	link, err := h.linkGetter.GetLink(r.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "short link not found")
			return
		}

		h.log.Error("get link for redirect", "code", code, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	h.visitRecorder.RecordVisitAsync(link.ID, r.UserAgent())
	http.Redirect(w, r, link.LongURL, http.StatusFound)
}
