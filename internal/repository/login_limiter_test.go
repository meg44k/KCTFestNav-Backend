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

func TestLoginLimiter_同じIDで10回失敗したら止める(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 9; i++ {
		require.NoError(t, l.Failed(ctx, "booth-1"))
	}
	blocked, err := l.Blocked(ctx, "booth-1")
	require.NoError(t, err)
	assert.False(t, blocked, "9 回まではまだ試せる")

	require.NoError(t, l.Failed(ctx, "booth-1"))
	blocked, _ = l.Blocked(ctx, "booth-1")
	assert.True(t, blocked, "10 回失敗したら止める")
	blocked, _ = l.Blocked(ctx, "booth-2")
	assert.False(t, blocked, "ほかの ID は別に数える")
}

func TestLoginLimiter_ログインできたら数を消す(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 9; i++ {
		require.NoError(t, l.Failed(ctx, "booth-1"))
	}
	require.NoError(t, l.Succeeded(ctx, "booth-1"))
	require.NoError(t, l.Failed(ctx, "booth-1"))
	blocked, _ := l.Blocked(ctx, "booth-1")
	assert.False(t, blocked, "消したので 1 回目から数え直す")
}

func TestLoginLimiter_全体で500回失敗したらどのIDも止める(t *testing.T) {
	l := repository.NewLoginLimiter(setupLoginLimiter(t))
	ctx := context.Background()
	for i := 0; i < 500; i++ {
		require.NoError(t, l.Failed(ctx, fmt.Sprintf("guess-%d", i)))
	}
	blocked, err := l.Blocked(ctx, "booth-1")
	require.NoError(t, err)
	assert.True(t, blocked)
}
