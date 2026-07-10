package router

import (
	"servicio-logistica/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handler.Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/logistica")
	{
		api.POST("/validacion", h.ProcesarValidacion)
		api.GET("/alertas", h.ListarAlertas)
		api.GET("/alertas/:id", h.ObtenerAlerta)
	}

	return r
}
