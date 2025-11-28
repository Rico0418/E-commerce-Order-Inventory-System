package events

import "time"

type OrderPlacedPayload struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Items     []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderResultPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"` 
	Reason  string `json:"reason,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
