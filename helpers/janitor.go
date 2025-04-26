package helpers

import (
	"log"
	"time"

	"github.com/Miku7676/webhook-delivery-service/models"
	"gorm.io/gorm"
)

func StartLogClean(db *gorm.DB) {
	ticker := time.NewTicker(6 * time.Hour) // 1*time.hour
	defer ticker.Stop()

	for range ticker.C {
		cleanupLogs(db)
	}
}

func cleanupLogs(db *gorm.DB) {
	cutoff := time.Now().Add(-24 * time.Hour)
	result := db.Where("created_at < ?", cutoff).Delete(&models.DeliveryLog{})

	if result.Error != nil {
		log.Printf("Failed to cleanup delivery logs: %v", result.Error)
		return
	}
	log.Printf("Cleaned up %d old delivery logs.", result.RowsAffected)
}
