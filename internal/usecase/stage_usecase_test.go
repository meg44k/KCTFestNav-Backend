package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 呼ばれたメソッド名を記録するだけの StageRepository
type fakeStageRepo struct {
	calls []string
	block *domain.StageBlock
}

func (f *fakeStageRepo) rec(name string) { f.calls = append(f.calls, name) }

func (f *fakeStageRepo) GetSchedule(context.Context) ([]*domain.StageSection, error) {
	f.rec("GetSchedule")
	return []*domain.StageSection{{ID: 1, Name: "Live1"}}, nil
}
func (f *fakeStageRepo) GetSection(context.Context, int) (*domain.StageSection, error) {
	f.rec("GetSection")
	return nil, nil
}
func (f *fakeStageRepo) GetBlock(context.Context, int) (*domain.StageBlock, error) {
	f.rec("GetBlock")
	return f.block, nil
}
func (f *fakeStageRepo) GetPerformer(context.Context, int) (*domain.Performer, error) {
	f.rec("GetPerformer")
	return nil, nil
}
func (f *fakeStageRepo) CreateSection(_ context.Context, s *domain.StageSection) (int, error) {
	f.rec("CreateSection:" + s.Name)
	return 11, nil
}
func (f *fakeStageRepo) UpdateSection(_ context.Context, s *domain.StageSection) error {
	f.rec("UpdateSection:" + s.Name)
	return nil
}
func (f *fakeStageRepo) DeleteSection(context.Context, int) error { f.rec("DeleteSection"); return nil }
func (f *fakeStageRepo) CreateBlock(context.Context, *domain.StageBlock) (int, error) {
	f.rec("CreateBlock")
	return 12, nil
}
func (f *fakeStageRepo) UpdateBlock(context.Context, *domain.StageBlock) error {
	f.rec("UpdateBlock")
	return nil
}
func (f *fakeStageRepo) DeleteBlock(context.Context, int) error { f.rec("DeleteBlock"); return nil }
func (f *fakeStageRepo) CreatePerformer(_ context.Context, p *domain.Performer) (int, error) {
	f.rec("CreatePerformer:" + p.Name)
	return 13, nil
}
func (f *fakeStageRepo) UpdatePerformer(context.Context, *domain.Performer) error {
	f.rec("UpdatePerformer")
	return nil
}
func (f *fakeStageRepo) DeletePerformer(context.Context, int) error {
	f.rec("DeletePerformer")
	return nil
}
func (f *fakeStageRepo) MovePerformer(_ context.Context, _ int, d domain.MoveDirection) error {
	f.rec("MovePerformer:" + string(d))
	return nil
}
func (f *fakeStageRepo) AdvanceBlock(context.Context, int) error { f.rec("AdvanceBlock"); return nil }
func (f *fakeStageRepo) RewindBlock(context.Context, int) error  { f.rec("RewindBlock"); return nil }

var (
	t0 = time.Date(2026, 10, 31, 4, 0, 0, 0, time.UTC)
	t1 = t0.Add(time.Hour)
)

// 管理者だけができる操作を、すべて正しい入力で呼ぶ
func adminOps(uc *StageUsecase, ctx context.Context) []error {
	_, e1 := uc.CreateSection(ctx, SectionInput{Name: "Live1"})
	_, e2 := uc.CreateBlock(ctx, 1, t0, t1)
	_, e3 := uc.CreatePerformer(ctx, 1, PerformerInput{Name: "A"})
	return []error{
		e1, e2, e3,
		uc.UpdateSection(ctx, 1, SectionInput{Name: "Live1"}),
		uc.DeleteSection(ctx, 1),
		uc.UpdateBlock(ctx, 1, t0, t1),
		uc.DeleteBlock(ctx, 1),
		uc.UpdatePerformer(ctx, 1, PerformerInput{Name: "A"}),
		uc.DeletePerformer(ctx, 1),
		uc.MovePerformer(ctx, 1, domain.MoveUp),
	}
}

func TestStageUsecase_Permissions(t *testing.T) {
	t.Run("管理者はすべてできる", func(t *testing.T) {
		repo := &fakeStageRepo{block: &domain.StageBlock{ID: 1}}
		uc := NewStageUsecase(repo)
		ctx := ctxWithRole(domain.RoleAdmin)
		for i, err := range adminOps(uc, ctx) {
			assert.NoError(t, err, "操作 %d", i)
		}
		_, err := uc.AdvanceBlock(ctx, 1)
		assert.NoError(t, err)
		_, err = uc.RewindBlock(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("学生会は進める・戻すだけ", func(t *testing.T) {
		repo := &fakeStageRepo{block: &domain.StageBlock{ID: 1, CurrentOrder: 2}}
		uc := NewStageUsecase(repo)
		ctx := ctxWithRole(domain.RoleGakuseikai)
		for i, err := range adminOps(uc, ctx) {
			assert.ErrorIs(t, err, ErrForbidden, "操作 %d", i)
		}
		b, err := uc.AdvanceBlock(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, 2, b.CurrentOrder, "更新後のブロックを返す")
		_, err = uc.RewindBlock(ctx, 1)
		require.NoError(t, err)
		assert.Equal(t, []string{"AdvanceBlock", "GetBlock", "RewindBlock", "GetBlock"}, repo.calls)
	})

	for _, ctx := range map[string]context.Context{
		"企画担当":  ctxWithRole(domain.RoleStudent),
		"未ログイン": context.Background(),
	} {
		repo := &fakeStageRepo{}
		uc := NewStageUsecase(repo)
		for i, err := range adminOps(uc, ctx) {
			assert.ErrorIs(t, err, ErrForbidden, "操作 %d", i)
		}
		_, err := uc.AdvanceBlock(ctx, 1)
		assert.ErrorIs(t, err, ErrForbidden)
		_, err = uc.RewindBlock(ctx, 1)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Empty(t, repo.calls, "権限が無ければ repository を呼ばない")
	}
}

func TestStageUsecase_Validation(t *testing.T) {
	repo := &fakeStageRepo{}
	uc := NewStageUsecase(repo)
	ctx := ctxWithRole(domain.RoleAdmin)

	_, err := uc.CreateSection(ctx, SectionInput{Name: " "})
	assert.ErrorIs(t, err, domain.ErrNameRequired)
	assert.ErrorIs(t, uc.UpdateSection(ctx, 1, SectionInput{Name: ""}), domain.ErrNameRequired)
	_, err = uc.CreateBlock(ctx, 1, t1, t0)
	assert.ErrorIs(t, err, domain.ErrEndTimeAfterStartTime)
	assert.ErrorIs(t, uc.UpdateBlock(ctx, 1, t0, t0), domain.ErrEndTimeAfterStartTime)
	_, err = uc.CreatePerformer(ctx, 1, PerformerInput{Name: ""})
	assert.ErrorIs(t, err, domain.ErrNameRequired)
	assert.ErrorIs(t, uc.UpdatePerformer(ctx, 1, PerformerInput{Name: "　"}), domain.ErrNameRequired)
	assert.ErrorIs(t, uc.MovePerformer(ctx, 1, "left"), domain.ErrInvalidDirection)
	assert.Empty(t, repo.calls, "入力が不正なら repository を呼ばない")
}

func TestStageUsecase_CreateReturnsID(t *testing.T) {
	repo := &fakeStageRepo{}
	uc := NewStageUsecase(repo)
	ctx := ctxWithRole(domain.RoleAdmin)
	id, err := uc.CreateSection(ctx, SectionInput{Name: "Live1", Location: "第一体育館", SortOrder: 1})
	require.NoError(t, err)
	assert.Equal(t, 11, id)
	id, _ = uc.CreateBlock(ctx, 11, t0, t1)
	assert.Equal(t, 12, id)
	id, _ = uc.CreatePerformer(ctx, 12, PerformerInput{Name: "A"})
	assert.Equal(t, 13, id)
	assert.Equal(t, []string{"CreateSection:Live1", "CreateBlock", "CreatePerformer:A"}, repo.calls)
}

func TestStageUsecase_GetScheduleIsPublic(t *testing.T) {
	uc := NewStageUsecase(&fakeStageRepo{})
	got, err := uc.GetSchedule(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "Live1", got[0].Name)
}

// context に Role を詰めるヘルパー
func ctxWithRole(role domain.Role) context.Context {
	return context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{Role: role})
}
