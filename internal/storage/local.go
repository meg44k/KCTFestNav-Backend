// Package storage は画像の置き場所。本番は R2、手元はフォルダ
package storage

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// 手元の開発用。フォルダに置き、バックエンドが /images/* で配る
type Local struct {
	dir     string
	baseURL string
}

func NewLocal(dir, baseURL string) *Local {
	return &Local{dir: dir, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *Local) Put(_ context.Context, key, _ string, data []byte) (string, error) {
	path := filepath.Join(s.dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return s.baseURL + "/" + key, nil
}

func (s *Local) Delete(_ context.Context, url string) error {
	key, ok := keyOf(s.baseURL, url)
	if !ok {
		return nil
	}
	err := os.Remove(filepath.Join(s.dir, filepath.FromSlash(key)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Local) KeyOf(url string) (string, bool) { return keyOf(s.baseURL, url) }

func (s *Local) Exists(_ context.Context, url string) (bool, error) {
	key, ok := keyOf(s.baseURL, url)
	if !ok {
		return false, nil
	}
	_, err := os.Stat(filepath.Join(s.dir, filepath.FromSlash(key)))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// base/ の後ろの部分。自分の置き場所の URL でなければ ok=false
func keyOf(base, url string) (string, bool) {
	key, ok := strings.CutPrefix(url, base+"/")
	// booths/3/x.webp のように必ずフォルダを含む。.. や「.」だけのものは通さない
	if !ok || !strings.Contains(key, "/") || strings.Contains(key, "..") {
		return "", false
	}
	return key, true
}
