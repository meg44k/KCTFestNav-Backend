package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestAnnouncementRepository(t *testing.T) {
	repo := NewAnnouncementRepository()
	ctx := context.Background()

	t.Run("Initial state", func(t *testing.T) {
		announcement, err := repo.Get(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "", announcement.Content)
	})

	t.Run("Update and Get", func(t *testing.T) {
		err := repo.Update(ctx, &domain.Announcement{Content: "new announcement"})
		assert.NoError(t, err)

		announcement, err := repo.Get(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "new announcement", announcement.Content)
	})

	t.Run("Concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = repo.Update(ctx, &domain.Announcement{Content: "concurrent"})
				_, _ = repo.Get(ctx)
			}()
		}
		wg.Wait()
		announcement, err := repo.Get(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "concurrent", announcement.Content)
	})
}
