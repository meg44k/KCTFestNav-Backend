package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
)

func TestVisitRepository_日ごとと合計を重なりなしで数える(t *testing.T) {
	rdb := setupLoginLimiter(t) // 同じテスト用 Redis(DB 2、毎回空にする)
	r := repository.NewVisitRepository(rdb)
	ctx := context.Background()

	require.NoError(t, r.Record(ctx, "2026-10-31", "a"))
	require.NoError(t, r.Record(ctx, "2026-10-31", "a")) // 同じ日の同じ番号は 1 人
	require.NoError(t, r.Record(ctx, "2026-10-31", "b"))
	require.NoError(t, r.Record(ctx, "2026-11-01", "a")) // 次の日にも来た

	per, total, err := r.Count(ctx, []string{"2026-10-30", "2026-10-31", "2026-11-01"})
	require.NoError(t, err)
	assert.Equal(t, map[string]int64{"2026-10-30": 0, "2026-10-31": 2, "2026-11-01": 1}, per)
	assert.Equal(t, int64(2), total, "2 日とも来た a は合計では 1 人")
}

func TestVisitRepository_記録は90日で消える(t *testing.T) {
	rdb := setupLoginLimiter(t)
	r := repository.NewVisitRepository(rdb)
	require.NoError(t, r.Record(context.Background(), "2026-10-31", "a"))

	ttl, err := rdb.TTL(context.Background(), "visits:2026-10-31").Result()
	require.NoError(t, err)
	assert.InDelta(t, (90 * 24 * time.Hour).Seconds(), ttl.Seconds(), 5)
}
