package api

import (
	"net/http"
	"time"

	"dalu-nongji-parts/backend/internal/model"
	"github.com/gin-gonic/gin"
)

func (h AdminHandler) ListNotifications(c *gin.Context) {
	var rows []model.UserNotification
	query := h.DB.Where("user_id = ?", c.GetUint("userId")).Order("id desc").Limit(200)
	if c.Query("unread") == "true" {
		query = query.Where("read_at IS NULL")
	}
	if query.Find(&rows).Error != nil {
		Fail(c, 500, 500, "消息加载失败")
		return
	}
	var unread int64
	h.DB.Model(&model.UserNotification{}).Where("user_id = ? AND read_at IS NULL", c.GetUint("userId")).Count(&unread)
	OK(c, gin.H{"items": rows, "unreadCount": unread})
}

func (h AdminHandler) ReadNotification(c *gin.Context) {
	now := time.Now()
	result := h.DB.Model(&model.UserNotification{}).Where("id = ? AND user_id = ?", c.Param("id"), c.GetUint("userId")).Update("read_at", &now)
	if result.Error != nil || result.RowsAffected != 1 {
		Fail(c, http.StatusNotFound, 404, "消息不存在")
		return
	}
	OK(c, gin.H{"read": true})
}

func (h AdminHandler) ReadAllNotifications(c *gin.Context) {
	now := time.Now()
	if h.DB.Model(&model.UserNotification{}).Where("user_id = ? AND read_at IS NULL", c.GetUint("userId")).Update("read_at", &now).Error != nil {
		Fail(c, 500, 500, "消息更新失败")
		return
	}
	OK(c, gin.H{"read": true})
}
