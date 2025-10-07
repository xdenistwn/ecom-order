package kafka

import (
	"context"
	"encoding/json"
	"order/cmd/order/service"
	"order/infrastructure/constant"
	"order/infrastructure/log"
	"order/models"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentSuccessConsumer struct {
	Reader       *kafka.Reader
	Producer     KafkaProducer
	OrderService service.OrderService
}

func NewPaymentSuccessConsumer(brokers []string, topic string, orderService service.OrderService, kafkaProducer KafkaProducer) *PaymentSuccessConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order",
	})

	return &PaymentSuccessConsumer{
		Reader:       reader,
		OrderService: orderService,
		Producer:     kafkaProducer,
	}
}

func (c *PaymentSuccessConsumer) Start(ctx context.Context) {
	log.Logger.Println("[KAFKA] Listening to topic: payment.success")

	for {
		message, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Logger.Println("[KAFKA] Error Read Message: ", err)
			continue
		}

		var event models.PaymentSuccessEvent
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Logger.Println("[KAFKA] EArror Unmarshal event message value: ", err)
			continue
		}

		log.Logger.Printf("[KAFKA] Received payment.success event for Order ID %d\n", event.OrderID)

		// update DB
		err = c.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCompleted)
		if err != nil {
			log.Logger.Println("[KAFKA] Error update order status")
			continue
		}

		// get order info from db
		order, err := c.OrderService.GetOrderInfoByOrderID(ctx, event.OrderID)
		if err != nil {
			log.Logger.Println("[KAFKA] Error get order info by order id")
			continue
		}

		// get order detail from db
		orderDetail, err := c.OrderService.GetOrderDetailByID(ctx, order.OrderDetailID)
		if err != nil {
			log.Logger.Println("[KAFKA] Error get order detail by id")
			continue
		}

		// get product list from order detail
		var products []models.CheckoutItem
		err = json.Unmarshal([]byte(orderDetail.Products), &products)
		if err != nil {
			log.Logger.Println("[KAFKA] Error unmarshal product list from order detail")
			continue
		}

		// public event product service
		err = c.Producer.PublishProductStockUpdate(ctx, models.ProductStockUpdateEvent{
			OrderID:   event.OrderID,
			Products:  convertCheckoutItemsToProductItems(products),
			EventTime: time.Now(),
		})
	}
}

func convertCheckoutItemsToProductItems(items []models.CheckoutItem) []models.ProductItem {
	result := make([]models.ProductItem, len(items))

	for index, item := range items {
		result[index] = models.ProductItem{
			ProductID: item.ProductID,
			Qty:       item.Quantity,
		}
	}

	return result
}
