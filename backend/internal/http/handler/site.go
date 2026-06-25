package handler

import (
	"net/http"

	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SiteHandler struct {
	site *service.SiteService
}

func NewSiteHandler(site *service.SiteService) *SiteHandler {
	return &SiteHandler{site: site}
}

func (h *SiteHandler) Public(c *gin.Context) {
	title, err := h.site.Title(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load site"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"title": title, "contact": h.site.Contact(c.Request.Context())})
}
