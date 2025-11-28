package main

import (
	"log"
	"os"

	"ecommerce-user/config"
	"ecommerce-user/handlers"
	"ecommerce-user/repositories"
	"ecommerce-user/usecase"
	"ecommerce-user/shared/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found — using system env")
	}
	db := config.GetDB()

	userRepo := repositories.NewGormUserRepo(db)
	userUC := usecase.NewUserUseCase(userRepo)
	userH := handlers.NewUserHandler(userUC)

	router := gin.Default()

	router.POST("/register", userH.Register)
	router.POST("/login", userH.Login)

	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", userH.GetProfile)
		protected.PUT("/profile", userH.UpdateProfile)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
