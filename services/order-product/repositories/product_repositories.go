package repositories

import (
	"ecommerce-order-product/entities"

	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll(name, category string) ([]entities.Product, error)
	FindByID(id string) (*entities.Product, error)
}
type GormProductRepo struct {
	db *gorm.DB
}

func NewGormProductRepo(db *gorm.DB) ProductRepository {
	return &GormProductRepo{db}
}
func (r *GormProductRepo) FindAll(name, category string) ([]entities.Product, error) {
	var products []entities.Product
	q := r.db.Model(&entities.Product{})

	if name != "" {
		q = q.Where("name ILIKE ?", "%"+name+"%")
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	err := q.Find(&products).Error
	return products, err
}

func (r *GormProductRepo) FindByID(id string) (*entities.Product, error) {
	var p entities.Product
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}