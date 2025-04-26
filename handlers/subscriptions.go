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

func (h *HandlerDependencies) CreateSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		var sub models.Subscription
		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub.ID = uuid.New()
		if err := h.DB.Create(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, sub)
	}
}

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
