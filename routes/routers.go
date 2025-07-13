package routes

import (
	"order/cmd/order/handler"
	"order/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, orderHandler handler.OrderHandler, jwtSecret string) {
	// context timeout and logger
	router.Use(middleware.RequestLogger(2))

	authMiddleware := middleware.AuthMiddleware(jwtSecret)
	private := router.Group("/v1/order")
	private.Use(authMiddleware)
	private.POST("/checkout", orderHandler.CheckoutOrder)
	private.GET("/history", orderHandler.GetOrderHistory)
}
