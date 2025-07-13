package handler

import (
	"net/http"
	"order/cmd/order/usecase"
	"order/infrastructure/log"
	"order/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OrderHandler struct {
	OrderUsecase *usecase.OrderUsecase
}

func NewOrderHandler(orderUsecase *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{
		OrderUsecase: orderUsecase,
	}
}

func (h *OrderHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "Ok",
	})
}

func (h *OrderHandler) CheckoutOrder(c *gin.Context) {
	var param models.CheckoutRequest

	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error_message": "Invalid request.",
			"error_detail":  err.Error(),
		})

		return
	}

	// auth session
	userIDstr, isExist := c.Get("user_id")
	if !isExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Unauthorized",
		})

		return
	}

	userID, ok := userIDstr.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Invalid user id",
		})

		return
	}

	if len(param.Items) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Invalid parameter",
		})

		return
	}

	param.UserID = int64(userID)
	orderID, err := h.OrderUsecase.CheckoutOrder(c.Request.Context(), &param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("h.OrderUsecase.CheckoutOrder() got error %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": "Internal server error",
			"error_detail":  err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error_message": "order created.", "order_id": orderID})

	return
}

func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
	var param models.OrderHistoryParam

	userIDstr, isExist := c.Get("user_id")
	if !isExist {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error_message": "Unauthorized",
		})

		return
	}

	userID, ok := userIDstr.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id"})

		return
	}

	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	param = models.OrderHistoryParam{
		UserID: int64(userID),
		Status: status,
	}

	orderHistory, err := h.OrderUsecase.GetOrderHistoryByUserID(c.Request.Context(), &param)
	if err != nil {
		log.Logger.WithFields(logrus.Fields{
			"param": param,
		}).Errorf("h.OrderUsecase.GetOrderHistoryByUserID() got error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error_message": "Internal server error",
			"error_detail":  err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"error_message": "success.", "data": orderHistory})

	return
}
