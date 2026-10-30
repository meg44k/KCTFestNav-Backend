package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 記録を残す長さ。祭りのあとしばらく見返せれば十分
const visitTTL = 90 * 24 * time.Hour

type visitRepository struct{ rdb *redis.Client }

func NewVisitRepository(rdb *redis.Client) domain.VisitRepository {
	return &visitRepository{rdb: rdb}
}

// 日ごとのキー。"visits:{2006-01-02}"
func visitKey(day string) string { return "visits:" + day }

// HyperLogLog に番号を足す(同じ番号は何度足しても 1 人)。2 コマンドを 1 往復で送る
func (r *visitRepository) Record(ctx context.Context, day, visitor string) error {
	_, err := r.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		p.PFAdd(ctx, visitKey(day), visitor)
		p.Expire(ctx, visitKey(day), visitTTL)
		return nil
	})
	return err
}

// 日ごとの数と、それらの日をまとめた重なりなしの数を 1 往復で数える
func (r *visitRepository) Count(ctx context.Context, days []string) (map[string]int64, int64, error) {
	keys := make([]string, len(days))
	for i, d := range days {
		keys[i] = visitKey(d)
	}
	perCmds := make([]*redis.IntCmd, len(days))
	var totalCmd *redis.IntCmd
	_, err := r.rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		for i, k := range keys {
			perCmds[i] = p.PFCount(ctx, k)
		}
		totalCmd = p.PFCount(ctx, keys...)
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	per := make(map[string]int64, len(days))
	for i, d := range days {
		per[d] = perCmds[i].Val()
	}
	return per, totalCmd.Val(), nil
}
