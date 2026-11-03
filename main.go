package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bszendro/videoservice/internal/metrics"
)

func newServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal(struct {
			World string `json:"world"`
		}{World: "Hello"})
		if err != nil {
			slog.Error("encoding", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(body); err != nil {
			slog.Error("writing response", "error", err) // client likely went away; too late for a status code
		}
	})

	return &http.Server{Addr: ":8000", Handler: mux, ReadHeaderTimeout: 10 * time.Second}
}

func listenAndServe(ctx context.Context, servers ...*http.Server) error {
	errCh := make(chan error, len(servers))
	var serveWG sync.WaitGroup
	for _, s := range servers {
		serveWG.Go(func() {
			if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				slog.Error("serve", "error", err)
				errCh <- err
			}
		})
	}

	var result error
	select {
	case result = <-errCh: // Only the first error is returned if any of the servers fail
	case <-ctx.Done():
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Go(func() {
			if err := s.Shutdown(shutCtx); err != nil {
				slog.Error("shutdown", "error", err)
				s.Close()
			}
		})
	}
	wg.Wait()
	serveWG.Wait()
	return result
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := newServer()
	m := metrics.New()
	metricsServer := metrics.NewMetricsServer(m)

	var wg sync.WaitGroup
	wg.Go(func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}

			m.VidUploaded.Inc()
		}
	})

	err := listenAndServe(ctx, server, metricsServer)

	cancel()
	wg.Wait()

	if err != nil {
		os.Exit(1)
	}
}
