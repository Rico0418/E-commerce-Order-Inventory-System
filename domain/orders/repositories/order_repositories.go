package repositories

import (
	"ecommerce-app/domain/orders/entities"
	"errors"

	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository interface {
	Create(order *entities.Order) error
	FindByID(id string) (*entities.Order, error)
	Update(order *entities.Order) error
}

type GormOrderRepo struct {
	db *gorm.DB
}

func NewGormOrderRepo(db *gorm.DB) *GormOrderRepo {
	return &GormOrderRepo{db}
}

func (r *GormOrderRepo) Create(order *entities.Order) error {
	return r.db.Create(order).Error
}

func (r *GormOrderRepo) FindByID(id string) (*entities.Order, error) {
	var order entities.Order
	err := r.db.Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

func (r *GormOrderRepo) Update(order *entities.Order) error {
	return r.db.Save(order).Error
}
