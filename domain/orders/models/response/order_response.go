package response

import "time"

type OrderItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderResponse struct {
	ID        string               `json:"id"`
	UserID    string               `json:"user_id"`
	Status    string               `json:"status"`
	Total     float64              `json:"total"`
	Items     []OrderItemResponse  `json:"items"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}