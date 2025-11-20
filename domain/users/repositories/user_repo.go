package repositories

import (
	"errors"
	"ecommerce-app/domain/users/entities"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(u *entities.User) error
	FindByEmail(email string) (*entities.User, error)
	FindByID(id string) (*entities.User, error)
	Update(u *entities.User) error
}

type GormUserRepo struct {
	db *gorm.DB
}

func NewGormUserRepo(db *gorm.DB) *GormUserRepo {
	return &GormUserRepo{db}
}

func (r *GormUserRepo) Create(u *entities.User) error {
	return r.db.Create(u).Error
}

func (r *GormUserRepo) FindByEmail(email string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (r *GormUserRepo) FindByID(id string) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (r *GormUserRepo) Update(u *entities.User) error {
	return r.db.Save(u).Error
}
