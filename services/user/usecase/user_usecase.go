package usecase

import (
	"errors"
	"os"
	"strconv"
	"time"

	"ecommerce-user/entities"
	modelsRequest "ecommerce-user/models/request"
	modelsResponse "ecommerce-user/models/response"
	"ecommerce-user/repositories"
	"ecommerce-user/shared/security"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailExists = errors.New("email already registered")

type UserUsecase struct {
	repo repositories.UserRepository
}

func NewUserUseCase(repo repositories.UserRepository) *UserUsecase {
	return &UserUsecase{repo}
}

func (uc *UserUsecase) Register(req *modelsRequest.RegisterRequest) (*entities.User, error) {
	_, err := uc.repo.FindByEmail(req.Email)
	if err == nil {
		return nil, ErrEmailExists
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	user := &entities.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := uc.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) Login(req *modelsRequest.LoginRequest) (string, time.Time, error) {
	user, err := uc.repo.FindByEmail(req.Email)
	if err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	expiry := 60
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if parsed, _ := strconv.Atoi(v); parsed > 0 {
			expiry = parsed
		}
	}

	token, exp, err := security.GenerateToken(user.ID, time.Duration(expiry)*time.Minute)
	return token, exp, err
}

func (uc *UserUsecase) GetProfile(id string) (*modelsResponse.ProfileResponse, error) {
	user, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &modelsResponse.ProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (uc *UserUsecase) UpdateProfile(id string, req *modelsRequest.UpdateProfileRequest) (*modelsResponse.ProfileResponse, error) {
	user, err := uc.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	_ = uc.repo.Update(user)

	return &modelsResponse.ProfileResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}
