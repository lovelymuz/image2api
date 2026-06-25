package bootstrap

import (
	"context"

	"backend/internal/model"
	"gorm.io/gorm"
)

func seedDefaults(ctx context.Context, db *gorm.DB) error {
	defaults := []model.SiteSetting{
		{Key: "site.title", Value: "Vivid"},
		{Key: "contact.qq", Value: "1114639355"},
		{Key: "contact.qq_link", Value: "https://qm.qq.com/q/ItgCcNA7ac"},
		{Key: "contact.qq_group", Value: "1106849765"},
		{Key: "contact.qq_group_link", Value: "https://qm.qq.com/q/976LeMFoHu"},
		{Key: "contact.email", Value: "vividairun@gmail.com"},
		{Key: "contact.shop", Value: "https://pay.ldxp.cn/shop/chiyi"},
		{Key: "auth.open", Value: "true"},
		{Key: "auth.email_code", Value: "false"},
		{Key: "auth.allow_password_reset", Value: "false"},
		{Key: "auth.allowed_email_domains", Value: ""},
		{Key: "auth.code_ttl_seconds", Value: "600"},
		{Key: "smtp.host", Value: ""},
		{Key: "smtp.port", Value: "587"},
		{Key: "smtp.username", Value: ""},
		{Key: "smtp.password", Value: ""},
		{Key: "smtp.from_addr", Value: ""},
		{Key: "smtp.use_tls", Value: "true"},
		{Key: "proxy.url", Value: ""},
		{Key: "credits.checkin_enabled", Value: "true"},
		{Key: "credits.checkin_reward", Value: "3"},
		{Key: "credits.invite_enabled", Value: "true"},
		{Key: "credits.invite_reward", Value: "3"},
		{Key: "logs.retention_days", Value: "30"},
		{Key: "media.retention_days", Value: "30"},
	}
	for _, item := range defaults {
		var count int64
		if err := db.WithContext(ctx).Model(&model.SiteSetting{}).Where("key = ?", item.Key).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.WithContext(ctx).Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}
