package usecase

import (
	"context"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type BoothUsecase struct {
	boothRepo domain.BoothRepository
}

func NewBoothUsecase(repo domain.BoothRepository) *BoothUsecase {
	return &BoothUsecase{
		boothRepo: repo,
	}
}

func (u *BoothUsecase) Create(
	ctx context.Context,
	p domain.BoothParams,
) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	booth, err := domain.NewBooth(p)
	if err != nil {
		return err
	}
	return u.boothRepo.Create(ctx, booth)
}

func (u *BoothUsecase) Update(
	ctx context.Context,
	id int,
	congestionStatus domain.CongestionStatus,
	p domain.BoothParams,
) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	// 管理者 or　学生会 or 学生出ない時、権限なしとして返す
	if !ok || (reqUser.Role != domain.RoleAdmin && reqUser.Role != domain.RoleGakuseikai && reqUser.Role != domain.RoleStudent) { // TODO: 複数のboothにアサインされている場合にも対応する
		return ErrForbidden
	}
	// 学生かつアサインされたブースでない時、権限なしとして返す
	if reqUser.Role == domain.RoleStudent && reqUser.AssignedBoothID != id {
		return ErrForbidden
	}

	booth, err := domain.ReconstructBooth(id, congestionStatus, p)
	if err != nil {
		return err
	}
	return u.boothRepo.Update(ctx, booth)
}

// REF: これBooth.CogestionStatusをカプセル化した意味がなくなっちゃってる。Redisで管理したいけど、どうするのがベストなんだろう...
func (u *BoothUsecase) UpdateCongestion(ctx context.Context, id int, congestionStatus domain.CongestionStatus) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	// 管理者 or　学生会 or 学生出ない時、権限なしとして返す
	if !ok || (reqUser.Role != domain.RoleAdmin && reqUser.Role != domain.RoleGakuseikai && reqUser.Role != domain.RoleStudent) { // TODO: 複数のboothにアサインされている場合にも対応する
		return ErrForbidden
	}
	// 学生かつアサインされたブースでない時、権限なしとして返す
	if reqUser.Role == domain.RoleStudent && reqUser.AssignedBoothID != id {
		return ErrForbidden
	}
	// 混雑度のバリデーション
	if err := domain.ValidateCongestionLevel(congestionStatus); err != nil {
		return err
	}

	if err := u.boothRepo.UpdateCongestion(ctx, id, congestionStatus); err != nil {
		return err
	}
	return nil
}
func (u *BoothUsecase) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	booth, err := u.boothRepo.GetByID(ctx, id)
	return booth, err
}

func (u *BoothUsecase) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	booths, err := u.boothRepo.GetAll(ctx)
	return booths, err
}

func (u *BoothUsecase) Delete(ctx context.Context, id int) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok || reqUser.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	return u.boothRepo.Delete(ctx, id)
}
