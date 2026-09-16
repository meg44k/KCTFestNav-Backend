package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

const NoExpiration time.Duration = 0

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
		Name:      b.Name,
		Organizer: b.Organizer,
		Detail:    b.Detail,
		Location:  toNullString(b.Location),
		ImageUrl:  toNullString(b.ImageURL),
		X:         float64(b.X),
		Y:         float64(b.Y),
		Z:         float64(b.Z),
		Latitude:  toNullFloat64(b.Latitude),
		Longitude: toNullFloat64(b.Longitude),
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

	// Redisから混雑状況を取ってくる
	rawStatus, err := br.cache.Get(ctx, formatRedisCongestionStatusKey(id)).Int()
	var congestionStatus domain.CongestionStatus
	if err != nil {
		if errors.Is(err, redis.Nil) { // congestionStatusが登録されていなかった時、空きにする
			congestionStatus = domain.BoothCongestionEmpty
		} else {
			return nil, err
		}
	} else {
		congestionStatus = domain.CongestionStatus(rawStatus)
	}

	booth, err := domain.ReconstructBooth(
		int(dbBooth.ID),
		congestionStatus,
		domain.BoothParams{
			Name:      dbBooth.Name,
			Organizer: dbBooth.Organizer,
			Detail:    dbBooth.Detail,
			Location:  dbBooth.Location.String,
			ImageURL:  dbBooth.ImageUrl.String,
			X:         float32(dbBooth.X),
			Y:         float32(dbBooth.Y),
			Z:         float32(dbBooth.Z),
			Latitude:  dbBooth.Latitude.Float64,
			Longitude: dbBooth.Longitude.Float64,
		},
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
		// Redisから混雑状況を取ってくる
		rawStatus, err := br.cache.Get(ctx, formatRedisCongestionStatusKey(int(b.ID))).Int()
		var congestionStatus domain.CongestionStatus
		if err != nil {
			if errors.Is(err, redis.Nil) { // congestionStatusが登録されていなかった時、空きにする
				congestionStatus = domain.BoothCongestionEmpty
			} else {
				return nil, err
			}
		} else {
			congestionStatus = domain.CongestionStatus(rawStatus)
		}

		booth, err := domain.ReconstructBooth(
			int(b.ID),
			congestionStatus,
			domain.BoothParams{
				Name:      b.Name,
				Organizer: b.Organizer,
				Detail:    b.Detail,
				Location:  b.Location.String,
				ImageURL:  b.ImageUrl.String,
				X:         float32(b.X),
				Y:         float32(b.Y),
				Z:         float32(b.Z),
				Latitude:  b.Latitude.Float64,
				Longitude: b.Longitude.Float64,
			},
		)
		if err != nil {
			return nil, err
		}
		booths[i] = booth
	}
	return booths, nil

}

func (br *boothRepository) Update(ctx context.Context, b *domain.Booth) error {
	arg := database.UpdateBoothParams{
		Name:      b.Name,
		Organizer: b.Organizer,
		Detail:    b.Detail,
		Location:  toNullString(b.Location),
		ImageUrl:  toNullString(b.ImageURL),
		X:         float64(b.X),
		Y:         float64(b.Y),
		Z:         float64(b.Z),
		Latitude:  toNullFloat64(b.Latitude),
		Longitude: toNullFloat64(b.Longitude),
		ID:        int32(b.ID),
	}

	if err := br.db.UpdateBooth(ctx, arg); err != nil {
		return err
	}
	return nil
}

func (br *boothRepository) UpdateCongestion(ctx context.Context, id int, congestionStatus domain.CongestionStatus) error {
	if err := br.cache.Set(ctx, formatRedisCongestionStatusKey(id), int(congestionStatus), NoExpiration).Err(); err != nil {
		return err
	}
	return nil
}

func (br *boothRepository) Delete(ctx context.Context, id int) error {
	if err := br.db.DeleteBooth(ctx, int32(id)); err != nil {
		return err
	}
	br.cache.Del(ctx, formatRedisCongestionStatusKey(id))
	return nil
}

// 空文字を NULL として保存するためのヘルパー関数。
// 読み出し側は NullString.String をそのまま使うため、NULL は空文字に戻る
func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// 座標未設定(ゼロ値)を NULL として保存するためのヘルパー関数。
// 高専祭の会場は緯度経度ともに 0 になり得ないため、0 を未設定として扱う
func toNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0}
}

// Redisの混雑度のキーをフォーマットするヘルパー関数。
// "congestion_status:{id}"がstring型で返される
func formatRedisCongestionStatusKey(id int) string {
	return fmt.Sprintf("congestion_status:%d", id)
}
