package usecase

import (
	"context"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type LiveUsecase struct {
	liveRepo domain.LiveRepository
}

func NewLiveUsecase(repo domain.LiveRepository) *LiveUsecase {
	return &LiveUsecase{
		liveRepo: repo,
	}
}

func (u *LiveUsecase) Create(
	ctx context.Context,
	p domain.LiveParams,
) error {
	// 権限チェック: Admin以外不可
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}

	live, err := domain.NewLive(p)
	if err != nil {
		return err
	}
	return u.liveRepo.Create(ctx, live)
}

func (u *LiveUsecase) Update(
	ctx context.Context,
	id int,
	status domain.LiveStatus,
	p domain.LiveParams,
) error {
	// 権限チェック: 学生会員以上の権限がないと不可
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || (reqUser.Role != domain.RoleAdmin && reqUser.Role != domain.RoleGakuseikai) {
		return ErrForbidden
	}

	live, err := domain.ReconstructLive(id, status, p)
	if err != nil {
		return err
	}
	return u.liveRepo.Update(ctx, live)
}

func (u *LiveUsecase) UpdateLiveStatus(ctx context.Context, id int, status domain.LiveStatus) error {
	// 権限チェック
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin && reqUser.Role != domain.RoleGakuseikai { // Adminと学生会以外不可
		return ErrForbidden
	}
	return u.liveRepo.UpdateLiveStatus(ctx, id, status)
}

func (u *LiveUsecase) Delete(ctx context.Context, id int) error {
	// 権限チェック: Admin以外不可
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return u.liveRepo.Delete(ctx, id)
}

func (u *LiveUsecase) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	live, err := u.liveRepo.GetByID(ctx, id)
	return live, err
}

func (u *LiveUsecase) GetAll(ctx context.Context) ([]*domain.Live, error) {
	lives, err := u.liveRepo.GetAll(ctx)
	return lives, err
}

func (u *LiveUsecase) GetCurrentLive(ctx context.Context) (*domain.Live, error) {
	live, err := u.liveRepo.GetCurrentLive(ctx)
	return live, err
}
