package repository

import (
	"context"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type liveRepositoryImpl struct {}

func NewLiveRepository() domain.LiveRepository{
	return &liveRepositoryImpl{}
}

func (lr *liveRepositoryImpl) Create(ctx context.Context, l *domain.Live) error {
	// TODO:実際の実装を書く
	return nil
}

func (lr *liveRepositoryImpl) GetByID(ctx context.Context, id string) (*domain.Live, error) {
	// TODO: 実装を行う
	return nil, nil
}

func (lr *liveRepositoryImpl) GetAll(ctx context.Context) ([]*domain.Live, error) {
	// TODO: 実装を行う
	return nil, nil

}

func (lr *liveRepositoryImpl) Update(ctx context.Context, l *domain.Live) error{
	// TODO: 実装を行う
	return nil
}

func (lr *liveRepositoryImpl) Delete(ctx context.Context, id string) error{
	// TODO: 実装を行う
	return nil
}

