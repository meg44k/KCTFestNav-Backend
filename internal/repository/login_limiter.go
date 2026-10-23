package repository

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ログインの失敗の上限。同じブラウザ(Next.js のサーバーが渡す cookie の番号)・同じ ID の失敗を、
// 最初の失敗から loginWindow の間に数え、超えたらそのブラウザのその ID だけ次の窓まで試させない。
// 止まるのは失敗したブラウザだけなので、同じ Wi-Fi の誰かが失敗やいたずらをしても、本人は締め出されない。
// cookie は作り直せるので、自動で繰り返す攻撃は BotID(フロント)で止め、パスワードはランダムで長いものにしておく
const (
	loginWindow       = 10 * time.Minute
	loginFailsPerPair = 5 // 同じブラウザ・同じ ID で試せる回数
)

type loginLimiter struct{ rdb *redis.Client }

func NewLoginLimiter(rdb *redis.Client) domain.LoginLimiter {
	return &loginLimiter{rdb: rdb}
}

func pairKey(client, loginID string) string { return "login:fail:" + client + ":" + loginID }

func (l *loginLimiter) count(ctx context.Context, key string) (int, error) {
	n, err := l.rdb.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return n, err
}

func (l *loginLimiter) Blocked(ctx context.Context, client, loginID string) (bool, error) {
	n, err := l.count(ctx, pairKey(client, loginID))
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

func (l *loginLimiter) Failed(ctx context.Context, client, loginID string) (int, error) {
	n, err := l.incr(ctx, pairKey(client, loginID))
	if err != nil {
		return 0, err
	}
	return max(0, loginFailsPerPair-n), nil
}

func (l *loginLimiter) Succeeded(ctx context.Context, client, loginID string) error {
	return l.rdb.Del(ctx, pairKey(client, loginID)).Err()
}
