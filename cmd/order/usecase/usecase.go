package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"order/cmd/order/service"
	"order/infrastructure/constant"
	"order/infrastructure/log"
	"order/kafka"
	"order/models"
	"time"

	"github.com/sirupsen/logrus"
)

type OrderUsecase struct {
	OrderService  *service.OrderService
	KafkaProducer *kafka.KafkaProducer
}

func NewOrderUsecase(orderService *service.OrderService, kafkaProducer *kafka.KafkaProducer) *OrderUsecase {
	return &OrderUsecase{
		OrderService:  orderService,
		KafkaProducer: kafkaProducer,
	}
}

func (uc *OrderUsecase) CheckoutOrder(ctx context.Context, param *models.CheckoutRequest) (int64, error) {
	var orderID int64

	// check idempotency token
	if param.IdempontencyToken != "" {
		isExist, err := uc.OrderService.CheckIdempotency(ctx, param.IdempontencyToken)
		if err != nil {
			return 0, err
		}

		if isExist {
			return 0, errors.New("Order already created, please check again!")
		}
	}

	// validate product
	err := uc.validateProduct(ctx, param.Items)
	if err != nil {
		return 0, err
	}

	// calculate product amount
	totalQty, totalAmount := uc.calculateOrderSummary(param.Items)

	// construct order detail
	products, orderHistory := uc.constructOrderDetail(param.Items)

	// save order dan order detail
	orderDetail := models.OrderDetail{
		Products:     products,
		OrderHistory: orderHistory,
	}

	order := models.Order{
		UserID:          param.UserID,
		Amount:          totalAmount,
		TotalQty:        totalQty,
		Status:          constant.OrderStatusCreated,
		PaymentMethod:   param.PaymentMethod,
		ShippingAddress: param.ShippingAddress,
	}

	orderID, err = uc.OrderService.SaveOrderAndOrderDetail(ctx, &order, &orderDetail)
	if err != nil {
		return 0, err
	}

	// save idempontecy token
	if param.IdempontencyToken != "" {
		err = uc.OrderService.SaveIdempontency(ctx, param.IdempontencyToken)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"err":   err.Error(),
				"token": param.IdempontencyToken,
			}).Info("uc.OrderService.SaveIdempotency() got error")
		}
	}

	// publish to payment service
	orderCreatedEvent := models.OrderCreatedEvent{
		OrderID:         orderID,
		UserID:          param.UserID,
		TotalAmount:     totalAmount,
		TotalQty:        totalQty,
		PaymentMethod:   param.PaymentMethod,
		ShippingAddress: param.ShippingAddress,
	}

	go func(ctx context.Context) {
		if err := uc.KafkaProducer.PublishOrderCreated(ctx, orderCreatedEvent); err != nil {
			fmt.Println("Failed publish order created event:", err)
		} else {
			fmt.Println("Successfully publish order created event for Order ID:", orderID)
		}
	}(context.Background())

	updateStockEvent := models.ProductStockUpdateEvent{
		OrderID:   orderID,
		Products:  convertCheckoutItemToProductItems(param.Items),
		EventTime: time.Now(),
	}

	go func(ctx context.Context) {
		if err := uc.KafkaProducer.PublishProductStockUpdate(ctx, updateStockEvent); err != nil {
			fmt.Println("Failed publish product stock update event:", err)
		} else {
			fmt.Println("Successfully publish product stock update event for Order ID:", orderID)
		}
	}(context.Background())

	return orderID, nil
}

func convertCheckoutItemToProductItems(items []models.CheckoutItem) []models.ProductItem {
	result := make([]models.ProductItem, 0, len(items))

	for index, item := range items {
		result[index] = models.ProductItem{
			ProductID: item.ProductID,
			Qty:       item.Quantity,
		}
	}
	return result
}

func (uc *OrderUsecase) validateProduct(ctx context.Context, items []models.CheckoutItem) error {
	seen := map[int64]bool{}
	for _, item := range items {
		// check duplicate
		if seen[item.ProductID] {
			return fmt.Errorf("Duplicate product: %d", item.ProductID)
		}

		// get product info at product service
		productInfo, err := uc.OrderService.GetProductInfo(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("Failed get product info: %d, err : %v", item.ProductID, err)
		}

		// quantity
		if item.Quantity <= 0 || item.Quantity > 1000 {
			return fmt.Errorf("invalid quantity product %d, maximum is 1000", item.ProductID)
		}

		// price
		if item.Price != productInfo.Price {
			return fmt.Errorf("Invalid price for product %d", item.ProductID)
		}

		// check stock
		if item.Quantity > productInfo.Stock {
			return fmt.Errorf("Invalid product quantity %d, stock left %d", item.ProductID, productInfo.Stock)
		}
	}

	return nil
}

func (uc *OrderUsecase) calculateOrderSummary(items []models.CheckoutItem) (int, float64) {
	var totalQty int
	var totalAmount float64

	for _, item := range items {
		totalQty += item.Quantity
		totalAmount += float64(item.Quantity) * item.Price
	}

	return totalQty, totalAmount
}

func (uc *OrderUsecase) constructOrderDetail(items []models.CheckoutItem) (string, string) {
	// products, order history
	productJSON, _ := json.Marshal(items)
	history := []map[string]any{
		{"status": "created", "timestamp": time.Now()},
	}

	historyJSON, _ := json.Marshal(history)

	return string(productJSON), string(historyJSON)
}

func (uc *OrderUsecase) GetOrderHistoryByUserID(ctx context.Context, param *models.OrderHistoryParam) ([]models.OrderHistoryResponse, error) {
	orderHistory, err := uc.OrderService.GetOrderHistoryByUserID(ctx, param)
	if err != nil {
		return nil, err
	}

	return orderHistory, nil
}
