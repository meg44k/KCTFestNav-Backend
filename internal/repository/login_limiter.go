package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ログインの失敗の上限。最初の失敗から loginWindow の間に数え、超えたら次の窓まで試させない。
// 接続元ごとに数えるので、わざと失敗を重ねても、ほかの場所からログインする本人は締め出されない。
// ID ごと(接続元を問わず)では数えない(推測できる ID を、誰でも止められてしまうため)
const (
	loginWindow       = 10 * time.Minute
	loginFailsPerPair = 5  // 同じ接続元・同じ ID で試せる回数
	loginFailsPerIP   = 30 // 同じ接続元で、ID を問わず失敗できる回数(たくさんの ID に少しずつ試すのを止める。学校の Wi-Fi のように大勢が同じ接続元でも困らない程度)
)

type loginLimiter struct{ rdb *redis.Client }

func NewLoginLimiter(rdb *redis.Client) domain.LoginLimiter {
	return &loginLimiter{rdb: rdb}
}

func pairKey(ip, loginID string) string { return "login:fail:pair:" + ip + ":" + loginID }
func ipKey(ip string) string            { return "login:fail:ip:" + ip }

func (l *loginLimiter) count(ctx context.Context, key string) (int, error) {
	n, err := l.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (l *loginLimiter) Blocked(ctx context.Context, ip, loginID string) (bool, error) {
	pair, err := l.count(ctx, pairKey(ip, loginID))
	if err != nil {
		return false, err
	}
	all, err := l.count(ctx, ipKey(ip))
	if err != nil {
		return false, err
	}
	return pair >= loginFailsPerPair || all >= loginFailsPerIP, nil
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
	pair, err := l.incr(ctx, pairKey(ip, loginID))
	if err != nil {
		return 0, err
	}
	all, err := l.incr(ctx, ipKey(ip))
	if err != nil {
		return 0, err
	}
	return max(0, min(loginFailsPerPair-pair, loginFailsPerIP-all)), nil
}

func (l *loginLimiter) Succeeded(ctx context.Context, ip, loginID string) error {
	return l.rdb.Del(ctx, pairKey(ip, loginID)).Err()
}
