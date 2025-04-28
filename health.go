package health

import (
	"context"
	"net/http"
	"sync/atomic"
)

// Health is a health checker.
type Health struct {
	isReady atomic.Value
}

// NewHealth returns a new health checker.
func NewHealth(ctx context.Context) (*Health, error) {
	var h Health

	// Readiness to false
	h.isReady.Store(false)

	return &h, nil
}

// Ready sets the readiness of the health checker.
func (h *Health) Ready(isReady bool) {
	h.isReady.Store(isReady)
}

// HTTPServe starts the health checker.
func (h *Health) HTTPServe(ctx context.Context) {
	http.HandleFunc("/liveness", h.liveness())
	http.HandleFunc("/readiness", h.readiness())
}

func (h *Health) liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Health) readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
