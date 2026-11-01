package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const readyTimeout = 2 * time.Second

// Pinger reports whether a dependency can be reached.
type Pinger interface {
	Ping(ctx context.Context) error
}

func handleReadyz(logger *slog.Logger, db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			logger.Error("readiness check failed", "err", err)
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
	}
}
