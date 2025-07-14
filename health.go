package health

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"
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

// Serve starts the health checker synchronously and listens for context cancellation.
func (h *Health) Serve(ctx context.Context) error {
	// Check if the server is already started
	if h.srv != nil {
		return errors.New("server already started")
	}

	// Register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/liveness", h.liveness())
	mux.HandleFunc("/readiness", h.readiness())

	// Create server
	h.srv = &http.Server{Addr: h.addr, Handler: mux}
	defer func() { h.srv = nil }()

	// Listen for connections
	ln, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- h.srv.Serve(ln)
	}()
	defer close(serverErr)

	select {
	case <-ctx.Done(): // Context cancelled
		// Set readiness to false
		h.Ready(false)

		// Shutdown the server using a context derived from the parent
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := h.srv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		return ctx.Err()
	case err := <-serverErr: // Server error
		return err
	}
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
