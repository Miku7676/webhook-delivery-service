package handlers

import (
	"net/http"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetDeliveryStatusByWebhook godoc
// @Summary Get delivery status for a webhook
// @Description Fetches all delivery attempts for a specific webhook ID
// @Tags Status
// @Produce json
// @Param webhook_id path string true "Webhook Task ID"
// @Success 200 {array} models.DeliveryLog
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /status/{webhook_id} [get]
func GetDeliveryStatusByWebhook(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract webhook ID
		webhookID := c.Param("webhook_id")

		id, err := uuid.Parse(webhookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			return
		}

		// fetch logs by webhook_task_id - latest first
		var logs []models.DeliveryLog
		if err := db.Where("webhook_task_id = ?", id).Order("created_at asc").Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}

// GetRecentLogsBySubscription godoc
// @Summary Get recent delivery logs for a subscription
// @Description Lists the last 20 delivery attempts for a subscription
// @Tags Status
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {array} models.DeliveryLog
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions/{id}/logs [get]
func GetRecentLogsBySubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		subID := c.Param("id")

		id, err := uuid.Parse(subID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
			return
		}

		// fetches most recent logs (last 20 logs) by the subscription id
		var logs []models.DeliveryLog
		if err := db.Where("subscription_id = ?", id).Order("created_at desc").Limit(20).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}
