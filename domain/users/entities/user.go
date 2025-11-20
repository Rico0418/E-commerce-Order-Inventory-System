package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Name         string    `gorm:"size:100" json:"name"`
	Email        string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New().String()
	return nil
}