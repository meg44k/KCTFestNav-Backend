package repository

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type liveRepository struct {
	db    *database.Queries
	cache *redis.Client
}

func NewLiveRepository(db *sql.DB, rdb *redis.Client) domain.LiveRepository {
	return &liveRepository{
		db:    database.New(db),
		cache: rdb,
	}
}

func (lr *liveRepository) Create(ctx context.Context, l *domain.Live) error {
	// TODO:実際の実装を書く
	return nil
}

func (lr *liveRepository) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	// TODO: 実装を行う
	return nil, nil
}

func (lr *liveRepository) GetAll(ctx context.Context) ([]*domain.Live, error) {
	// TODO: 実装を行う
	return nil, nil

}

func (lr *liveRepository) Update(ctx context.Context, l *domain.Live) error {
	// TODO: 実装を行う
	return nil
}

func (lr *liveRepository) Delete(ctx context.Context, id int) error {
	// TODO: 実装を行う
	return nil
}
