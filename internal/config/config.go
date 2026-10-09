// Package config は環境変数からサーバーの設定を作る。
// 手元の開発(.env)と Cloud Run(PORT・Cloud SQL のソケット・Upstash の TLS)の両方を扱う
package config

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Port         string
	MySQL        *mysql.Config
	MaxOpenConns int
	Redis        *redis.Options
	// いいねの投票者番号に署名する秘密。空ならいいねは全部 401 になる
	VoterSecret []byte
}

func Load(getenv func(string) string) (*Config, error) {
	port := getenv("PORT")
	if port == "" {
		port = "1323"
	}

	my := mysql.NewConfig()
	my.User = getenv("DB_USER")
	my.Passwd = getenv("DB_PASS")
	my.DBName = getenv("DB_NAME")
	my.ParseTime = true
	// Cloud Run では Cloud SQL の Unix ソケットでつなぐ
	if sock := getenv("DB_SOCKET"); sock != "" {
		my.Net = "unix"
		my.Addr = sock
	} else {
		my.Net = "tcp"
		my.Addr = getenv("DB_HOST") + ":" + getenv("DB_PORT")
	}

	maxOpen := 5
	if v := getenv("DB_MAX_OPEN_CONNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return nil, fmt.Errorf("DB_MAX_OPEN_CONNS は 1 以上の整数: %q", v)
		}
		maxOpen = n
	}

	rd := &redis.Options{
		Addr:     getenv("REDIS_ADDR"),
		Password: getenv("REDIS_PASSWORD"),
		DB:       0,
	}
	// Upstash は TLS が要る
	if v := getenv("REDIS_TLS"); v != "" {
		on, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("REDIS_TLS は true/false: %q", v)
		}
		if on {
			rd.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}
	}

	return &Config{Port: port, MySQL: my, MaxOpenConns: maxOpen, Redis: rd, VoterSecret: []byte(getenv("VOTER_SECRET"))}, nil
}
