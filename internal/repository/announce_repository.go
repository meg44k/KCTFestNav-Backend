package repository

import (
	"context"
	"sync"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

type announcementRepository struct {
	mu           sync.RWMutex
	announcement *domain.Announcement
}

func NewAnnouncementRepository() domain.AnnouncementRepository {
	return &announcementRepository{
		announcement: &domain.Announcement{
			Content: "", // 初期値
		},
	}
}

// オンメモリで実装しているからスケールできないかも。ちゃんと作るならRedis上に保存したほうがいい
func (r *announcementRepository) Get(ctx context.Context) (*domain.Announcement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 呼び出し元で予期せぬ変更が起きないように値コピーを返す
	return &domain.Announcement{
		Content: r.announcement.Content,
	}, nil
}

func (r *announcementRepository) Update(ctx context.Context, announcement *domain.Announcement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.announcement.Content = announcement.Content
	return nil
}
