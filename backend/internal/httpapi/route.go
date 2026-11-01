// Package httpapi contains the HTTP routes and handlers of the API, built on Gin.
package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewHandler returns the root HTTP handler with every route registered.
//
// It returns http.Handler instead of *gin.Engine so main only depends on the
// standard interface and keeps serving through its own http.Server.
func NewHandler(logger *slog.Logger, db Pinger) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	// Gin answers 404 when a known path is called with the wrong method;
	// net/http answers 405. Keep the correct status.
	r.HandleMethodNotAllowed = true

	// Gin trusts every proxy by default, so any client could fake its IP through
	// X-Forwarded-For. Trust none until a real proxy sits in front of the API.
	// A nil list never returns an error.
	_ = r.SetTrustedProxies(nil)

	r.GET("/healthz", handleHealthz)
	r.GET("/readyz", handleReadyz(logger, db))
	return r
}
