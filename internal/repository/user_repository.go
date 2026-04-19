package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type userRepositoryImpl struct {}

func NewUserRepository() domain.UserRepository{
	return &userRepositoryImpl{}
}

func (ur *userRepositoryImpl) Create(ctx context.Context,u *domain.User) error{
	// TODO: 実際の処理を書く
	return nil
}

func (ur *userRepositoryImpl) Update(ctx context.Context, u *domain.User) error {
	// TODO: 実際の処理を書く
	return nil
}

func (ur *userRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: 実際の処理を書く
	return nil
} 

func (ur *userRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error){
	// TODO: 実際の処理を書く
	return nil, nil
}

func (ur *userRepositoryImpl) GetAll(ctx context.Context) ([]*domain.User, error) {
	// TODO: 実際の処理を書く
	return nil, nil
}
