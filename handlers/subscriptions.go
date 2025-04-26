package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HandlerDependencies struct {
	DB          *gorm.DB
	RedisClient *redis.Client
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Creates a subscription with a target URL and optional secret
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param subscription body models.CreateSubscriptionRequest true "Create Subscription Request Body"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /subscriptions [post]
func (h *HandlerDependencies) CreateSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateSubscriptionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub := models.Subscription{
			ID:        uuid.New(),
			TargetURL: req.TargetURL,
			Secret:    req.Secret,
		}
		if err := h.DB.Create(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, sub)
	}
}

// GetSubscription godoc
// @Summary Get a subscription
// @Description Retrieves a subscription by ID
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} models.Subscription
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [get]
func (h *HandlerDependencies) GetSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var sub models.Subscription
		if err := h.DB.First(&sub, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

// UpdateSubscription godoc
// @Summary Update a subscription
// @Description Updates the target URL or secret of a subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param subscription body models.Subscription true "Updated Subscription object"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [put]
func (h *HandlerDependencies) UpdateSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var sub models.Subscription
		if err := h.DB.First(&sub, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		var updateData models.Subscription
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub.TargetURL = updateData.TargetURL
		sub.Secret = updateData.Secret
		h.DB.Save(&sub)

		// Update cache
		subBytes, _ := json.Marshal(sub)
		cacheKey := fmt.Sprintf("subscription:%s", sub.ID)
		h.RedisClient.Set(c.Request.Context(), cacheKey, subBytes, time.Hour)

		c.JSON(http.StatusOK, sub)
	}
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Deletes a subscription by ID
// @Tags Subscriptions
// @Param id path string true "Subscription ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Router /subscriptions/{id} [delete]
func (h *HandlerDependencies) DeleteSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.DB.Delete(&models.Subscription{}, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)

		cacheKey := fmt.Sprintf("subscription:%s", id)
		h.RedisClient.Del(c.Request.Context(), cacheKey)
	}
}
