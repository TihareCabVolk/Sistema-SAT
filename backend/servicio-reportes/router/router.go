package router

import (
	"net/http"
	"servicio-reportes/handler"
)


func SetupRouter() {
	http.HandleFunc("POST /reportes", handler.HandleReport)
}