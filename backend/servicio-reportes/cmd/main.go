package main

import (
	"fmt"
	"net/http"

	"servicio-reportes/rabbitmq"
	"servicio-reportes/repository"
	"servicio-reportes/router"
)

func main() {
	// 1. Inicializar BD
	repository.InitDB()
	rabbitmq.InitRabbit()
	defer repository.Client.Close()

	// 2. Cargar rutas
	go rabbitmq.StartConsumer()
	router.SetupRouter()

	// 3. Levantar servidor
	fmt.Println("Servicio 1 (Centro de Reportes) escuchando ...")
	if err := http.ListenAndServe(":4001", nil); err != nil {
		fmt.Println("Error al levantar el servidor:", err)
	}
}
