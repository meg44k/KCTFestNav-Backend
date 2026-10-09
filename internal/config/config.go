// Package config は環境変数からサーバーの設定を作る。
// 手元の開発(.env)と Cloud Run(PORT・Cloud SQL のソケット・Upstash の TLS)の両方を扱う
package config

import (
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

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
	Images      ImagesConfig
}

// 画像の置き場所。R2Bucket が空なら手元のフォルダ(UploadDir)に置く
type ImagesConfig struct {
	BaseURL     string
	UploadDir   string
	R2AccountID string
	R2Bucket    string
	R2KeyID     string
	R2Secret    string
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

	images := ImagesConfig{
		BaseURL:     strings.TrimRight(getenv("IMAGE_BASE_URL"), "/"),
		UploadDir:   getenv("UPLOAD_DIR"),
		R2AccountID: getenv("R2_ACCOUNT_ID"),
		R2Bucket:    getenv("R2_BUCKET"),
		R2KeyID:     getenv("R2_ACCESS_KEY_ID"),
		R2Secret:    getenv("R2_SECRET_ACCESS_KEY"),
	}
	// 手元の開発: フロントは https なので、Next.js の開発サーバーが /dev-images/* を中継する
	if images.BaseURL == "" {
		images.BaseURL = "https://localhost:3000/dev-images"
	}
	if images.UploadDir == "" {
		images.UploadDir = "./uploads"
	}

	return &Config{Port: port, MySQL: my, MaxOpenConns: maxOpen, Redis: rd, VoterSecret: []byte(getenv("VOTER_SECRET")), Images: images}, nil
}
