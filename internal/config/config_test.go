package config

import (
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad_手元の開発の値(t *testing.T) {
	c, err := Load(env(map[string]string{
		"DB_USER": "root", "DB_HOST": "127.0.0.1", "DB_PORT": "3306", "DB_NAME": "kctfestnav",
		"REDIS_ADDR": "127.0.0.1:6379",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "1323" {
		t.Errorf("Port = %q, want 1323", c.Port)
	}
	if c.MySQL.Net != "tcp" || c.MySQL.Addr != "127.0.0.1:3306" {
		t.Errorf("MySQL = %s %s", c.MySQL.Net, c.MySQL.Addr)
	}
	if !c.MySQL.ParseTime || c.MySQL.DBName != "kctfestnav" || c.MySQL.User != "root" {
		t.Errorf("MySQL = %+v", c.MySQL)
	}
	if c.MaxOpenConns != 5 {
		t.Errorf("MaxOpenConns = %d, want 5", c.MaxOpenConns)
	}
	if c.Redis.TLSConfig != nil {
		t.Error("Redis は TLS なしのはず")
	}
}

func TestLoad_CloudRunの値(t *testing.T) {
	c, err := Load(env(map[string]string{
		"PORT": "8080", "DB_SOCKET": "/cloudsql/p:r:i", "DB_HOST": "ignored", "DB_PORT": "3306",
		"DB_MAX_OPEN_CONNS": "3", "REDIS_ADDR": "x.upstash.io:6379", "REDIS_PASSWORD": "pw", "REDIS_TLS": "TRUE",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "8080" {
		t.Errorf("Port = %q", c.Port)
	}
	if c.MySQL.Net != "unix" || c.MySQL.Addr != "/cloudsql/p:r:i" {
		t.Errorf("ソケットを優先するはず: %s %s", c.MySQL.Net, c.MySQL.Addr)
	}
	if c.MaxOpenConns != 3 {
		t.Errorf("MaxOpenConns = %d", c.MaxOpenConns)
	}
	if c.Redis.TLSConfig == nil || c.Redis.Password != "pw" {
		t.Error("Redis は TLS ありでパスワードを持つはず")
	}
}

func TestLoad_不正な値はエラー(t *testing.T) {
	for _, m := range []map[string]string{
		{"DB_MAX_OPEN_CONNS": "abc"},
		{"DB_MAX_OPEN_CONNS": "0"},
		{"REDIS_TLS": "yes-please"},
	} {
		if _, err := Load(env(m)); err == nil {
			t.Errorf("%v はエラーのはず", m)
		}
	}
}

func TestLoad_いいねの秘密(t *testing.T) {
	c, err := Load(env(map[string]string{"VOTER_SECRET": "vs"}))
	if err != nil {
		t.Fatal(err)
	}
	if string(c.VoterSecret) != "vs" {
		t.Errorf("VoterSecret = %q", c.VoterSecret)
	}
	// 無くても起動はできる(いいねが 401 になるだけ)
	c, err = Load(env(map[string]string{}))
	if err != nil || len(c.VoterSecret) != 0 {
		t.Errorf("空でよい: %v %q", err, c.VoterSecret)
	}
}
