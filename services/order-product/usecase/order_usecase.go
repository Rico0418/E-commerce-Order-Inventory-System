package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	orderEntities "ecommerce-order-product/entities"
	orderModelsRequest "ecommerce-order-product/models/request"
	orderModelsResponse "ecommerce-order-product/models/response"
	productRepo "ecommerce-order-product/repositories"
	orderRepo "ecommerce-order-product/repositories"
	"ecommerce-order-product/config"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrEmptyItems = errors.New("order items cannot be empty")

type OrderUsecase struct {
	orderRepo   orderRepo.OrderRepository
	productRepo productRepo.ProductRepository
	redis       *redis.Client
}

func NewOrderUsecase(or orderRepo.OrderRepository, pr productRepo.ProductRepository, r *redis.Client) *OrderUsecase {
	return &OrderUsecase{
		orderRepo:   or,
		productRepo: pr,
		redis:       r,
	}
}

type OrderPlacedPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Items   []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID string, req *orderModelsRequest.CreateOrderRequest) (string, error) {
	if len(req.Items) == 0 {
		return "", ErrEmptyItems
	}

	for _, it := range req.Items {
		if _, err := uc.productRepo.FindByID(it.ProductID); err != nil {
			return "", err
		}
	}


	orderID := uuid.NewString()
	order := &orderEntities.Order{
		ID:     orderID,
		UserID: userID,
		Status: "PENDING",
		Items:  make([]orderEntities.OrderItem, 0, len(req.Items)),
	}

	var total float64
	for _, it := range req.Items {
		p, err := uc.productRepo.FindByID(it.ProductID)
		if err != nil {
			return "", err
		}
		total += p.Price * float64(it.Quantity)
		order.Items = append(order.Items, orderEntities.OrderItem{
			ID:        uuid.NewString(),
			OrderID:   orderID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}
	order.Total = total

	if err := uc.orderRepo.Create(order); err != nil {
		return "", err
	}

	go func() {
		ch, err := config.NewChannel()
		if err != nil {

			return
		}
		defer ch.Close()

		exchange := os.Getenv("RABBITMQ_EXCHANGE")
		if exchange == "" {
			exchange = "orders_direct"
		}
		routingKey := os.Getenv("RABBITMQ_ROUTING_KEY")
		if routingKey == "" {
			routingKey = "order.placed"
		}
		queue := os.Getenv("RABBITMQ_QUEUE")
		if queue == "" {
			queue = "order_placed_queue"
		}

		if err := config.EnsureDirectExchange(ch, exchange); err != nil {
			return
		}
		_, _ = config.DeclareQuorumQueue(ch, queue, exchange, routingKey)

		payload := OrderPlacedPayload{
			OrderID: orderID,
			UserID:  userID,
			CreatedAt: time.Now().UTC(),
		}
		for _, it := range req.Items {
			payload.Items = append(payload.Items, struct {
				ProductID string `json:"product_id"`
				Quantity  int    `json:"quantity"`
			}{
				ProductID: it.ProductID,
				Quantity:  it.Quantity,
			})
		}
		b, _ := json.Marshal(payload)
		_ = config.PublishJSON(ch, exchange, routingKey, b)
	}()

	return orderID, nil
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, userID, orderID string) (*orderModelsResponse.OrderResponse, error) {
	order, err := uc.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("not authorized to view this order")
	}

	resp := &orderModelsResponse.OrderResponse{
		ID:     order.ID,
		UserID: order.UserID,
		Status: order.Status,
		Total:  order.Total,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		Items:  []orderModelsResponse.OrderItemResponse{},
	}
	for _, it := range order.Items {
		resp.Items = append(resp.Items, orderModelsResponse.OrderItemResponse{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
	}
	return resp, nil
}
