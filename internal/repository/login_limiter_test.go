package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
)

func setupLoginLimiter(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 2})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	require.NoError(t, rdb.FlushDB(context.Background()).Err())
	return rdb
}

const ip1, ip2 = "203.0.113.1", "198.51.100.2"

func TestLoginLimiter_同じ接続元と同じIDは5回まで(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for want := 4; want >= 1; want-- {
		left, err := l.Failed(ctx, ip1, "booth-1")
		require.NoError(t, err)
		assert.Equal(t, want, left, "あと何回")
		blocked, _ := l.Blocked(ctx, ip1, "booth-1")
		assert.False(t, blocked)
	}
	left, _ := l.Failed(ctx, ip1, "booth-1")
	assert.Equal(t, 0, left)
	blocked, _ := l.Blocked(ctx, ip1, "booth-1")
	assert.True(t, blocked, "5 回失敗したら止める")
}

func TestLoginLimiter_別の接続元の本人は止めない(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, _ = l.Failed(ctx, ip1, "booth-1")
	}
	blocked, err := l.Blocked(ctx, ip2, "booth-1")
	require.NoError(t, err)
	assert.False(t, blocked, "わざと失敗を重ねても、ほかの場所からの本人は締め出されない")
}

func TestLoginLimiter_同じ接続元でもほかのIDは止めない(t *testing.T) {
	// 学内 Wi-Fi のように係の全員が同じ接続元でも、誰かの失敗やいたずらで全員が止まらないように
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_, _ = l.Failed(ctx, ip1, fmt.Sprintf("guess-%d", i))
	}
	blocked, err := l.Blocked(ctx, ip1, "booth-1")
	require.NoError(t, err)
	assert.False(t, blocked)
	left, _ := l.Failed(ctx, ip1, "booth-1")
	assert.Equal(t, 4, left, "その ID の残りだけを数える")
}

func TestLoginLimiter_ログインできたらそのIDの数を消す(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		_, _ = l.Failed(ctx, ip1, "booth-1")
	}
	require.NoError(t, l.Succeeded(ctx, ip1, "booth-1"))
	left, _ := l.Failed(ctx, ip1, "booth-1")
	assert.Equal(t, 4, left, "1 回目から数え直す")
}
