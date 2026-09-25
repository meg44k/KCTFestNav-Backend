package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

var currentLiveKey string = "lives:current"

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

type currentLiveDTO struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	Detail        string            `json:"detail"`
	ThumbnailURL  string            `json:"thumbnail_url"`
	StartTime     time.Time         `json:"start_time"`
	EndTime       time.Time         `json:"end_time"`
	SessionNumber int8              `json:"session_number"`
	Status        domain.LiveStatus `json:"status"`
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
		domain.LiveStatus(dbLive.Status),
		domain.LiveParams{
			Name:          dbLive.Name,
			Detail:        dbLive.Detail.String,
			ThumbnailURL:  dbLive.Thumbnailurl.String,
			StartTime:     dbLive.StartTime,
			EndTime:       dbLive.EndTime,
			SessionNumber: int8(dbLive.SessionNumber.Int16),
		},
	)
	return live, nil
}

func (lr *liveRepository) GetAll(ctx context.Context) ([]*domain.Live, error) {
	dbLives, err := lr.db.GetAllLives(ctx)
	if err != nil {
		return nil, err
	}

	lives := make([]*domain.Live, len(dbLives))
	for i, dbLive := range dbLives {
		live, err := domain.ReconstructLive(
			int(dbLive.ID),
			domain.LiveStatus(dbLive.Status),
			domain.LiveParams{
				Name:          dbLive.Name,
				Detail:        dbLive.Detail.String,
				ThumbnailURL:  dbLive.Thumbnailurl.String,
				StartTime:     dbLive.StartTime,
				EndTime:       dbLive.EndTime,
				SessionNumber: int8(dbLive.SessionNumber.Int16),
			},
		)
		if err != nil {
			return nil, err
		}
		lives[i] = live
	}
	return lives, nil

}

func (lr *liveRepository) GetCurrentLive(ctx context.Context) (*domain.Live, error) {
	var liveDTO currentLiveDTO
	val, err := lr.cache.Get(ctx, currentLiveKey).Result()
	if errors.Is(err, redis.Nil) { // キャッシュヒットしなかった時
		dbLive, err := lr.db.GetCurrentLive(ctx) // MySQLから開催中のライブを取ってくる(1つ)
		if err != nil {
			return nil, err
		}
		liveDTO = currentLiveDTO{
			int(dbLive.ID),
			dbLive.Name,
			dbLive.Detail.String,
			dbLive.Thumbnailurl.String,
			dbLive.StartTime,
			dbLive.EndTime,
			int8(dbLive.SessionNumber.Int16),
			domain.LiveStatus(dbLive.Status),
		}
		jsonData, err := json.Marshal(liveDTO)
		if err != nil {
			return nil, err
		}
		if err := lr.cache.Set(ctx, currentLiveKey, jsonData, NoExpiration).Err(); err != nil {
			return nil, err
		} // キャッシュに登録する
	} else if err != nil {
		return nil, err
	} else { // キャッシュヒットした時
		// Redis型をリポジトリ層内に隠蔽するために、DTOに一回展開。
		if err := json.Unmarshal([]byte(val), &liveDTO); err != nil {
			return nil, err
		}
	}
	// live型へ
	live, err := domain.ReconstructLive(
		liveDTO.ID,
		liveDTO.Status,
		domain.LiveParams{
			Name:          liveDTO.Name,
			Detail:        liveDTO.Detail,
			ThumbnailURL:  liveDTO.ThumbnailURL,
			StartTime:     liveDTO.StartTime,
			EndTime:       liveDTO.EndTime,
			SessionNumber: liveDTO.SessionNumber,
		},
	)
	if err != nil {
		return nil, err
	}

	return live, nil
}

func (lr *liveRepository) Update(ctx context.Context, l *domain.Live) error {
	arg := database.UpdateLiveParams{
		Name: l.Name,
		Detail: sql.NullString{
			String: l.Detail,
			Valid:  l.Detail != "",
		},
		Thumbnailurl: sql.NullString{
			String: l.ThumbnailURL,
			Valid:  l.ThumbnailURL != "",
		},
		StartTime: l.StartTime,
		EndTime:   l.EndTime,
		SessionNumber: sql.NullInt16{
			Int16: int16(l.SessionNumber()),
			Valid: true,
		},
		Status: int8(l.Status()),
		ID:     int32(l.ID),
	}
	if err := lr.db.UpdateLive(ctx, arg); err != nil {
		return err
	}
	return lr.cache.Del(ctx, currentLiveKey).Err()
}

func (lr *liveRepository) UpdateLiveStatus(ctx context.Context, id int, status domain.LiveStatus) error {
	// mysqlのliveステータスを変更する
	arg := database.UpdateLiveStatusParams{
		Status: int8(status),
		ID:     int32(id),
	}
	if err := lr.db.UpdateLiveStatus(ctx, arg); err != nil {
		return err
	}

	// redisの現在のライブ情報を削除する
	return lr.cache.Del(ctx, currentLiveKey).Err()
}

func (lr *liveRepository) Delete(ctx context.Context, id int) error {
	if err := lr.db.DeleteLive(ctx, int32(id)); err != nil {
		return err
	}
	// キャッシュ削除
	return lr.cache.Del(ctx, currentLiveKey).Err()
}
