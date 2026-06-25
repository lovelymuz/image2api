package handler

import (
	"net/http"

	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ShowcaseHandler struct {
	showcase *service.ShowcaseService
}

func NewShowcaseHandler(showcase *service.ShowcaseService) *ShowcaseHandler {
	return &ShowcaseHandler{showcase: showcase}
}

func (h *ShowcaseHandler) List(c *gin.Context) {
	grouped, err := h.showcase.Grouped(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load showcase"})
		return
	}
	out := gin.H{}
	for kind, items := range grouped {
		rows := make([]gin.H, 0, len(items))
		for _, item := range items {
			rows = append(rows, gin.H{
				"id":         item.ID,
				"kind":       item.Kind,
				"title":      item.Title,
				"subtitle":   item.Subtitle,
				"prompt":     item.Prompt,
				"gradient":   item.Gradient,
				"span":       item.Span,
				"image":      item.Image,
				"weight":     item.Weight,
				"created_at": item.CreatedAt,
				"updated_at": item.UpdatedAt,
			})
		}
		out[kind] = rows
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
