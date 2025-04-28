package health

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
)

// Health is a health checker.
type Health struct {
	isReady atomic.Value
	addr    string
	srv     *http.Server
}

// New returns a new health checker.
func New(addr string) (*Health, error) {
	var h Health

	// Set the address
	h.addr = addr

	// Readiness to false
	h.isReady.Store(false)

	return &h, nil
}

// Ready sets the readiness of the health checker.
func (h *Health) Ready(isReady bool) {
	h.isReady.Store(isReady)
}

// Serve starts the health checker synchronously.
func (h *Health) Serve() error {
	// Check if the server is started
	if h.srv != nil {
		return errors.New("server already started")
	}

	// Register the handlers
	http.HandleFunc("/liveness", h.liveness())
	http.HandleFunc("/readiness", h.readiness())

	// Start the server
	// NOTE: This is a blocking call
	h.srv = &http.Server{Addr: h.addr}
	err := h.srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Shutdown stops the health checker.
func (h *Health) Shutdown(ctx context.Context) error {
	// Check if the server is started
	if h.srv == nil {
		return errors.New("server not started")
	}

	// Set the readiness to false just in case
	h.Ready(false)

	// Shutdown the server
	err := h.srv.Shutdown(ctx)
	if err != nil {
		return err
	}

	// Set the server to nil
	h.srv = nil

	return nil
}

func (h *Health) liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Health) readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if !h.isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
