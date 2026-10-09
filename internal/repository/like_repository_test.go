package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
)

// いいねの表とブースを空にし、クラス展示のブースを 1 つ作って ID を返す
func setupLikeTest(t *testing.T) (*sql.DB, *redis.Client, int) {
	t.Helper()
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/kctfest_test_repository?parseTime=true")
	require.NoError(t, err)
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379", DB: 2})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { db.Close(); rdb.Close() })
	require.NoError(t, rdb.FlushDB(context.Background()).Err())
	require.NoError(t, truncateTables(db, "likes", "like_removals", "booths"))
	res, err := db.Exec(`INSERT INTO booths (name, organizer, detail, x, y, z) VALUES ('お化け屋敷', '3-2', '', 0, 0, 0)`)
	require.NoError(t, err)
	id, _ := res.LastInsertId()
	return db, rdb, int(id)
}

func TestLikeRepository_押す_取り消す_自分の一覧(t *testing.T) {
	db, rdb, boothID := setupLikeTest(t)
	r := repository.NewLikeRepository(db, rdb)
	ctx := context.Background()
	now := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)

	require.NoError(t, r.Like(ctx, boothID, "v1", now))
	require.NoError(t, r.Like(ctx, boothID, "v1", now.Add(time.Minute)), "2 回目もエラーにしない")
	ids, err := r.Mine(ctx, "v1")
	require.NoError(t, err)
	assert.Equal(t, []int{boothID}, ids)

	times, err := r.AllLikeTimes(ctx)
	require.NoError(t, err)
	assert.Equal(t, []time.Time{now}, times[boothID], "1 件だけ、最初の時刻のまま")

	require.NoError(t, r.Unlike(ctx, boothID, "v1"))
	require.NoError(t, r.Unlike(ctx, boothID, "v1"), "押していなくてもエラーにしない")
	ids, err = r.Mine(ctx, "v1")
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestLikeRepository_時間帯で消すと記録が残る(t *testing.T) {
	db, rdb, boothID := setupLikeTest(t)
	r := repository.NewLikeRepository(db, rdb)
	ctx := context.Background()
	base := time.Date(2026, 10, 31, 5, 0, 0, 0, time.UTC)
	for i, v := range []string{"a", "b", "c"} {
		require.NoError(t, r.Like(ctx, boothID, v, base.Add(time.Duration(i)*10*time.Minute)))
	}
	// 5:00 以上 5:20 未満 → a と b
	n, err := r.RemoveRange(ctx, boothID, base, base.Add(20*time.Minute), "admin", base)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	var logged int
	var booth sql.NullInt32
	require.NoError(t, db.QueryRow("SELECT booth_id, removed FROM like_removals").Scan(&booth, &logged))
	assert.Equal(t, 2, logged)
	assert.Equal(t, int32(boothID), booth.Int32)

	n, err = r.RemoveAll(ctx, "admin", base)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	var all int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM like_removals WHERE booth_id IS NULL").Scan(&all))
	assert.Equal(t, 1, all)
}

func TestLikeRepository_切り替えの上限(t *testing.T) {
	db, rdb, _ := setupLikeTest(t)
	r := repository.NewLikeRepository(db, rdb)
	ctx := context.Background()
	for i := 0; i < 60; i++ {
		ok, err := r.AllowToggle(ctx, "v1")
		require.NoError(t, err)
		require.True(t, ok, "%d 回目", i+1)
	}
	ok, err := r.AllowToggle(ctx, "v1")
	require.NoError(t, err)
	assert.False(t, ok, "61 回目は弾く")
	ok, _ = r.AllowToggle(ctx, "v2")
	assert.True(t, ok, "別の番号は別に数える")
}

func TestLikeRepository_番号を作る上限(t *testing.T) {
	db, rdb, _ := setupLikeTest(t)
	r := repository.NewLikeRepository(db, rdb)
	ctx := context.Background()
	for i := 0; i < 3000; i++ {
		ok, err := r.AllowNewVoter(ctx)
		require.NoError(t, err)
		require.True(t, ok, "%d 個目", i+1)
	}
	ok, _ := r.AllowNewVoter(ctx)
	assert.False(t, ok, "3001 個目は弾く")
}
