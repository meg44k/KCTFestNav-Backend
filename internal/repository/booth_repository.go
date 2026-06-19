package repository

import (
	"context"
	"database/sql"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

type boothRepository struct {
	db    *database.Queries
	cache *redis.Client
}

func NewBoothRepository(db *sql.DB, rdb *redis.Client) domain.BoothRepository {
	return &boothRepository{
		db:    database.New(db),
		cache: rdb,
	}
}

func (br *boothRepository) Create(ctx context.Context, b *domain.Booth) error {
	arg := database.CreateBoothParams{
		Name:             b.Name,
		Organizer:        b.Organizer,
		Detail:           b.Detail,
		CongestionStatus: b.CongestionStatus(),
		X:                float64(b.X),
		Y:                float64(b.Y),
		Z:                float64(b.Z),
	}
	err := br.db.CreateBooth(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (br *boothRepository) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	dbBooth, err := br.db.GetBoothByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	booth, err := domain.ReconstructBooth(
		int(dbBooth.ID),
		dbBooth.Name,
		dbBooth.Organizer,
		dbBooth.Detail,
		dbBooth.CongestionStatus,
		float32(dbBooth.X),
		float32(dbBooth.Y),
		float32(dbBooth.Z),
	)
	if err != nil {
		return nil, err
	}
	return booth, nil
}

func (br *boothRepository) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	dbBooths, err := br.db.GetAllBooths(ctx)
	if err != nil {
		return nil, err
	}
	booths := make([]*domain.Booth, len(dbBooths))
	for i, b := range dbBooths {
		booth, err := domain.ReconstructBooth(
			int(b.ID),
			b.Name,
			b.Organizer,
			b.Detail,
			b.CongestionStatus,
			float32(b.X),
			float32(b.Y),
			float32(b.Z),
		)
		if err != nil {
			return nil, err
		}
		booths[i] = booth
	}
	return booths, nil

}

func (br *boothRepository) Update(ctx context.Context, b *domain.Booth) error {
	// TODO: 実装を行う
	return nil
}

func (br *boothRepository) UpdateCongestion(ctx context.Context, id int, congestionLevel int8) error {
	// TODO: 実装を行う
	return nil
}

func (br *boothRepository) Delete(ctx context.Context, id int) error {
	if err := br.db.DeleteBooth(ctx, int32(id)); err != nil {
		return err
	}
	return nil
}
