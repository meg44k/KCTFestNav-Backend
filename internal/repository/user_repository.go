package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

type userRepository struct {
	db    *database.Queries
	cache *redis.Client
}

func NewUserRepository(db *sql.DB, rdb *redis.Client) domain.UserRepository {
	return &userRepository{
		db:    database.New(db),
		cache: rdb,
	}
}

func (ur *userRepository) Create(ctx context.Context, u *domain.User) error {
	arg := database.CreateUserParams{
		ID:      u.ID.String(),
		Name:    u.Name,
		LoginID: u.LoginID,
		AssignedBoothID: sql.NullInt32{
			Int32: int32(u.AssignedBoothID),
			Valid: u.AssignedBoothID != 0,
		},
		Password: string(u.Password),
		Role:     string(u.Role),
	}
	return ur.db.CreateUser(ctx, arg)
}

func (ur *userRepository) Update(ctx context.Context, u *domain.User) error {
	// TODO: 実際の処理を書く
	return nil
}

func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO: 実際の処理を書く
	return nil
}

func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	// TODO: 実際の処理を書く
	return nil, nil
}

func (ur *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	// TODO: 実際の処理を書く
	return nil, nil
}

func (ur *userRepository) GetByLoginID(ctx context.Context, loginID string) (*domain.User, error) {
	dbUser, err := ur.db.GetUserByLoginID(ctx, loginID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	parsedID, err := uuid.Parse(dbUser.ID)
	if err != nil {
		return nil, err
	}

	user, err := domain.ReconstructUser(
		parsedID,
		dbUser.Name,
		dbUser.LoginID,
		[]byte(dbUser.Password),
		int(dbUser.AssignedBoothID.Int32),
		domain.Role(dbUser.Role),
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
