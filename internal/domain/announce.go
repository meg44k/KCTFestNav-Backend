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

type AnnouncementParams struct {
	Content string
}

// 新規お知らせ作成用コンストラクタ
func NewAnnouncement(p AnnouncementParams) (*Announcement, error) {
	if strings.TrimSpace(p.Content) == "" {
		return nil, ErrContentRequired
	}

	return &Announcement{
		Content: p.Content,
	}, nil
}

type AnnouncementRepository interface {
	Get(ctx context.Context) (*Announcement, error)
	Update(ctx context.Context, announcement *Announcement) error
}
