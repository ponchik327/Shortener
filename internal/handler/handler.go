package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/wb-go/wbf/logger"
)

// Handler содержит зависимости HTTP-хендлеров.
type Handler struct {
	linkCreator       linkCreator
	linkGetter        linkGetter
	visitRecorder     visitRecorder
	analyticsProvider analyticsProvider
	tmpl              *template.Template
	log               logger.Logger
}

// New создаёт новый Handler и парсит шаблон UI.
// linkSvc должен реализовывать linkCreator и linkGetter;
// visitSvc должен реализовывать visitRecorder и analyticsProvider.
func New(
	linkSvc interface {
		linkCreator
		linkGetter
	},
	visitSvc interface {
		visitRecorder
		analyticsProvider
	},
	tmplPath string,
	log logger.Logger,
) (*Handler, error) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}

	return &Handler{
		linkCreator:       linkSvc,
		linkGetter:        linkSvc,
		visitRecorder:     visitSvc,
		analyticsProvider: visitSvc,
		tmpl:              tmpl,
		log:               log,
	}, nil
}

// Register регистрирует все маршруты на переданном мультиплексоре.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /shorten", h.Shorten)
	mux.HandleFunc("GET /s/{short_url}", h.Redirect)
	mux.HandleFunc("GET /analytics/{short_url}", h.Analytics)
	mux.HandleFunc("GET /", h.UI)
}

// responseWriter оборачивает http.ResponseWriter для перехвата статус-кода при логировании.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware логирует каждый запрос: метод, путь, статус и длительность.
func LoggingMiddleware(next http.Handler, log logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.LogRequest(r.Context(), r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}

// writeJSON записывает v как JSON с указанным статус-кодом.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("write json response", "error", err)
	}
}

// writeError записывает JSON-ответ с ошибкой.
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
