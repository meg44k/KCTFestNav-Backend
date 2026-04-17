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

func (u *BoothUsecase) CreateBooth(ctx context.Context, b *domain.Booth) error {
	u.boothRepo.Create(ctx, b)
	return nil
} 

func (u *BoothUsecase) UpdateBooth(ctx context.Context, b *domain.Booth) error {
	u.boothRepo.Update(ctx, b)
	return nil
}

func (u *BoothUsecase) GetBoothByID(ctx context.Context, id string) (*domain.Booth, error) {
	booth, err := u.boothRepo.GetByID(ctx, id)
	return booth, err 
}

func (u *BoothUsecase) GetAllBooths(ctx context.Context) ([]*domain.Booth, error) {
	booths, err := u.boothRepo.GetAllBooths(ctx)
	return booths, err
}

// REF: これBooth.CogestionStatusをカプセル化した意味がなくなっちゃってる。Redisで管理したいけど、どうするのがベストなんだろう...
func (u *BoothUsecase) UpdateBoothCongestion(ctx context.Context, id string, congestionLevel int) error {
	if err := domain.ValidateCongestionLevel(congestionLevel); err != nil{
		return err
	}
	u.boothRepo.UpdateCongestion(ctx, id, congestionLevel)
	return nil
}

