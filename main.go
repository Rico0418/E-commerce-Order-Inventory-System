package main

import (
	"context"
	"log"
	"os"

	"ecommerce-app/config"
	"ecommerce-app/domain/users/handlers"
	"ecommerce-app/domain/users/repositories"
	"ecommerce-app/domain/users/usecase"
	"ecommerce-app/shared/middleware"

	productHandlers "ecommerce-app/domain/products/handlers"
	productRepositories "ecommerce-app/domain/products/repositories"
	productUseCase "ecommerce-app/domain/products/usecase"

	orderHandlers "ecommerce-app/domain/orders/handlers"
	orderRepositories "ecommerce-app/domain/orders/repositories"
	orderUseCase "ecommerce-app/domain/orders/usecase"

	inventory "ecommerce-app/workers/inventory"
	notification "ecommerce-app/workers/notification"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found — using system env")
	}
	db := config.GetDB()
	redisClient := config.GetRedis()

	userRepo := repositories.NewGormUserRepo(db)
	userUC := usecase.NewUserUseCase(userRepo)
	userH := handlers.NewUserHandler(userUC)

	productRepo := productRepositories.NewGormProductRepo(db)
	productUC := productUseCase.NewProductUsecase(productRepo, redisClient)
	productH := productHandlers.NewProductHandler(productUC)

	orderRepo := orderRepositories.NewGormOrderRepo(db)
	orderUC := orderUseCase.NewOrderUsecase(orderRepo, productRepo, redisClient)
	orderHandler := orderHandlers.NewOrderHandler(orderUC)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := inventory.StartInventoryWorker(ctx, db, orderRepo, productRepo); err != nil {
		log.Fatalf("failed to start inventory worker: %v", err)
	}
	if err := notification.StartNotificationWorker(ctx); err != nil {
		log.Fatalf("failed to start notification worker: %v", err)
	}

	router := gin.Default()

	router.POST("/register", userH.Register)
	router.POST("/login", userH.Login)

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", userH.GetProfile)
		protected.PUT("/profile", userH.UpdateProfile)

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
