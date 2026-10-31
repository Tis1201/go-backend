package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantJSON   bool
	}{
		{
			name:       "GET returns 200",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantJSON:   true,
		},
		{
			name:       "POST returns 405",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantJSON:   false,
		},
	}

	handler := NewHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			request := httptest.NewRequest(
				tt.method,
				"/healthz",
				nil,
			)
			response := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(response, request)

			// Assert: Status Code
			if response.Code != tt.wantStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.wantStatus,
					response.Code,
				)
			}

			// Chỉ kiểm tra JSON khi cần
			if !tt.wantJSON {
				return
			}

			// Assert: Content-Type
			contentType := response.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf(
					"expected application/json, got %s",
					contentType,
				)
			}

			// Assert: JSON Body
			var body struct {
				Status string `json:"status"`
			}

			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatalf("decode JSON failed: %v", err)
			}

			if body.Status != "ok" {
				t.Errorf("expected status ok, got %s", body.Status)
			}
		})
	}
}
