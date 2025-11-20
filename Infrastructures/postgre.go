package infrastructures

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ( 
	dbPool *pgxpool.Pool
	once sync.Once
)
func InitPostgres() *pgxpool.Pool {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL environment variable is required")
		}

		cfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("Failed parsing PostgreSQL DSN: %v", err)
		}

		cfg.MaxConns = 10
		cfg.MinConns = 2
		cfg.MaxConnLifetime = time.Hour
		cfg.HealthCheckPeriod = time.Minute * 1

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			log.Fatalf("Failed connecting to PostgreSQL: %v", err)
		}

		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("PostgreSQL ping failed: %v", err)
		}

		log.Println("Connected to PostgreSQL ✨")
		dbPool = pool
	})

	return dbPool
}

func ClosePostgres() {
	if dbPool != nil {
		dbPool.Close()
		log.Println("PostgreSQL connection closed")
	}
}