package usecase

import (
	"context"
	"time"

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
	name string,
	detail string,
	thumbnailURL string,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8,
) error {
	// 権限チェック: Admin以外不可
	role, ok := ctx.Value(domain.UserRoleKey).(domain.Role)
	if !ok || role != domain.RoleAdmin {
		return ErrForbidden
	}

	live, err := domain.NewLive(name, detail, thumbnailURL, startTime, endTime, sessionNumber)
	if err != nil {
		return err
	}
	return u.liveRepo.Create(ctx, live)
}

func (u *LiveUsecase) Update(
	ctx context.Context,
	id int,
	name string,
	detail string,
	thumbnailURL string,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8,
	status domain.LiveStatus,
) error {
	// 権限チェック: 学生会員以上の権限がないと不可
	role, ok := ctx.Value(domain.UserRoleKey).(domain.Role)
	if !ok || (role != domain.RoleAdmin && role != domain.RoleGakuseikai) {
		return ErrForbidden
	}

	live, err := domain.ReconstructLive(id, name, detail, thumbnailURL, startTime, endTime, sessionNumber, status)
	if err != nil {
		return err
	}
	return u.liveRepo.Update(ctx, live)
}

func (u *LiveUsecase) Delete(ctx context.Context, id int) error {
	// 権限チェック: Admin以外不可
	role, ok := ctx.Value(domain.UserRoleKey).(domain.Role)
	if !ok || role != domain.RoleAdmin {
		return ErrForbidden
	}

	err := u.liveRepo.Delete(ctx, id)
	return err
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
