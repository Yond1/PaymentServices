package router

import (
	"github.com/gin-gonic/gin"
	"paymentSystem/internal/repository"
	"paymentSystem/internal/router/wallet"
)

func SetupRoutes(r *gin.Engine, repo *repository.Repository) {
	v1 := r.Group("api/v1")
	wallet.Router(v1, repo)
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})
}
