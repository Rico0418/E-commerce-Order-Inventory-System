package main

import (
	"context"
	"log"
	"os"

	"ecommerce-order-product/config"
	"ecommerce-order-product/shared/middleware"

	"ecommerce-order-product/handlers"
	"ecommerce-order-product/repositories"
	"ecommerce-order-product/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"ecommerce-order-product/workers/inventory"
	"ecommerce-order-product/workers/notification"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found — using system env")
	}
	db := config.GetDB()
	redisClient := config.GetRedis()

	productRepo := repositories.NewGormProductRepo(db)
	productUC := usecase.NewProductUsecase(productRepo, redisClient)
	productH := handlers.NewProductHandler(productUC)

	orderRepo := repositories.NewGormOrderRepo(db)
	orderUC := usecase.NewOrderUsecase(orderRepo, productRepo, redisClient)
	orderHandler := handlers.NewOrderHandler(orderUC)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := inventory.StartInventoryWorker(ctx, db, orderRepo, productRepo); err != nil {
		log.Fatalf("failed to start inventory worker: %v", err)
	}
	if err := notification.StartNotificationWorker(ctx); err != nil {
		log.Fatalf("failed to start notification worker: %v", err)
	}

	router := gin.Default()

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/products", productH.GetProducts)
		protected.GET("/products/:id", productH.GetProduct)

		protected.POST("/orders", orderHandler.CreateOrder)
		protected.GET("/orders/:id", orderHandler.GetOrder)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
