package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readyTimeout = 2 * time.Second

// Pinger reports whether a dependency can be reached.
type Pinger interface {
	Ping(ctx context.Context) error
}

func handleReadyz(logger *slog.Logger, db Pinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the request's context, not c itself: with Gin's default settings
		// c.Done() is nil, so c never reports that the client went away.
		ctx, cancel := context.WithTimeout(c.Request.Context(), readyTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			logger.Error("readiness check failed", "err", err)
			c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
			return
		}
		c.JSON(http.StatusOK, healthResponse{Status: "ok"})
	}
}
