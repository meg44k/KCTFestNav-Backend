package domain

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseImageTarget(t *testing.T) {
	ok := map[string]ImageTarget{"booth:3": {"booth", 3}, "performer:12": {"performer", 12}}
	for in, want := range ok {
		got, err := ParseImageTarget(in)
		if err != nil || got != want {
			t.Errorf("%q: %v %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "booth", "booth:", "booth:0", "booth:-1", "booth:x", "user:1", "booth:1:2"} {
		if _, err := ParseImageTarget(bad); err != ErrInvalidImageTarget {
			t.Errorf("%q は弾くはず", bad)
		}
	}
}

func TestCheckImage(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 32))
	jpg := []byte("\xff\xd8\xff\xe0" + strings.Repeat("\x00", 32))
	webp := []byte("RIFF\x00\x00\x00\x00WEBPVP8 " + strings.Repeat("\x00", 32))
	for data, ext := range map[string]string{string(png): "png", string(jpg): "jpg", string(webp): "webp"} {
		got, ct, err := CheckImage([]byte(data))
		if err != nil || got != ext || !strings.HasPrefix(ct, "image/") {
			t.Errorf("%s: %s %s %v", ext, got, ct, err)
		}
	}
	big := append(append([]byte{}, png...), make([]byte, 2<<20)...)
	for _, bad := range [][]byte{nil, []byte("<html>"), []byte("GIF89a......"), big} {
		if _, _, err := CheckImage(bad); err != ErrInvalidImage {
			t.Errorf("弾くはず: %d bytes", len(bad))
		}
	}
}

func TestImageKey(t *testing.T) {
	k := ImageKey(ImageTarget{"booth", 3}, "webp")
	if !regexp.MustCompile(`^booths/3/[0-9a-f-]{36}\.webp$`).MatchString(k) {
		t.Error(k)
	}
	if !strings.HasPrefix(ImageKey(ImageTarget{"performer", 12}, "jpg"), "performers/12/") {
		t.Error("performers/")
	}
}
