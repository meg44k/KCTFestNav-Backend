package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
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

func (uu *UserUsecase) Create(ctx context.Context, name string, loginID string, inputPassword []byte, assignedID int, role domain.Role) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(inputPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user, err := domain.NewUser(name, loginID, hashedPassword, assignedID, role)
	if err != nil {
		return err
	}
	return uu.userRepo.Create(ctx, user)
}

func (uu *UserUsecase) Authenticate(ctx context.Context, loginID string, inputPassword []byte) (*domain.User, error) {
	// データベースからパスワードを取得
	user, err := uu.userRepo.GetByLoginID(ctx, loginID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUnauthorized
		}
		return nil, err
	}

	// 入力されたパスワードと登録されているパスワードの比較
	if err := bcrypt.CompareHashAndPassword(user.Password, inputPassword); err != nil {
		return nil, ErrUnauthorized
	}
	return user, nil
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
