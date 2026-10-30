package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 訪問者のリポジトリの作りもの。日ごとに番号の集合を持つ
type fakeVisitRepo struct{ days map[string]map[string]bool }

func (f *fakeVisitRepo) Record(_ context.Context, day, visitor string) error {
	if f.days[day] == nil {
		f.days[day] = map[string]bool{}
	}
	f.days[day][visitor] = true
	return nil
}

func (f *fakeVisitRepo) Count(_ context.Context, days []string) (map[string]int64, int64, error) {
	per := map[string]int64{}
	all := map[string]bool{}
	for _, d := range days {
		per[d] = int64(len(f.days[d]))
		for v := range f.days[d] {
			all[v] = true
		}
	}
	return per, int64(len(all)), nil
}

const visitorA = "0b7f2c1e-3a4d-4c5e-8f60-7182a3b4c5d6"
const visitorB = "9a8b7c6d-5e4f-4a3b-9c2d-1e0f2a3b4c5d"

func TestVisitUsecase_日付は日本時間で決める(t *testing.T) {
	repo := &fakeVisitRepo{days: map[string]map[string]bool{}}
	// 10/30 23:30 JST = 10/30 14:30 UTC、10/31 00:30 JST = 10/30 15:30 UTC
	now := time.Date(2026, 10, 30, 15, 30, 0, 0, time.UTC)
	uc := NewVisitUsecase(repo, func() time.Time { return now })

	require.NoError(t, uc.Record(context.Background(), visitorA))
	assert.True(t, repo.days["2026-10-31"][visitorA], "日本時間では 10/31")
}

func TestVisitUsecase_番号の形がおかしければ断る(t *testing.T) {
	repo := &fakeVisitRepo{days: map[string]map[string]bool{}}
	uc := NewVisitUsecase(repo, time.Now)
	for _, v := range []string{"", "abc", "0b7f2c1e3a4d4c5e8f607182a3b4c5d6x", visitorA + "x"} {
		assert.ErrorIs(t, uc.Record(context.Background(), v), domain.ErrInvalidVisitor, v)
	}
	assert.Empty(t, repo.days)
}

func TestVisitUsecase_直近の日ごとの数と合計(t *testing.T) {
	repo := &fakeVisitRepo{days: map[string]map[string]bool{}}
	day1 := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC) // 10:00 JST
	now := day1
	uc := NewVisitUsecase(repo, func() time.Time { return now })
	ctx := context.Background()
	require.NoError(t, uc.Record(ctx, visitorA))
	require.NoError(t, uc.Record(ctx, visitorA)) // 同じ日にもう一度来ても 1 人
	require.NoError(t, uc.Record(ctx, visitorB))
	now = day1.Add(24 * time.Hour)
	require.NoError(t, uc.Record(ctx, visitorA)) // 次の日にも来た

	got, err := uc.Count(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, []domain.VisitDay{
		{Day: "2026-10-30", Visitors: 0},
		{Day: "2026-10-31", Visitors: 2},
		{Day: "2026-11-01", Visitors: 1},
	}, got.Days, "古い日から順")
	assert.Equal(t, int64(2), got.Total, "2 日とも来た人は合計では 1 人")
}

func TestVisitUsecase_日数は1から60まで(t *testing.T) {
	uc := NewVisitUsecase(&fakeVisitRepo{days: map[string]map[string]bool{}}, time.Now)
	for _, n := range []int{0, -1, 61} {
		_, err := uc.Count(context.Background(), n)
		assert.ErrorIs(t, err, domain.ErrInvalidRange, n)
	}
}
