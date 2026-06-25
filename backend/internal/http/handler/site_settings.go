package handler

import (
	"net/http"
	"strings"

	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SiteSettingsHandler struct {
	site *service.SiteService
}

func NewSiteSettingsHandler(site *service.SiteService) *SiteSettingsHandler {
	return &SiteSettingsHandler{site: site}
}

func (h *SiteSettingsHandler) Get(c *gin.Context) {
	title, err := h.site.Title(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load site settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"title": title, "contact": h.site.Contact(c.Request.Context())})
}

func (h *SiteSettingsHandler) Put(c *gin.Context) {
	var body struct {
		Title   string          `json:"title"`
		Contact service.Contact `json:"contact"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "网页主标题不能为空"})
		return
	}
	updated, err := h.site.SetTitle(c.Request.Context(), title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to save site settings"})
		return
	}
	if err := h.site.SetContact(c.Request.Context(), body.Contact); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to save contact info"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": gin.H{"title": updated, "contact": h.site.Contact(c.Request.Context())}})
}
