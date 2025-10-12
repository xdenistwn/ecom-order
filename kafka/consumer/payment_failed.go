package consumer

import (
	"context"
	"encoding/json"
	"order/cmd/order/service"
	"order/infrastructure/constant"
	"order/infrastructure/log"
	kafkaOrder "order/kafka"
	"order/models"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentFailedEvent struct {
	Reader       *kafka.Reader
	Producer     kafkaOrder.KafkaProducer
	OrderService service.OrderService
}

func NewPaymentFailedConsumer(brokers []string, topic string, orderService service.OrderService, kafkaProducer kafkaOrder.KafkaProducer) *PaymentFailedEvent {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "order",
	})

	return &PaymentFailedEvent{
		Reader:       reader,
		OrderService: orderService,
		Producer:     kafkaProducer,
	}
}

func (c *PaymentFailedEvent) Start(ctx context.Context) {
	log.Logger.Println("[KAFKA] Listening to topic: payment.failed")

	for {
		message, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			log.Logger.Println("[KAFKA] Error Read Message: ", err)
			continue
		}

		var event models.PaymentUpdateStatusEvent
		err = json.Unmarshal(message.Value, &event)
		if err != nil {
			log.Logger.Println("[KAFKA] EArror Unmarshal event message value: ", err)
			continue
		}

		// update DB status order
		err = c.OrderService.UpdateOrderStatus(ctx, event.OrderID, constant.OrderStatusCancelled)
		if err != nil {
			log.Logger.Println("[KAFKA] Error update order status")
			continue
		}

		// order info
		orderInfo, err := c.OrderService.GetOrderInfoByOrderID(ctx, event.OrderID)
		if err != nil {
			log.Logger.Println("[KAFKA] Error get order info by order id")
			continue
		}

		// order detail info
		orderDetailInfo, err := c.OrderService.GetOrderDetailByID(ctx, orderInfo.OrderDetailID)
		if err != nil {
			log.Logger.Println("[KAFKA] Error get order detail by order detail id")
			continue
		}

		// construct product
		productStockUpdate := make([]models.ProductItem, 0)
		err = json.Unmarshal([]byte(orderDetailInfo.Products), &productStockUpdate)
		if err != nil {
			log.Logger.Println("[KAFKA] Error unmarshal product from order detail")
			continue
		}

		// publish event product stock.rollback
		updateStockEvent := models.ProductStockUpdateEvent{
			OrderID:   event.OrderID,
			Products:  productStockUpdate,
			EventTime: time.Now(),
		}

		err = c.Producer.PublishProductStockRollback(ctx, updateStockEvent)
		if err != nil {
			log.Logger.Println("[KAFKA] Error publish product stock rollback event")
			continue
		}
	}
}
