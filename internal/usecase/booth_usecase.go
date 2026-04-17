package usecase

import "github.com/meg44k/KCTFestNav-Backend/internal/domain"

type BoothUsecase struct {
	repo domain.BoothRepository
}

func NewBoothUsecase(repo domain.BoothRepository) *BoothUsecase {
	return &BoothUsecase{
		repo: repo,
	}
}  
