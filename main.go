package main

import (
	"context"
	"fmt"
	"order/cmd/order/handler"
	"order/cmd/order/repository"
	"order/cmd/order/resource"
	"order/cmd/order/service"
	"order/cmd/order/usecase"
	"order/config"
	"order/infrastructure/log"
	"order/kafka"
	kafkaConsumer "order/kafka/consumer"
	"order/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// init config
	cfg := config.LoadConfig()

	// init connection
	db := resource.InitDb(&cfg)

	// kafka producer init
	kafkaHost := fmt.Sprintf("%s:%s", cfg.Kafka.Host, cfg.App.Port)
	kafkaProducer := kafka.NewKafkaProducer([]string{kafkaHost})
	defer kafkaProducer.Close()

	// setup logger
	log.SetupLogger()

	// user setup
	orderRepository := repository.NewOrderRepository(db, cfg.Product.Host)
	orderService := service.NewOrderService(orderRepository)
	orderUsecase := usecase.NewOrderUsecase(orderService, kafkaProducer)
	orderHandler := handler.NewOrderHandler(orderUsecase)

	port := cfg.App.Port
	router := gin.Default()
	routes.SetupRoutes(router, *orderHandler, cfg.Jwt.Secret)
	router.Run(":" + port)

	// kafka consumer
	kafkaPaymentSuccessConsumer := kafkaConsumer.NewPaymentSuccessConsumer([]string{"localhost:9093"}, "payment.success", *orderService, *kafkaProducer)
	kafkaPaymentSuccessConsumer.Start(context.Background())

	kafkaPaymentFailedConsumer := kafkaConsumer.NewPaymentFailedConsumer([]string{"localhost:9093"}, "payment.failed", *orderService, *kafkaProducer)
	kafkaPaymentFailedConsumer.Start(context.Background())

	log.Logger.Printf("Server listening on port: %s", port)
}
