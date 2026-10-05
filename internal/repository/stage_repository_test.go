package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 実際の MySQL(kctfest_test_repository)に繋ぎ、ステージの 3 テーブルを空にして返す
func setupStageRepo(t *testing.T) (domain.StageRepository, *sql.DB) {
	t.Helper()
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/kctfest_test_repository?parseTime=true")
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	// 外部キーがあるので子から消す
	for _, table := range []string{"performers", "stage_blocks", "stage_sections"} {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			t.Fatalf("%s を空にできない: %v", table, err)
		}
	}
	return NewStageRepository(db), db
}

var jst = time.FixedZone("JST", 9*3600)

// 部品を作るヘルパー
func mustSection(t *testing.T, r domain.StageRepository, name string, order int) int {
	t.Helper()
	s, err := domain.NewStageSection(name, "第一体育館", order)
	require.NoError(t, err)
	id, err := r.CreateSection(context.Background(), s)
	require.NoError(t, err)
	return id
}

func mustBlock(t *testing.T, r domain.StageRepository, sectionID int, start time.Time) int {
	t.Helper()
	b, err := domain.NewStageBlock(sectionID, start, start.Add(40*time.Minute))
	require.NoError(t, err)
	id, err := r.CreateBlock(context.Background(), b)
	require.NoError(t, err)
	return id
}

func mustPerformer(t *testing.T, r domain.StageRepository, blockID int, name string) int {
	t.Helper()
	p, err := domain.NewPerformer(blockID, name, name+"の紹介", "")
	require.NoError(t, err)
	id, err := r.CreatePerformer(context.Background(), p)
	require.NoError(t, err)
	return id
}

func names(ps []*domain.Performer) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Name
	}
	return out
}

func TestStageRepository_GetSchedule(t *testing.T) {
	r, _ := setupStageRepo(t)
	ctx := context.Background()
	day := time.Date(2026, 10, 31, 13, 0, 0, 0, jst)

	// 作った順と並び順を変えておく
	iyashi := mustSection(t, r, "癒し系", 2)
	live1 := mustSection(t, r, "Live1", 1)
	late := mustBlock(t, r, iyashi, day.Add(110*time.Minute))
	early := mustBlock(t, r, iyashi, day.Add(60*time.Minute))
	lb := mustBlock(t, r, live1, day)
	mustPerformer(t, r, lb, "バンドA")
	mustPerformer(t, r, lb, "バンドB")
	mustPerformer(t, r, early, "奏者X")
	_ = late

	got, err := r.GetSchedule(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Live1", got[0].Name)
	assert.Equal(t, "第一体育館", got[0].Location)
	require.Len(t, got[0].Blocks, 1)
	assert.Equal(t, []string{"バンドA", "バンドB"}, names(got[0].Blocks[0].Performers))
	assert.Equal(t, []int{1, 2}, []int{got[0].Blocks[0].Performers[0].PerformOrder, got[0].Blocks[0].Performers[1].PerformOrder})
	assert.True(t, got[0].Blocks[0].StartTime.Equal(day), "JSTで保存した時刻が同じ瞬間で戻る: %v", got[0].Blocks[0].StartTime)

	assert.Equal(t, "癒し系", got[1].Name)
	require.Len(t, got[1].Blocks, 2)
	assert.Equal(t, early, got[1].Blocks[0].ID, "ブロックは開始時刻の順")
	assert.Equal(t, late, got[1].Blocks[1].ID)
	assert.Equal(t, []string{"奏者X"}, names(got[1].Blocks[0].Performers))
	assert.Empty(t, got[1].Blocks[1].Performers)
	assert.NotNil(t, got[1].Blocks[1].Performers, "出演者0人でも nil ではなく空")
}

func TestStageRepository_UpdateAndDelete(t *testing.T) {
	r, db := setupStageRepo(t)
	ctx := context.Background()
	start := time.Date(2026, 10, 31, 13, 0, 0, 0, jst)
	sid := mustSection(t, r, "Live1", 1)
	bid := mustBlock(t, r, sid, start)
	pid := mustPerformer(t, r, bid, "バンドA")

	t.Run("セクションを更新できる。値が同じでも成功する", func(t *testing.T) {
		s := &domain.StageSection{ID: sid, Name: "Live 1", Location: "中庭", SortOrder: 5}
		require.NoError(t, r.UpdateSection(ctx, s))
		require.NoError(t, r.UpdateSection(ctx, s))
		got, err := r.GetSection(ctx, sid)
		require.NoError(t, err)
		assert.Equal(t, "Live 1", got.Name)
		assert.Equal(t, "中庭", got.Location)
		assert.Equal(t, 5, got.SortOrder)
	})

	t.Run("ブロックを更新しても current_order は変わらない", func(t *testing.T) {
		require.NoError(t, r.AdvanceBlock(ctx, bid))
		b := &domain.StageBlock{ID: bid, StartTime: start.Add(time.Hour), EndTime: start.Add(2 * time.Hour)}
		require.NoError(t, r.UpdateBlock(ctx, b))
		got, err := r.GetBlock(ctx, bid)
		require.NoError(t, err)
		assert.True(t, got.StartTime.Equal(start.Add(time.Hour)))
		assert.Equal(t, 1, got.CurrentOrder)
		assert.Equal(t, []string{"バンドA"}, names(got.Performers))
	})

	t.Run("出演者を更新しても出演順は変わらない", func(t *testing.T) {
		require.NoError(t, r.UpdatePerformer(ctx, &domain.Performer{ID: pid, Name: "バンドA'", Detail: "新", ThumbnailURL: "https://example.com/a.jpg"}))
		got, err := r.GetPerformer(ctx, pid)
		require.NoError(t, err)
		assert.Equal(t, "バンドA'", got.Name)
		assert.Equal(t, "https://example.com/a.jpg", got.ThumbnailURL)
		assert.Equal(t, 1, got.PerformOrder)
		assert.Equal(t, bid, got.BlockID)
	})

	t.Run("存在しない id は sql.ErrNoRows", func(t *testing.T) {
		assert.ErrorIs(t, r.UpdateSection(ctx, &domain.StageSection{ID: 999999, Name: "x"}), sql.ErrNoRows)
		assert.ErrorIs(t, r.UpdateBlock(ctx, &domain.StageBlock{ID: 999999, StartTime: start, EndTime: start.Add(time.Hour)}), sql.ErrNoRows)
		assert.ErrorIs(t, r.UpdatePerformer(ctx, &domain.Performer{ID: 999999, Name: "x"}), sql.ErrNoRows)
		assert.ErrorIs(t, r.DeleteSection(ctx, 999999), sql.ErrNoRows)
		assert.ErrorIs(t, r.DeleteBlock(ctx, 999999), sql.ErrNoRows)
		assert.ErrorIs(t, r.DeletePerformer(ctx, 999999), sql.ErrNoRows)
		assert.ErrorIs(t, r.AdvanceBlock(ctx, 999999), sql.ErrNoRows)
		assert.ErrorIs(t, r.RewindBlock(ctx, 999999), sql.ErrNoRows)
		assert.ErrorIs(t, r.MovePerformer(ctx, 999999, domain.MoveUp), sql.ErrNoRows)
		_, err := r.GetBlock(ctx, 999999)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		_, err = r.CreateBlock(ctx, &domain.StageBlock{SectionID: 999999, StartTime: start, EndTime: start.Add(time.Hour)})
		assert.ErrorIs(t, err, sql.ErrNoRows, "親のセクションが無い")
		_, err = r.CreatePerformer(ctx, &domain.Performer{BlockID: 999999, Name: "x"})
		assert.ErrorIs(t, err, sql.ErrNoRows, "親のブロックが無い")
	})

	t.Run("セクションを消すと中のブロックと出演者も消える", func(t *testing.T) {
		require.NoError(t, r.DeleteSection(ctx, sid))
		var n int
		require.NoError(t, db.QueryRow("SELECT (SELECT COUNT(*) FROM stage_blocks) + (SELECT COUNT(*) FROM performers)").Scan(&n))
		assert.Equal(t, 0, n)
	})
}

func TestStageRepository_MovePerformer(t *testing.T) {
	r, _ := setupStageRepo(t)
	ctx := context.Background()
	bid := mustBlock(t, r, mustSection(t, r, "Live1", 1), time.Date(2026, 10, 31, 13, 0, 0, 0, jst))
	a := mustPerformer(t, r, bid, "A")
	mustPerformer(t, r, bid, "B")
	c := mustPerformer(t, r, bid, "C")
	order := func() []string {
		b, err := r.GetBlock(ctx, bid)
		require.NoError(t, err)
		return names(b.Performers)
	}

	require.NoError(t, r.MovePerformer(ctx, c, domain.MoveUp))
	assert.Equal(t, []string{"A", "C", "B"}, order())
	require.NoError(t, r.MovePerformer(ctx, a, domain.MoveDown))
	assert.Equal(t, []string{"C", "A", "B"}, order())
	// 端では何もしない
	require.NoError(t, r.MovePerformer(ctx, c, domain.MoveUp))
	assert.Equal(t, []string{"C", "A", "B"}, order())

	// 途中を消して番号が飛んでも、隣と入れ替わる
	require.NoError(t, r.DeletePerformer(ctx, a))
	require.NoError(t, r.MovePerformer(ctx, c, domain.MoveDown))
	assert.Equal(t, []string{"B", "C"}, order())
	// 追加は最後に付く
	mustPerformer(t, r, bid, "D")
	assert.Equal(t, []string{"B", "C", "D"}, order())
}

func TestStageRepository_AdvanceRewind(t *testing.T) {
	r, _ := setupStageRepo(t)
	ctx := context.Background()
	sid := mustSection(t, r, "Live1", 1)
	start := time.Date(2026, 10, 31, 13, 0, 0, 0, jst)
	bid := mustBlock(t, r, sid, start)
	for _, n := range []string{"A", "B"} {
		mustPerformer(t, r, bid, n)
	}
	current := func(id int) int {
		b, err := r.GetBlock(ctx, id)
		require.NoError(t, err)
		return b.CurrentOrder
	}

	t.Run("戻すのは 0 で止まる", func(t *testing.T) {
		require.NoError(t, r.RewindBlock(ctx, bid))
		assert.Equal(t, 0, current(bid))
	})
	t.Run("進めるのは出演者数 + 1(終了)で止まる", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			require.NoError(t, r.AdvanceBlock(ctx, bid))
		}
		assert.Equal(t, 3, current(bid))
		require.NoError(t, r.RewindBlock(ctx, bid))
		assert.Equal(t, 2, current(bid))
	})
	t.Run("出演者 0 人なら 1(終了)で止まる", func(t *testing.T) {
		empty := mustBlock(t, r, sid, start.Add(time.Hour))
		require.NoError(t, r.AdvanceBlock(ctx, empty))
		require.NoError(t, r.AdvanceBlock(ctx, empty))
		assert.Equal(t, 1, current(empty))
	})
	t.Run("同時に押されても上限を超えない", func(t *testing.T) {
		b := mustBlock(t, r, sid, start.Add(2*time.Hour))
		mustPerformer(t, r, b, "X")
		errs := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func() { errs <- r.AdvanceBlock(ctx, b) }()
		}
		for i := 0; i < 10; i++ {
			if err := <-errs; err != nil && !errors.Is(err, context.Canceled) {
				t.Fatalf("AdvanceBlock: %v", err)
			}
		}
		assert.Equal(t, 2, current(b))
	})
}
