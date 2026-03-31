package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
)

func TestRedirect_NotFound(t *testing.T) {
	lg := &mockLinkGetter{}
	lg.On("GetLink", mock.Anything, "abc").Return(nil, domain.ErrNotFound).Once()

	h := newTestHandler(nil, lg, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/s/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	lg.AssertExpectations(t)
}

func TestRedirect_ServiceError(t *testing.T) {
	lg := &mockLinkGetter{}
	lg.On("GetLink", mock.Anything, "abc").Return(nil, errors.New("db")).Once()

	h := newTestHandler(nil, lg, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/s/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	lg.AssertExpectations(t)
}

func TestRedirect_Success(t *testing.T) {
	lg := &mockLinkGetter{}
	vr := &mockVisitRecorder{}

	link := &domain.Link{ID: 1, ShortCode: "abc", LongURL: "http://dest.com"}
	lg.On("GetLink", mock.Anything, "abc").Return(link, nil).Once()
	vr.On("RecordVisitAsync", int64(1), mock.Anything).Return().Once()

	h := newTestHandler(nil, lg, vr, nil)
	req := httptest.NewRequest(http.MethodGet, "/s/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Redirect(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "http://dest.com", w.Header().Get("Location"))

	lg.AssertExpectations(t)
	vr.AssertExpectations(t)
}
