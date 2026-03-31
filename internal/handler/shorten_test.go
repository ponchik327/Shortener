package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
)

func makeShortenReq(body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req, httptest.NewRecorder()
}

func TestShorten_InvalidJSON(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	req, w := makeShortenReq("not-json")
	h.Shorten(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_EmptyLongURL(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":""}`)
	h.Shorten(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_InvalidURL(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":"notaurl"}`)
	h.Shorten(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_CustomCode_TooShort(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":"http://example.com","custom_code":"ab"}`)
	h.Shorten(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_CustomCode_TooLong(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	code := strings.Repeat("a", 31)
	req, w := makeShortenReq(`{"long_url":"http://example.com","custom_code":"` + code + `"}`)
	h.Shorten(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_CustomCode_InvalidChars(t *testing.T) {
	h := newTestHandler(&mockLinkCreator{}, nil, nil, nil)
	for _, code := range []string{"has space", "has_under"} {
		req, w := makeShortenReq(`{"long_url":"http://example.com","custom_code":"` + code + `"}`)
		h.Shorten(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code, "code: %q", code)
	}
}

func TestShorten_CodeTaken(t *testing.T) {
	lc := &mockLinkCreator{}
	lc.On("CreateLink", mock.Anything, "http://example.com", "mycode").
		Return(nil, domain.ErrCodeTaken).Once()

	h := newTestHandler(lc, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":"http://example.com","custom_code":"mycode"}`)
	h.Shorten(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	lc.AssertExpectations(t)
}

func TestShorten_ServiceError(t *testing.T) {
	lc := &mockLinkCreator{}
	lc.On("CreateLink", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("db")).Once()

	h := newTestHandler(lc, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":"http://example.com"}`)
	h.Shorten(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	lc.AssertExpectations(t)
}

func TestShorten_Success(t *testing.T) {
	lc := &mockLinkCreator{}
	link := &domain.Link{ID: 1, ShortCode: "abc123", LongURL: "http://example.com"}
	lc.On("CreateLink", mock.Anything, "http://example.com", "").Return(link, nil).Once()

	h := newTestHandler(lc, nil, nil, nil)
	req, w := makeShortenReq(`{"long_url":"http://example.com"}`)
	h.Shorten(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp shortenResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "abc123", resp.ShortCode)
	assert.Contains(t, resp.ShortURL, "abc123")

	lc.AssertExpectations(t)
}
