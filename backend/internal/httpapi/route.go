package httpapi

import (
	"log/slog"
	"net/http"
)

// NewHandler returns the root HTTP handler with every route registered.
func NewHandler(logger *slog.Logger, db Pinger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)
	mux.HandleFunc("GET /readyz", handleReadyz(logger, db))
	return mux
}
