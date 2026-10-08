package usecase

import (
	"context"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type AnnouncementUsecase interface {
	Get(ctx context.Context) (*domain.Announcement, error)
	Update(ctx context.Context, p domain.AnnouncementParams) error
}

type announcementUsecase struct {
	repo domain.AnnouncementRepository
}

func NewAnnouncementUsecase(repo domain.AnnouncementRepository) AnnouncementUsecase {
	return &announcementUsecase{repo: repo}
}

func (u *announcementUsecase) Get(ctx context.Context) (*domain.Announcement, error) {
	return u.repo.Get(ctx)
}

func (u *announcementUsecase) Update(ctx context.Context, p domain.AnnouncementParams) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || (reqUser.Role != domain.RoleAdmin && reqUser.Role != domain.RoleGakuseikai) {
		return ErrForbidden
	}

	announcement, err := domain.NewAnnouncement(p)
	if err != nil {
		return err
	}
	return u.repo.Update(ctx, announcement)
}
