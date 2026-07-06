package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListarAlertas(c *gin.Context) {
	alertas, err := h.svc.ListarAlertas(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alertas)
}

func (h *Handler) ObtenerAlerta(c *gin.Context) {
	id := c.Param("id")
	alerta, err := h.svc.ObtenerAlerta(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if alerta == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alerta no encontrada"})
		return
	}
	c.JSON(http.StatusOK, alerta)
}
