package main

import (
	"fmt"
	"order/cmd/order/handler"
	"order/cmd/order/repository"
	"order/cmd/order/resource"
	"order/cmd/order/service"
	"order/cmd/order/usecase"
	"order/config"
	"order/infrastructure/log"
	"order/kafka"
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
	kafkaProducer := kafka.NewKafkaProducer([]string{kafkaHost}, "order.created")
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

	log.Logger.Printf("Server listening on port: %s", port)
}
