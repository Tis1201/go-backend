package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status string `json:"status"`
}

func handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}
