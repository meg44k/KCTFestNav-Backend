package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocal_置いて消す(t *testing.T) {
	dir := t.TempDir()
	s := NewLocal(dir, "https://localhost:3000/dev-images")
	ctx := context.Background()

	url, err := s.Put(ctx, "booths/3/a.webp", "image/webp", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://localhost:3000/dev-images/booths/3/a.webp" {
		t.Errorf("url = %q", url)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "booths/3/a.webp")); err != nil || string(b) != "data" {
		t.Fatalf("ファイルが無い: %v", err)
	}

	if err := s.Delete(ctx, url); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "booths/3/a.webp")); !os.IsNotExist(err) {
		t.Error("消えていない")
	}
	if err := s.Delete(ctx, url); err != nil {
		t.Error("無くてもエラーにしない")
	}
}

func TestLocal_自分の置き場所だけ(t *testing.T) {
	s := NewLocal(t.TempDir(), "https://localhost:3000/dev-images")
	for url, want := range map[string]bool{
		"https://localhost:3000/dev-images/booths/3/a.webp": true,
		"https://example.com/a.png":                         false,
		"https://localhost:3000/dev-images-x/a.png":         false,
		"https://localhost:3000/dev-images/../secret":       false,
		"": false,
	} {
		if got := s.Owns(url); got != want {
			t.Errorf("%q: %v", url, got)
		}
	}
	// 自分のでない URL を消そうとしても何もしない
	if err := s.Delete(context.Background(), "https://example.com/a.png"); err != nil {
		t.Error(err)
	}
}
