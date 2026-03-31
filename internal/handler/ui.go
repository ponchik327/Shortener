package handler

import (
	"net/http"
)

// UI обрабатывает GET / и отдаёт одностраничный HTML-интерфейс.
func (h *Handler) UI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.tmpl.Execute(w, nil); err != nil {
		h.log.Error("render ui template", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
