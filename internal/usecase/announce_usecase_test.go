package usecase

import (
	"context"
	"testing"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

type mockAnnouncementRepository struct {
	getFn    func(ctx context.Context) (*domain.Announcement, error)
	updateFn func(ctx context.Context, announcement *domain.Announcement) error
}

func (m *mockAnnouncementRepository) Get(ctx context.Context) (*domain.Announcement, error) {
	return m.getFn(ctx)
}

func (m *mockAnnouncementRepository) Update(ctx context.Context, announcement *domain.Announcement) error {
	return m.updateFn(ctx, announcement)
}

func TestAnnouncementUsecase_Get(t *testing.T) {
	mockRepo := &mockAnnouncementRepository{
		getFn: func(ctx context.Context) (*domain.Announcement, error) {
			return &domain.Announcement{Content: "hello"}, nil
		},
	}
	uc := NewAnnouncementUsecase(mockRepo)

	announcement, err := uc.Get(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "hello", announcement.Content)
}

func TestAnnouncementUsecase_Update(t *testing.T) {
	t.Run("success admin", func(t *testing.T) {
		var updatedContent string
		mockRepo := &mockAnnouncementRepository{
			updateFn: func(ctx context.Context, announcement *domain.Announcement) error {
				updatedContent = announcement.Content
				return nil
			},
		}
		uc := NewAnnouncementUsecase(mockRepo)

		ctx := context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{Role: domain.RoleAdmin})
		err := uc.Update(ctx, "new hello")
		assert.NoError(t, err)
		assert.Equal(t, "new hello", updatedContent)
	})

	t.Run("success gakuseikai", func(t *testing.T) {
		mockRepo := &mockAnnouncementRepository{
			updateFn: func(ctx context.Context, announcement *domain.Announcement) error {
				return nil
			},
		}
		uc := NewAnnouncementUsecase(mockRepo)

		ctx := context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{Role: domain.RoleGakuseikai})
		err := uc.Update(ctx, "new hello")
		assert.NoError(t, err)
	})

	t.Run("forbidden", func(t *testing.T) {
		uc := NewAnnouncementUsecase(&mockAnnouncementRepository{})

		ctx := context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{Role: domain.RoleStudent})
		err := uc.Update(ctx, "new hello")
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("empty content", func(t *testing.T) {
		uc := NewAnnouncementUsecase(&mockAnnouncementRepository{})
		ctx := context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{Role: domain.RoleAdmin})
		err := uc.Update(ctx, "")
		assert.ErrorIs(t, err, domain.ErrContentRequired)
	})
}
