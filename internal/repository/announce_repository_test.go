package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

func setupAnnouncementRedis(t *testing.T) *redis.Client {
	t.Helper()
	// internal/repository 用の論理DB(他パッケージの FlushDB と干渉しないよう分けている)
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 2})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	require.NoError(t, rdb.FlushDB(context.Background()).Err())
	return rdb
}

func TestAnnouncementRepository(t *testing.T) {
	rdb := setupAnnouncementRedis(t)
	ctx := context.Background()
	repo := NewAnnouncementRepository(rdb)

	t.Run("はじめは空", func(t *testing.T) {
		a, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "", a.Content)
	})

	t.Run("入れたものを読める", func(t *testing.T) {
		require.NoError(t, repo.Update(ctx, &domain.Announcement{Content: "new announcement"}))
		a, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "new announcement", a.Content)
	})

	t.Run("別のサーバー(作り直し・再起動)でも同じお知らせを読める", func(t *testing.T) {
		other := NewAnnouncementRepository(rdb)
		a, err := other.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "new announcement", a.Content)
	})

	t.Run("同時に読み書きしても壊れない", func(t *testing.T) {
		var wg sync.WaitGroup
		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = repo.Update(ctx, &domain.Announcement{Content: "concurrent"})
				_, _ = repo.Get(ctx)
			}()
		}
		wg.Wait()
		a, err := repo.Get(ctx)
		require.NoError(t, err)
		assert.Equal(t, "concurrent", a.Content)
	})
}

func TestAnnouncementRepository_読んだものを10秒だけ覚えておく(t *testing.T) {
	rdb := setupAnnouncementRedis(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	a := newAnnouncementRepository(rdb, clock) // 来場者がよく読むサーバー
	b := newAnnouncementRepository(rdb, clock) // 学生会が書き換えるサーバー

	require.NoError(t, b.Update(ctx, &domain.Announcement{Content: "開会式は 9:25 から"}))
	got, _ := a.Get(ctx)
	require.Equal(t, "開会式は 9:25 から", got.Content)

	require.NoError(t, b.Update(ctx, &domain.Announcement{Content: "開会式は 10 分遅れます"}))
	got, _ = b.Get(ctx)
	assert.Equal(t, "開会式は 10 分遅れます", got.Content, "書き換えたサーバーはすぐ新しいもの")

	now = now.Add(9 * time.Second)
	got, _ = a.Get(ctx)
	assert.Equal(t, "開会式は 9:25 から", got.Content, "ほかのサーバーは 10 秒までは覚えているもの")

	now = now.Add(2 * time.Second)
	got, _ = a.Get(ctx)
	assert.Equal(t, "開会式は 10 分遅れます", got.Content, "10 秒を過ぎたら読み直す")
}
