package usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"ecommerce-order-product/models/response"
	"ecommerce-order-product/repositories"

	"github.com/redis/go-redis/v9"
	"context"
)

type ProductUsecase struct {
	repo     repositories.ProductRepository
	cache    *redis.Client
}

func NewProductUsecase(repo repositories.ProductRepository, cache *redis.Client) *ProductUsecase {
	return &ProductUsecase{repo, cache}
}

func (uc *ProductUsecase) GetProducts(name, category string) ([]response.ProductResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("products:%s:%s", name, category)

	if data, err := uc.cache.Get(ctx, cacheKey).Result(); err == nil {
		var cached []response.ProductResponse
		_ = json.Unmarshal([]byte(data), &cached)
		return cached, nil
	}

	products, err := uc.repo.FindAll(name, category)
	if err != nil {
		return nil, err
	}

	res := make([]response.ProductResponse, 0)
	for _, p := range products {
		res = append(res, response.ProductResponse{
			ID: p.ID, Name: p.Name, Category: p.Category, Price: p.Price, Stock: p.Stock,
			CreatedAt: p.CreatedAt,
		})
	}

	bytes, _ := json.Marshal(res)
	uc.cache.Set(ctx, cacheKey, bytes, 5*time.Minute)

	return res, nil
}

func (uc *ProductUsecase) GetProduct(id string) (*response.ProductResponse, error) {
	p, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &response.ProductResponse{
		ID: p.ID, Name: p.Name, Category: p.Category, Price: p.Price, Stock: p.Stock,
		CreatedAt: p.CreatedAt,
	}, nil
}