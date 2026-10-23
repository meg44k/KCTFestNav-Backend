package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ログインの失敗の上限。同じ接続元・同じ ID の失敗を、最初の失敗から loginWindow の間に数え、超えたら次の窓まで試させない。
// 接続元ごとに数えるので、わざと失敗を重ねても、ほかの場所からログインする本人は締め出されない。
// ID ごと(接続元を問わず)や接続元ごと(ID を問わず)では止めない。推測できる ID や、学内 Wi-Fi のように
// 係の全員が同じ接続元のときに、誰かの失敗やいたずらで大勢が止まってしまうため。
// たくさんの ID に少しずつ試す攻撃は BotID(フロント)で止める
const (
	loginWindow       = 10 * time.Minute
	loginFailsPerPair = 5 // 同じ接続元・同じ ID で試せる回数
)

type loginLimiter struct{ rdb *redis.Client }

func NewLoginLimiter(rdb *redis.Client) domain.LoginLimiter {
	return &loginLimiter{rdb: rdb}
}

func pairKey(ip, loginID string) string { return "login:fail:pair:" + ip + ":" + loginID }

func (l *loginLimiter) count(ctx context.Context, key string) (int, error) {
	n, err := l.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (l *loginLimiter) Blocked(ctx context.Context, ip, loginID string) (bool, error) {
	n, err := l.count(ctx, pairKey(ip, loginID))
	if err != nil {
		return false, err
	}
	return n >= loginFailsPerPair, nil
}

// 数を 1 つ増やす。最初の失敗から数える
func (l *loginLimiter) incr(ctx context.Context, key string) (int, error) {
	n, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 {
		if err := l.rdb.Expire(ctx, key, loginWindow).Err(); err != nil {
			return 0, err
		}
	}
	return int(n), nil
}

func (l *loginLimiter) Failed(ctx context.Context, ip, loginID string) (int, error) {
	n, err := l.incr(ctx, pairKey(ip, loginID))
	if err != nil {
		return 0, err
	}
	return max(0, loginFailsPerPair-n), nil
}

func (l *loginLimiter) Succeeded(ctx context.Context, ip, loginID string) error {
	return l.rdb.Del(ctx, pairKey(ip, loginID)).Err()
}
