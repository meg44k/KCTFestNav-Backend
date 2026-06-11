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
	arg := database.CreateLiveParams{
		Name:      l.Name,
		StartTime: l.StartTime,
		EndTime:   l.EndTime,
		Status:    int8(l.Status()),
		Detail: sql.NullString{
			String: l.Detail,
			Valid:  l.Detail != "",
		},
		Thumbnailurl: sql.NullString{
			String: l.ThumbnailURL,
			Valid:  l.ThumbnailURL != "",
		},
		SessionNumber: sql.NullInt16{
			Int16: int16(l.SessionNumber()),
			Valid: true,
		},
	}
	err := lr.db.CreateLive(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (lr *liveRepository) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	dbLive, err := lr.db.GetLiveByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	live, err := domain.ReconstructLive(
		int(dbLive.ID),
		dbLive.Name,
		dbLive.Detail.String,
		dbLive.Thumbnailurl.String,
		dbLive.StartTime,
		dbLive.EndTime,
		int8(dbLive.SessionNumber.Int16),
		domain.LiveStatus(dbLive.Status))
	return live, nil
}

func (lr *liveRepository) GetAll(ctx context.Context) ([]*domain.Live, error) {

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
