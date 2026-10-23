package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ログインの失敗の上限。最初の失敗から loginWindow の間に数え、超えたら次の窓まで試させない。
// ログインは Next.js のサーバーを通って来るので、来場者の接続元では数えられない。ID ごとと、サーバー全体で数える
const (
	loginWindow       = 10 * time.Minute
	loginFailsPerID   = 10  // 1 つの ID で続けて失敗できる回数(本人の打ち間違いには十分)
	loginFailsOverall = 500 // 全体の失敗。たくさんの ID に少しずつ試す攻撃と、重い照合の計算から守る
)

const loginFailsAllKey = "login:fail:all"

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
	id, err := l.count(ctx, loginFailsKey(loginID))
	if err != nil {
		return false, err
	}
	all, err := l.count(ctx, loginFailsAllKey)
	if err != nil {
		return false, err
	}
	return id >= loginFailsPerID || all >= loginFailsOverall, nil
}

func (l *loginLimiter) Failed(ctx context.Context, loginID string) error {
	for _, key := range []string{loginFailsKey(loginID), loginFailsAllKey} {
		n, err := l.rdb.Incr(ctx, key).Result()
		if err != nil {
			return err
		}
		// 最初の失敗から数える
		if n == 1 {
			if err := l.rdb.Expire(ctx, key, loginWindow).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *loginLimiter) Succeeded(ctx context.Context, loginID string) error {
	return l.rdb.Del(ctx, loginFailsKey(loginID)).Err()
}
