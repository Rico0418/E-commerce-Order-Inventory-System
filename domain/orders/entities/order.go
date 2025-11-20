package entities

import (
	"time"
)

type OrderItem struct {
	ID        string `gorm:"primaryKey;size:36"`
	OrderID   string `gorm:"index;size:36"`
	ProductID string `gorm:"size:36;not null"`
	Quantity  int    `gorm:"not null"`
}

type Order struct {
	ID        string      `gorm:"primaryKey;size:36"`
	UserID    string      `gorm:"index;size:36;not null"`
	Status    string      `gorm:"size:50;not null"` 
	Total     float64     `gorm:"not null;default:0"`
	Items     []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
