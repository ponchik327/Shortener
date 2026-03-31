package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/ponchik327/Shortener/internal/domain"
)

type linkCreator interface {
	CreateLink(ctx context.Context, longURL, customCode string) (*domain.Link, error)
}

const (
	_minCustomCodeLen = 3
	_maxCustomCodeLen = 30
)

var (
	_customCodeRegexp        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	errCodeInvalidLength     = fmt.Errorf("short code must be %d-%d characters", _minCustomCodeLen, _maxCustomCodeLen)
	errCodeInvalidCharacters = errors.New("short code must contain only letters, digits and hyphens")
)

func validateCustomCode(code string) error {
	if len(code) < _minCustomCodeLen || len(code) > _maxCustomCodeLen {
		return errCodeInvalidLength
	}

	if !_customCodeRegexp.MatchString(code) {
		return errCodeInvalidCharacters
	}

	return nil
}

type shortenRequest struct {
	LongURL    string `json:"long_url"`
	CustomCode string `json:"custom_code"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

// Shorten обрабатывает POST /shorten.
// Создаёт новую короткую ссылку и возвращает короткий код и полный короткий URL.
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.LongURL == "" {
		h.writeError(w, http.StatusBadRequest, "long_url is required")
		return
	}

	if _, err := url.ParseRequestURI(req.LongURL); err != nil {
		h.writeError(w, http.StatusBadRequest, "long_url is not a valid URL")
		return
	}

	if req.CustomCode != "" {
		if err := validateCustomCode(req.CustomCode); err != nil {
			h.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	link, err := h.linkCreator.CreateLink(r.Context(), req.LongURL, req.CustomCode)
	if err != nil {
		if errors.Is(err, domain.ErrCodeTaken) {
			h.writeError(w, http.StatusConflict, "short code is already taken")
			return
		}

		h.log.Error("create link", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")

		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	resp := shortenResponse{
		ShortCode: link.ShortCode,
		ShortURL:  fmt.Sprintf("%s://%s/s/%s", scheme, r.Host, link.ShortCode),
	}

	h.writeJSON(w, http.StatusCreated, resp)
}
