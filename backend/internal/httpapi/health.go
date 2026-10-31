package httpapi

import "net/http"

type healthResponse struct {
	Status string `json:"status"`
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJson(w, http.StatusOK, healthResponse{Status: "ok"})
}
