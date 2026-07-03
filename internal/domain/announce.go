package domain

import (
	"context"
	"strings"
)

// Announcement は電光掲示板のように画面に流すお知らせを表す型です。
// オンメモリで単一のテキストとして管理されることを想定しています。
type Announcement struct {
	Content string // お知らせの本文
}

// 新規お知らせ作成用コンストラクタ
func NewAnnouncement(content string) (*Announcement, error) {
	if strings.TrimSpace(content) == "" {
		return nil, ErrContentRequired
	}

	return &Announcement{
		Content: content,
	}, nil
}

type AnnouncementRepository interface {
	Get(ctx context.Context) (*Announcement, error)
	Update(ctx context.Context, announcement *Announcement) error
}
