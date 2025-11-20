package infrastructures

import (
	"ecommerce-app/domain/users/entities"
	"log"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetDB() *gorm.DB {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL is required")
		}

		conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
		err = conn.AutoMigrate(&entities.User{})
		if err != nil {
			log.Fatalf("failed running migrations: %v", err)
		}

		log.Println("Database connected & migrated")
		db = conn
	})

	return db
}
