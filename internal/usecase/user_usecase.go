package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type UserUsecase struct {
	userRepo domain.UserRepository
}

// contextにJWTから読み取ったRoleを入れるためのキー
// contextはkey-valueで入ってる
type contextKey string

const ContextRequestUserKey contextKey = "requestUser"

type RequestUser struct {
	ID              uuid.UUID
	Role            domain.Role
	AssignedBoothID int
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: repo,
	}
}

func (uu *UserUsecase) Create(ctx context.Context, user *domain.User) error {
	return uu.userRepo.Create(ctx, user)
}

func (uu *UserUsecase) Update(ctx context.Context, user *domain.User) error {
	return uu.userRepo.Update(ctx, user)
}

func (uu *UserUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return uu.userRepo.Delete(ctx, id)
}

func (uu *UserUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, err := uu.userRepo.GetByID(ctx, id); err != nil {
		return user, nil
	}
	return nil, nil
}

func (uu *UserUsecase) GetAll(ctx context.Context) ([]*domain.User, error) {
	if users, err := uu.userRepo.GetAll(ctx); err != nil {
		return users, nil
	}
	return nil, nil
}
