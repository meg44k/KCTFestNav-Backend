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

func (u *LiveUsecase) Create(ctx context.Context, l *domain.Live) error {
	return u.liveRepo.Create(ctx, l)
}

func (u *LiveUsecase) Update(ctx context.Context, l *domain.Live) error {
	return u.liveRepo.Update(ctx, l)
}

func (u *LiveUsecase) GetByID(ctx context.Context, id int) (*domain.Live, error) {
	live, err := u.liveRepo.GetByID(ctx, id)
	return live, err
}

func (u *LiveUsecase) GetAll(ctx context.Context) ([]*domain.Live, error) {
	lives, err := u.liveRepo.GetAll(ctx)
	return lives, err
}
