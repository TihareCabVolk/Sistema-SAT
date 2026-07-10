package handler

import (
	"net/http"

	"servicio-logistica/models"
	"servicio-logistica/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.LogisticaService
}

func NewHandler(svc *service.LogisticaService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ProcesarValidacion(c *gin.Context) {
	var vp models.ValidacionPositiva
	if err := c.ShouldBindJSON(&vp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alerta, err := h.svc.ProcesarValidacion(c.Request.Context(), &vp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerta)
}
