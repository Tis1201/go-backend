package httpapi

import (
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain puts Gin in test mode so it does not print debug output during tests.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}
