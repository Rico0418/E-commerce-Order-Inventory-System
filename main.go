package main

import (
	"log"
	"os"

	"ecommerce-app/domain/users/handlers"
	"ecommerce-app/domain/users/repositories"
	"ecommerce-app/domain/users/usecase"
	"ecommerce-app/config"
	"ecommerce-app/shared/middleware"


	productHandlers "ecommerce-app/domain/products/handlers"
	productRepositories "ecommerce-app/domain/products/repositories"
	productUseCase "ecommerce-app/domain/products/usecase"
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
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
