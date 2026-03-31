package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ponchik327/Shortener/internal/domain"
)

func TestAnalytics_NotFound(t *testing.T) {
	ap := &mockAnalyticsProvider{}
	ap.On("GetAnalytics", mock.Anything, "abc").Return(nil, domain.ErrNotFound).Once()

	h := newTestHandler(nil, nil, nil, ap)
	req := httptest.NewRequest(http.MethodGet, "/analytics/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Analytics(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ap.AssertExpectations(t)
}

func TestAnalytics_ServiceError(t *testing.T) {
	ap := &mockAnalyticsProvider{}
	ap.On("GetAnalytics", mock.Anything, "abc").Return(nil, errors.New("db")).Once()

	h := newTestHandler(nil, nil, nil, ap)
	req := httptest.NewRequest(http.MethodGet, "/analytics/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Analytics(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	ap.AssertExpectations(t)
}

func TestAnalytics_Success(t *testing.T) {
	ap := &mockAnalyticsProvider{}
	analytics := &domain.Analytics{TotalVisits: 7}
	ap.On("GetAnalytics", mock.Anything, "abc").Return(analytics, nil).Once()

	h := newTestHandler(nil, nil, nil, ap)
	req := httptest.NewRequest(http.MethodGet, "/analytics/abc", nil)
	req.SetPathValue("short_url", "abc")
	w := httptest.NewRecorder()

	h.Analytics(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var result domain.Analytics
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, int64(7), result.TotalVisits)

	ap.AssertExpectations(t)
}
