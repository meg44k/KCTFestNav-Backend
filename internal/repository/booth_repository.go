package repository

import (
	"context"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type boothRepositoryImpl struct{}

func NewBoothRepository() domain.BoothRepository {
	return &boothRepositoryImpl{}
}

func (br *boothRepositoryImpl) Create(ctx context.Context, b *domain.Booth) error {
	// TODO: 実際の実装を行う
	return nil
}

func (br *boothRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	// TODO: 実装を行う
	return nil, nil
}

func (br *boothRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	// TODO: 実装を行う
	return nil, nil

}

func (br *boothRepositoryImpl) Update(ctx context.Context, b *domain.Booth) error {
	// TODO: 実装を行う
	return nil
}

func (br *boothRepositoryImpl) UpdateCongestion(ctx context.Context, id int, congestionLevel int8) error {
	// TODO: 実装を行う
	return nil
}
