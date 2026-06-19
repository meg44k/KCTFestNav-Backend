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
	name string,
	organizer string,
	detail string,
	congestionStatus int8,
	x float32,
	y float32,
	z float32,
) error {
	role, ok := ctx.Value(ContextUserRoleKey).(domain.Role)
	if !ok || role != domain.RoleAdmin {
		return ErrForbidden
	}
	booth, err := domain.NewBooth(name, organizer, detail, x, y, z)
	if err != nil {
		return err
	}
	return u.boothRepo.Create(ctx, booth)
}

func (u *BoothUsecase) Update(ctx context.Context, b *domain.Booth) error {
	return u.boothRepo.Update(ctx, b)
}

func (u *BoothUsecase) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	booth, err := u.boothRepo.GetByID(ctx, id)
	return booth, err
}

func (u *BoothUsecase) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	booths, err := u.boothRepo.GetAll(ctx)
	return booths, err
}

// REF: これBooth.CogestionStatusをカプセル化した意味がなくなっちゃってる。Redisで管理したいけど、どうするのがベストなんだろう...
func (u *BoothUsecase) UpdateCongestion(ctx context.Context, id int, congestionLevel int8) error {
	role, ok := ctx.Value(ContextUserRoleKey).(domain.Role)
	if !ok || role != domain.RoleAdmin && role != domain.RoleGakuseikai && role != domain.RoleStudent { // TODO: 所属するクラスと同じだった時に変更できるようにする
		return ErrForbidden
	}
	if err := domain.ValidateCongestionLevel(congestionLevel); err != nil {
		return err
	}
	u.boothRepo.UpdateCongestion(ctx, id, congestionLevel)
	return nil
}

func (u *BoothUsecase) Delete(ctx context.Context, id int) error {
	role, ok := ctx.Value(ContextUserRoleKey).(domain.Role)
	if !ok || role != domain.RoleAdmin {
		return ErrForbidden
	}
	return u.boothRepo.Delete(ctx, id)
}
