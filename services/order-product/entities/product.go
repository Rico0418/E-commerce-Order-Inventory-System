package entities

import "time"

type Product struct {
	ID          string  `gorm:"primaryKey"`
	Name        string  `gorm:"size:255;not null"`
	Category    string  `gorm:"size:100"`
	Description string  `gorm:"size:500"`
	Price       float64 `gorm:"not null"`
	Stock       int     `gorm:"not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}