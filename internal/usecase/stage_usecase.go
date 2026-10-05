package usecase

import (
	"context"
	"time"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 番組表(セクション・ブロック・出演者)の編集は管理者だけ。
// 当日の「次のバンドへ」「前に戻す」は学生会もできる
type StageUsecase struct {
	repo domain.StageRepository
}

func NewStageUsecase(repo domain.StageRepository) *StageUsecase {
	return &StageUsecase{repo: repo}
}

type SectionInput struct {
	Name      string
	Location  string
	SortOrder int
}

type PerformerInput struct {
	Name         string
	Detail       string
	ThumbnailURL string
}

func hasRole(ctx context.Context, roles ...domain.Role) bool {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok {
		return false
	}
	for _, r := range roles {
		if reqUser.Role == r {
			return true
		}
	}
	return false
}

func isAdmin(ctx context.Context) bool { return hasRole(ctx, domain.RoleAdmin) }

func (u *StageUsecase) GetSchedule(ctx context.Context) ([]*domain.StageSection, error) {
	return u.repo.GetSchedule(ctx)
}

func (u *StageUsecase) CreateSection(ctx context.Context, in SectionInput) (int, error) {
	if !isAdmin(ctx) {
		return 0, ErrForbidden
	}
	s, err := domain.NewStageSection(in.Name, in.Location, in.SortOrder)
	if err != nil {
		return 0, err
	}
	return u.repo.CreateSection(ctx, s)
}

func (u *StageUsecase) UpdateSection(ctx context.Context, id int, in SectionInput) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	s, err := domain.NewStageSection(in.Name, in.Location, in.SortOrder)
	if err != nil {
		return err
	}
	s.ID = id
	return u.repo.UpdateSection(ctx, s)
}

func (u *StageUsecase) DeleteSection(ctx context.Context, id int) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	return u.repo.DeleteSection(ctx, id)
}

func (u *StageUsecase) CreateBlock(ctx context.Context, sectionID int, start, end time.Time) (int, error) {
	if !isAdmin(ctx) {
		return 0, ErrForbidden
	}
	b, err := domain.NewStageBlock(sectionID, start, end)
	if err != nil {
		return 0, err
	}
	return u.repo.CreateBlock(ctx, b)
}

func (u *StageUsecase) UpdateBlock(ctx context.Context, id int, start, end time.Time) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	b, err := domain.NewStageBlock(0, start, end)
	if err != nil {
		return err
	}
	b.ID = id
	return u.repo.UpdateBlock(ctx, b)
}

func (u *StageUsecase) DeleteBlock(ctx context.Context, id int) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	return u.repo.DeleteBlock(ctx, id)
}

func (u *StageUsecase) CreatePerformer(ctx context.Context, blockID int, in PerformerInput) (int, error) {
	if !isAdmin(ctx) {
		return 0, ErrForbidden
	}
	p, err := domain.NewPerformer(blockID, in.Name, in.Detail, in.ThumbnailURL)
	if err != nil {
		return 0, err
	}
	return u.repo.CreatePerformer(ctx, p)
}

func (u *StageUsecase) UpdatePerformer(ctx context.Context, id int, in PerformerInput) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	p, err := domain.NewPerformer(0, in.Name, in.Detail, in.ThumbnailURL)
	if err != nil {
		return err
	}
	p.ID = id
	return u.repo.UpdatePerformer(ctx, p)
}

func (u *StageUsecase) DeletePerformer(ctx context.Context, id int) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	return u.repo.DeletePerformer(ctx, id)
}

func (u *StageUsecase) MovePerformer(ctx context.Context, id int, d domain.MoveDirection) error {
	if !isAdmin(ctx) {
		return ErrForbidden
	}
	if err := domain.ValidateMoveDirection(d); err != nil {
		return err
	}
	return u.repo.MovePerformer(ctx, id, d)
}

// 演奏中を 1 組進める。画面にすぐ反映できるよう更新後のブロックを返す
func (u *StageUsecase) AdvanceBlock(ctx context.Context, id int) (*domain.StageBlock, error) {
	if !hasRole(ctx, domain.RoleAdmin, domain.RoleGakuseikai) {
		return nil, ErrForbidden
	}
	if err := u.repo.AdvanceBlock(ctx, id); err != nil {
		return nil, err
	}
	return u.repo.GetBlock(ctx, id)
}

func (u *StageUsecase) RewindBlock(ctx context.Context, id int) (*domain.StageBlock, error) {
	if !hasRole(ctx, domain.RoleAdmin, domain.RoleGakuseikai) {
		return nil, ErrForbidden
	}
	if err := u.repo.RewindBlock(ctx, id); err != nil {
		return nil, err
	}
	return u.repo.GetBlock(ctx, id)
}
