package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ログインの失敗の上限。最初の失敗から loginWindow の間に ID ごとに数え、超えたらその ID だけ次の窓まで試させない。
// ログインは Next.js のサーバーを通って来るので、来場者の接続元では数えられない。
// サーバー全体では数えない(でたらめな ID で失敗を重ねるだけで、誰もログインできなくなるため)
const (
	loginWindow     = 10 * time.Minute
	loginFailsPerID = 10 // 1 つの ID で続けて失敗できる回数(本人の打ち間違いには十分)
)

type loginLimiter struct{ rdb *redis.Client }

func NewLoginLimiter(rdb *redis.Client) domain.LoginLimiter {
	return &loginLimiter{rdb: rdb}
}

func loginFailsKey(loginID string) string { return "login:fail:id:" + loginID }

func (l *loginLimiter) count(ctx context.Context, key string) (int64, error) {
	n, err := l.rdb.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (l *loginLimiter) Blocked(ctx context.Context, loginID string) (bool, error) {
	n, err := l.count(ctx, loginFailsKey(loginID))
	if err != nil {
		return false, err
	}
	return n >= loginFailsPerID, nil
}

func (l *loginLimiter) Failed(ctx context.Context, loginID string) error {
	key := loginFailsKey(loginID)
	n, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	// 最初の失敗から数える
	if n == 1 {
		return l.rdb.Expire(ctx, key, loginWindow).Err()
	}
	return nil
}

func (l *loginLimiter) Succeeded(ctx context.Context, loginID string) error {
	return l.rdb.Del(ctx, loginFailsKey(loginID)).Err()
}
