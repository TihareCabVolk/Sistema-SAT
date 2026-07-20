package router

import (
	"net/http"
	"servicio-reportes/handler"
)

func SetupRouter() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	http.HandleFunc("/api/reportes", handler.HandleReport)
}