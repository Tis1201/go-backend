package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	handler := NewHandler()
	request, _ := http.NewRequest("GET", "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 OK got %d", response.Code)
	}

	contentType := response.Header()["Content-Type"][0]

	if contentType != "application/json" {
		t.Errorf("expected application/json got %s", contentType)
	}

	var body struct {
		Status string `json:"status"`
	}

	err := json.NewDecoder(response.Body).Decode(&body)

	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf(
			"expected status ok, got %s",
			body.Status,
		)
	}
}
