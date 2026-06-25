package middleware

import (
	"net/http"

	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

const currentUserKey = "current_user"
const currentSessionKey = "current_session"

func RequireSession(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, session, err := auth.CurrentUserFromRequest(
			c.Request.Context(),
			c.GetHeader("Authorization"),
			readCookie(c, "vivid_session"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to validate session"})
			c.Abort()
			return
		}
		if user == nil || session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "未登录或会话已过期"})
			c.Abort()
			return
		}
		c.Set(currentUserKey, user)
		c.Set(currentSessionKey, session)
		c.Next()
	}
}

func RequireAdminSession(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, session, err := auth.CurrentUserFromRequest(
			c.Request.Context(),
			c.GetHeader("Authorization"),
			readCookie(c, "vivid_session"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"detail": "failed to validate session"})
			c.Abort()
			return
		}
		if user == nil || session == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "未登录或会话已过期"})
			c.Abort()
			return
		}
		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"detail": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Set(currentUserKey, user)
		c.Set(currentSessionKey, session)
		c.Next()
	}
}

func readCookie(c *gin.Context, name string) string {
	v, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return v
}
