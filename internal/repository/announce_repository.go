package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// お知らせのキー。1 つだけ持ち、期限は付けない
const announcementKey = "announcement"

// 読んだお知らせを覚えておく長さ。来場者のページを開くたびに読むので、Redis へはサーバー 1 台につき
// この間に 1 回だけ聞きに行く(書き換えはほかのサーバーに最大この長さ遅れて届く)
const announcementCacheTTL = 10 * time.Second

// お知らせは Redis に置く。サーバーのメモリに置くと、デプロイやサーバーが止まったときに消え、
// 混んでサーバーが増えるとサーバーごとに違うお知らせになるため
type announcementRepository struct {
	rdb *redis.Client
	now func() time.Time

	mu        sync.Mutex
	cached    string
	fetchedAt time.Time
	hasCache  bool
}

func NewAnnouncementRepository(rdb *redis.Client) domain.AnnouncementRepository {
	return newAnnouncementRepository(rdb, time.Now)
}

func newAnnouncementRepository(rdb *redis.Client, now func() time.Time) *announcementRepository {
	return &announcementRepository{rdb: rdb, now: now}
}

func (r *announcementRepository) Get(ctx context.Context) (*domain.Announcement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.hasCache && r.now().Sub(r.fetchedAt) < announcementCacheTTL {
		return &domain.Announcement{Content: r.cached}, nil
	}
	content, err := r.rdb.Get(ctx, announcementKey).Result()
	if errors.Is(err, redis.Nil) {
		content, err = "", nil // まだ一度も入れていない
	}
	if err != nil {
		return nil, err
	}
	r.cached, r.fetchedAt, r.hasCache = content, r.now(), true
	return &domain.Announcement{Content: content}, nil
}

func (r *announcementRepository) Update(ctx context.Context, announcement *domain.Announcement) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.rdb.Set(ctx, announcementKey, announcement.Content, 0).Err(); err != nil {
		return err
	}
	// 書き換えたサーバーではすぐ新しいものを出す
	r.cached, r.fetchedAt, r.hasCache = announcement.Content, r.now(), true
	return nil
}
