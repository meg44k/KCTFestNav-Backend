package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// いいねのリポジトリの作りもの。押されたものを map で持つ
type fakeLikeRepo struct {
	likes       map[int]map[string]time.Time
	newVoterOK  bool
	toggleOK    bool
	removedFrom time.Time
	removedTo   time.Time
	removedBy   string
}

func newFakeLikeRepo() *fakeLikeRepo {
	return &fakeLikeRepo{likes: map[int]map[string]time.Time{}, newVoterOK: true, toggleOK: true}
}

func (f *fakeLikeRepo) Like(_ context.Context, b int, v string, at time.Time) error {
	if f.likes[b] == nil {
		f.likes[b] = map[string]time.Time{}
	}
	if _, ok := f.likes[b][v]; !ok {
		f.likes[b][v] = at
	}
	return nil
}
func (f *fakeLikeRepo) Unlike(_ context.Context, b int, v string) error {
	delete(f.likes[b], v)
	return nil
}
func (f *fakeLikeRepo) Mine(_ context.Context, v string) ([]int, error) {
	ids := []int{}
	for b, m := range f.likes {
		if _, ok := m[v]; ok {
			ids = append(ids, b)
		}
	}
	return ids, nil
}
func (f *fakeLikeRepo) AllLikeTimes(context.Context) (map[int][]time.Time, error) {
	out := map[int][]time.Time{}
	for b, m := range f.likes {
		for _, t := range m {
			out[b] = append(out[b], t)
		}
	}
	return out, nil
}
func (f *fakeLikeRepo) RemoveRange(_ context.Context, b int, from, to time.Time, by string, _ time.Time) (int, error) {
	f.removedFrom, f.removedTo, f.removedBy = from, to, by
	n := 0
	for v, t := range f.likes[b] {
		if !t.Before(from) && t.Before(to) {
			delete(f.likes[b], v)
			n++
		}
	}
	return n, nil
}
func (f *fakeLikeRepo) RemoveAll(_ context.Context, by string, _ time.Time) (int, error) {
	f.removedBy = by
	n := 0
	for _, m := range f.likes {
		n += len(m)
	}
	f.likes = map[int]map[string]time.Time{}
	return n, nil
}
func (f *fakeLikeRepo) AllowNewVoter(context.Context) (bool, error)       { return f.newVoterOK, nil }
func (f *fakeLikeRepo) AllowToggle(context.Context, string) (bool, error) { return f.toggleOK, nil }

var likeSecret = []byte("voter-secret")

func likeBooths() *mockBoothRepository {
	booths := []*domain.Booth{
		{ID: 1, Name: "お化け屋敷", Organizer: "3-2"},
		{ID: 2, Name: "からあげ", Organizer: "天文部"},
		{ID: 3, Name: "迷路", Organizer: "1-1"},
	}
	return &mockBoothRepository{
		getByIDFn: func(_ context.Context, id int) (*domain.Booth, error) {
			for _, b := range booths {
				if b.ID == id {
					return b, nil
				}
			}
			return nil, sql.ErrNoRows
		},
		getAllFn: func(context.Context) ([]*domain.Booth, error) { return booths, nil },
	}
}

var likeNow = time.Date(2026, 10, 31, 3, 7, 0, 0, time.UTC)

func newLikeUC(repo *fakeLikeRepo) *LikeUsecase {
	return NewLikeUsecase(repo, likeBooths(), likeSecret, func() time.Time { return likeNow })
}

func roleCtx(role domain.Role) context.Context {
	return context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{ID: uuid.New(), Role: role})
}

func TestLikeUsecase_番号を作って押す(t *testing.T) {
	repo := newFakeLikeRepo()
	uc := newLikeUC(repo)
	ctx := context.Background()

	token, err := uc.IssueVoter(ctx)
	require.NoError(t, err)
	_, err = auth.ParseVoterToken(token, likeSecret)
	require.NoError(t, err, "署名付きの番号を返す")

	require.NoError(t, uc.Like(ctx, token, 1))
	mine, err := uc.Mine(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, []int{1}, mine)

	require.NoError(t, uc.Unlike(ctx, token, 1))
	mine, _ = uc.Mine(ctx, token)
	assert.Empty(t, mine)
}

func TestLikeUsecase_番号を作る上限(t *testing.T) {
	repo := newFakeLikeRepo()
	repo.newVoterOK = false
	_, err := newLikeUC(repo).IssueVoter(context.Background())
	assert.ErrorIs(t, err, ErrTooMany)
}

func TestLikeUsecase_押すときの弾き方(t *testing.T) {
	repo := newFakeLikeRepo()
	uc := newLikeUC(repo)
	ctx := context.Background()
	token, _ := uc.IssueVoter(ctx)

	assert.ErrorIs(t, uc.Like(ctx, "bad", 1), ErrUnauthorized, "不正な番号")
	assert.ErrorIs(t, uc.Unlike(ctx, "bad", 1), ErrUnauthorized, "不正な番号")
	assert.ErrorIs(t, uc.Like(ctx, token, 2), domain.ErrNotClassBooth, "バザー")
	assert.ErrorIs(t, uc.Like(ctx, token, 99), sql.ErrNoRows, "無いブース")

	repo.toggleOK = false
	assert.ErrorIs(t, uc.Like(ctx, token, 1), ErrTooMany)
	assert.ErrorIs(t, uc.Unlike(ctx, token, 1), ErrTooMany)
}

func TestLikeUsecase_不正な番号の自分のいいねは空(t *testing.T) {
	mine, err := newLikeUC(newFakeLikeRepo()).Mine(context.Background(), "bad")
	require.NoError(t, err)
	assert.Empty(t, mine)
}

func TestLikeUsecase_集計(t *testing.T) {
	repo := newFakeLikeRepo()
	base := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)
	// 迷路: 1:00 台に 2 件。お化け屋敷: 1:00 に 1 件、1:20 に 3 件
	repo.Like(nil, 3, "a", base.Add(1*time.Minute))
	repo.Like(nil, 3, "b", base.Add(9*time.Minute))
	repo.Like(nil, 1, "c", base)
	for _, v := range []string{"d", "e", "f"} {
		repo.Like(nil, 1, v, base.Add(25*time.Minute))
	}
	got, err := newLikeUC(repo).Summary(roleCtx(domain.RoleGakuseikai))
	require.NoError(t, err)

	require.Len(t, got, 2, "クラス展示だけ(バザーは入れない)")
	assert.Equal(t, 1, got[0].BoothID, "多い順")
	assert.Equal(t, 4, got[0].Total)
	assert.Equal(t, "お化け屋敷", got[0].Name)
	// 全体の範囲 1:00〜1:20 を 10 分ごとに 0 で埋める
	require.Len(t, got[0].Buckets, 3)
	assert.Equal(t, []int{1, 0, 3}, []int{got[0].Buckets[0].Count, got[0].Buckets[1].Count, got[0].Buckets[2].Count})
	assert.Equal(t, base.Add(20*time.Minute), got[0].Buckets[2].Start)
	assert.Equal(t, []int{2, 0, 0}, []int{got[1].Buckets[0].Count, got[1].Buckets[1].Count, got[1].Buckets[2].Count})
}

func TestLikeUsecase_集計で急に増えた所に印(t *testing.T) {
	repo := newFakeLikeRepo()
	base := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		repo.Like(nil, 1, uuid.NewString(), base.Add(time.Duration(i)*10*time.Minute))
	}
	for i := 0; i < 40; i++ {
		repo.Like(nil, 1, uuid.NewString(), base.Add(30*time.Minute))
	}
	got, err := newLikeUC(repo).Summary(roleCtx(domain.RoleAdmin))
	require.NoError(t, err)
	assert.True(t, got[0].Burst)
	assert.True(t, got[0].Buckets[3].Burst)
}

func TestLikeUsecase_集計の権限(t *testing.T) {
	uc := newLikeUC(newFakeLikeRepo())
	for _, r := range []domain.Role{domain.RoleStudent, domain.RoleMember} {
		_, err := uc.Summary(roleCtx(r))
		assert.ErrorIs(t, err, ErrForbidden)
	}
	_, err := uc.Summary(context.Background())
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestLikeUsecase_時間帯で取り消す(t *testing.T) {
	repo := newFakeLikeRepo()
	uc := newLikeUC(repo)
	base := time.Date(2026, 10, 31, 1, 0, 0, 0, time.UTC)
	repo.Like(nil, 1, "a", base)
	repo.Like(nil, 1, "b", base.Add(15*time.Minute))

	n, err := uc.RemoveRange(roleCtx(domain.RoleGakuseikai), 1, base, base.Add(10*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.NotEmpty(t, repo.removedBy, "誰が消したかを残す")

	_, err = uc.RemoveRange(roleCtx(domain.RoleGakuseikai), 1, base, base)
	assert.ErrorIs(t, err, domain.ErrInvalidRange)
	_, err = uc.RemoveRange(roleCtx(domain.RoleStudent), 1, base, base.Add(time.Hour))
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestLikeUsecase_全部消すのは管理者だけ(t *testing.T) {
	repo := newFakeLikeRepo()
	uc := newLikeUC(repo)
	repo.Like(nil, 1, "a", likeNow)
	_, err := uc.RemoveAll(roleCtx(domain.RoleGakuseikai))
	assert.ErrorIs(t, err, ErrForbidden)
	n, err := uc.RemoveAll(roleCtx(domain.RoleAdmin))
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}
