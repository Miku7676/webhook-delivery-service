package handlers

import (
	"net/http"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func GetDeliveryStatusByWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		webhookID := c.Param("webhook_id")

		id, err := uuid.Parse(webhookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			return
		}

		var logs []models.DeliveryLog
		if err := db.Where("webhook_task_id = ?", id).Order("created_at asc").Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}

func GetRecentLogsBySubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		subID := c.Param("id")

		id, err := uuid.Parse(subID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		var logs []models.DeliveryLog
		if err := db.Where("subscription_id = ?", id).Order("created_at desc").Limit(20).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}
