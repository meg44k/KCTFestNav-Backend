package domain

import (
	"testing"
	"time"
)

func TestIsClassOrganizer(t *testing.T) {
	for in, want := range map[string]bool{
		"1-1": true, "5-4": true, " 3 - 2 ": true,
		"0-1": false, "6-1": false, "天文部": false, "1-": false, "": false, "12-1": false,
	} {
		if got := IsClassOrganizer(in); got != want {
			t.Errorf("%q: got %v", in, got)
		}
	}
}

func bucket(count int) LikeBucket { return LikeBucket{Start: time.Time{}, Count: count} }

func TestMarkBursts(t *testing.T) {
	// 0 を除くと [3,4,5,40,4] → 中央値 4。40 は 5 倍以上かつ 30 以上 → ⚠
	got := MarkBursts([]LikeBucket{bucket(3), bucket(0), bucket(4), bucket(5), bucket(40), bucket(4)})
	want := []bool{false, false, false, false, true, false}
	for i := range got {
		if got[i].Burst != want[i] {
			t.Errorf("%d: got %v", i, got[i].Burst)
		}
	}
}

func TestMarkBursts_30件未満は出さない(t *testing.T) {
	got := MarkBursts([]LikeBucket{bucket(1), bucket(1), bucket(1), bucket(29)})
	if got[3].Burst {
		t.Error("30 件未満は ⚠ にしない")
	}
}

func TestMarkBursts_5倍未満は出さない(t *testing.T) {
	got := MarkBursts([]LikeBucket{bucket(10), bucket(10), bucket(10), bucket(49)})
	if got[3].Burst {
		t.Error("5 倍未満は ⚠ にしない")
	}
}

func TestMarkBursts_元を書き換えない(t *testing.T) {
	in := []LikeBucket{bucket(1), bucket(1), bucket(1), bucket(40)}
	MarkBursts(in)
	if in[3].Burst {
		t.Error("渡したものは変えない")
	}
}

func TestMarkBursts_空(t *testing.T) {
	if len(MarkBursts(nil)) != 0 {
		t.Error("空は空")
	}
}
