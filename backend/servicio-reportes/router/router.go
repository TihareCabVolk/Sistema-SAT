package router

import (
	"net/http"
	"servicio-reportes/handler"
)

func SetupRouter() {
	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	http.HandleFunc("POST /api/reportes", handler.HandleReport)
}