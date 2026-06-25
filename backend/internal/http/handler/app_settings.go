package handler

import (
	"net/http"

	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type AppSettingsHandler struct {
	settings *service.AppSettingsService
}

func NewAppSettingsHandler(settings *service.AppSettingsService) *AppSettingsHandler {
	return &AppSettingsHandler{settings: settings}
}

func (h *AppSettingsHandler) RegistrationGet(c *gin.Context) {
	data, err := h.settings.Registration(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load registration settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) RegistrationPut(c *gin.Context) {
	var body service.RegistrationSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveRegistration(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) SMTPGet(c *gin.Context) {
	data, err := h.settings.SMTP(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load smtp settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) SMTPPut(c *gin.Context) {
	var body service.SMTPSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveSMTP(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) SMTPTest(c *gin.Context) {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	if err := h.settings.TestSMTP(c.Request.Context(), body.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "detail": "测试邮件已发送"})
}

func (h *AppSettingsHandler) ProxyGet(c *gin.Context) {
	data, err := h.settings.Proxy(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load proxy settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) ProxyPut(c *gin.Context) {
	var body struct {
		Proxy string `json:"proxy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveProxy(c.Request.Context(), body.Proxy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) ProxyTest(c *gin.Context) {
	var body struct {
		Proxy string `json:"proxy"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.TestProxy(c.Request.Context(), body.Proxy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) CreditsGet(c *gin.Context) {
	data, err := h.settings.Credits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load credit settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) CreditsPut(c *gin.Context) {
	var body service.CreditSettings
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveCredits(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) LogsGet(c *gin.Context) {
	data, err := h.settings.Logs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load log settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) LogsPut(c *gin.Context) {
	var body struct {
		RetentionDays int `json:"retention_days"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveLogs(c.Request.Context(), body.RetentionDays)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data})
}

func (h *AppSettingsHandler) MediaGet(c *gin.Context) {
	data, err := h.settings.Media(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to load media settings"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AppSettingsHandler) MediaPut(c *gin.Context) {
	var body struct {
		RetentionDays int `json:"retention_days"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid request body"})
		return
	}
	data, err := h.settings.SaveMedia(c.Request.Context(), body.RetentionDays)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "data": data.Settings, "removed": data.Removed, "freed_bytes": data.FreedBytes})
}
