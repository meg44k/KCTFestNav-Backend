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
	arg := database.UpdateUserParams{
		LoginID: u.LoginID,
		Name:    u.Name,
		AssignedBoothID: sql.NullInt32{
			Int32: int32(u.AssignedBoothID),
			Valid: u.AssignedBoothID != 0,
		},
		Password: string(u.Password),
		Role:     string(u.Role),
		ID:       u.ID.String(),
	}
	return ur.db.UpdateUser(ctx, arg)
}

func (ur *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return ur.db.DeleteUser(ctx, id.String())
}

func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	dbUser, err := ur.db.GetUserByID(ctx, id.String())
	if err != nil {
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

func (ur *userRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	dbUsers, err := ur.db.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, len(dbUsers))
	for i, u := range dbUsers {
		parsedID, err := uuid.Parse(u.ID)
		if err != nil {
			return nil, err
		}
		user, err := domain.ReconstructUser(
			parsedID,
			u.Name,
			u.LoginID,
			[]byte(u.Password),
			int(u.AssignedBoothID.Int32),
			domain.Role(u.Role),
		)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
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
