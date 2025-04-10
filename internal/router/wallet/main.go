package wallet

import (
	"context"
	"github.com/gin-gonic/gin"
	"paymentSystem/internal/models"
	"paymentSystem/internal/repository"
)

func Router(r *gin.RouterGroup, repository *repository.Repository) {

	r.POST("/wallet", func(c *gin.Context) {
		ctx := context.Background()
		var walletRequest *models.WalletRequest
		err := c.ShouldBindJSON(&walletRequest)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err = repository.ChangeBalance(ctx, walletRequest.WalletID, walletRequest.Amount, walletRequest.OperationType)
		if err != nil {
			c.JSON(409, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	r.GET("/wallets/:id", func(c *gin.Context) {
		id := c.Param("id")
		balance, err := repository.GetBalance(id)
		if err != nil {
			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"Balance": balance,
		})
	})

}
