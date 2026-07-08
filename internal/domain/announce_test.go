package domain

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewAnnouncement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		announcement, err := NewAnnouncement(AnnouncementParams{Content: "test content"})
		assert.NoError(t, err)
		assert.Equal(t, "test content", announcement.Content)
	})

	t.Run("empty content", func(t *testing.T) {
		announcement, err := NewAnnouncement(AnnouncementParams{Content: "   "})
		assert.ErrorIs(t, err, ErrContentRequired)
		assert.Nil(t, announcement)
	})
}
