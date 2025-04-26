package handlers

import (
	"net/http"

	"github.com/Miku7676/webhook-delivery-service/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler struct {
	*gorm.DB
}

// type Subscription struct {
// 	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
// 	TargetURL string    `json:"target_url" binding:"required"`
// 	Secret    string    `json:"secret"`
// }

func CreateSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sub models.Subscription
		if err := c.ShouldBindJSON(&sub); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		sub.ID = uuid.New()
		if err := db.Create(&sub).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, sub)
	}
}

func GetSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var sub models.Subscription
		if err := db.First(&sub, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
			return
		}
		c.JSON(http.StatusOK, sub)
	}
}

func UpdateSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var sub models.Subscription
		if err := db.First(&sub, "id = ?", id).Error; err != nil {
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
		db.Save(&sub)
		c.JSON(http.StatusOK, sub)
	}
}

func DeleteSubscription(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := db.Delete(&models.Subscription{}, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	}
}
