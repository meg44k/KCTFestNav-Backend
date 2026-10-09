package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/database"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 回数の上限(1 分あたり)
const (
	newVotersPerMinute = 300 // サーバー全体で新しく作る投票者番号
	togglesPerMinute   = 60  // 1 つの番号でのいいねの切り替え
)

type likeRepository struct {
	sqlDB *sql.DB
	db    *database.Queries
	rdb   *redis.Client
	now   func() time.Time
}

func NewLikeRepository(db *sql.DB, rdb *redis.Client) domain.LikeRepository {
	return &likeRepository{sqlDB: db, db: database.New(db), rdb: rdb, now: time.Now}
}

func (r *likeRepository) Like(ctx context.Context, boothID int, voterID string, at time.Time) error {
	// 押してあれば何もしない(主キーで 1 件に保つ)
	return r.db.InsertLike(ctx, database.InsertLikeParams{BoothID: int32(boothID), VoterID: voterID, CreatedAt: at.UTC()})
}

func (r *likeRepository) Unlike(ctx context.Context, boothID int, voterID string) error {
	return r.db.DeleteLike(ctx, database.DeleteLikeParams{BoothID: int32(boothID), VoterID: voterID})
}

func (r *likeRepository) Mine(ctx context.Context, voterID string) ([]int, error) {
	rows, err := r.db.ListLikedBoothIDs(ctx, voterID)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(rows))
	for i, id := range rows {
		ids[i] = int(id)
	}
	return ids, nil
}

func (r *likeRepository) AllLikeTimes(ctx context.Context) (map[int][]time.Time, error) {
	rows, err := r.db.ListLikeTimes(ctx)
	if err != nil {
		return nil, err
	}
	out := map[int][]time.Time{}
	for _, row := range rows {
		out[int(row.BoothID)] = append(out[int(row.BoothID)], row.CreatedAt.UTC())
	}
	return out, nil
}

func (r *likeRepository) RemoveRange(ctx context.Context, boothID int, from, to time.Time, by string, at time.Time) (int, error) {
	return r.remove(ctx, func(q *database.Queries) (int64, error) {
		return q.DeleteLikesInRange(ctx, database.DeleteLikesInRangeParams{BoothID: int32(boothID), CreatedAt: from.UTC(), CreatedAt_2: to.UTC()})
	}, database.InsertLikeRemovalParams{
		BoothID: sql.NullInt32{Int32: int32(boothID), Valid: true},
		FromAt:  sql.NullTime{Time: from.UTC(), Valid: true},
		ToAt:    sql.NullTime{Time: to.UTC(), Valid: true},
	}, by, at)
}

func (r *likeRepository) RemoveAll(ctx context.Context, by string, at time.Time) (int, error) {
	return r.remove(ctx, func(q *database.Queries) (int64, error) {
		return q.DeleteAllLikes(ctx)
	}, database.InsertLikeRemovalParams{}, by, at)
}

// 消して、その記録を同じトランザクションで残す
func (r *likeRepository) remove(ctx context.Context, del func(*database.Queries) (int64, error), log database.InsertLikeRemovalParams, by string, at time.Time) (int, error) {
	tx, err := r.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	q := r.db.WithTx(tx)
	n, err := del(q)
	if err != nil {
		return 0, err
	}
	log.Removed = int32(n)
	log.RemovedBy = by
	log.CreatedAt = at.UTC()
	if err := q.InsertLikeRemoval(ctx, log); err != nil {
		return 0, err
	}
	return int(n), tx.Commit()
}

// 1 分ごとのキーで数える。超えていなければ true
func (r *likeRepository) allow(ctx context.Context, key string, limit int64) (bool, error) {
	n, err := r.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		r.rdb.Expire(ctx, key, 2*time.Minute)
	}
	return n <= limit, nil
}

func (r *likeRepository) minute() string { return r.now().UTC().Format("200601021504") }

func (r *likeRepository) AllowNewVoter(ctx context.Context) (bool, error) {
	return r.allow(ctx, "likes:new-voters:"+r.minute(), newVotersPerMinute)
}

func (r *likeRepository) AllowToggle(ctx context.Context, voterID string) (bool, error) {
	return r.allow(ctx, "likes:toggle:"+voterID+":"+r.minute(), togglesPerMinute)
}
