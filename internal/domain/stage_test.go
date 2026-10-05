package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewStageSection(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"正常系", "Live1", nil},
		{"異常系：名前が空", "", ErrNameRequired},
		{"異常系：半角スペースのみ", " ", ErrNameRequired},
		{"異常系：全角スペースのみ", "　", ErrNameRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewStageSection(tt.input, "第一体育館", 2)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if err == nil && (s.Name != tt.input || s.Location != "第一体育館" || s.SortOrder != 2) {
				t.Fatalf("値が入っていない: %+v", s)
			}
		})
	}
}

func TestNewStageBlock(t *testing.T) {
	base := time.Date(2026, 10, 31, 13, 0, 0, 0, time.UTC)
	if _, err := NewStageBlock(1, base, base.Add(time.Hour)); err != nil {
		t.Fatalf("正常系でエラー: %v", err)
	}
	if _, err := NewStageBlock(1, base, base); !errors.Is(err, ErrEndTimeAfterStartTime) {
		t.Fatalf("同時刻: err = %v", err)
	}
	if _, err := NewStageBlock(1, base, base.Add(-time.Minute)); !errors.Is(err, ErrEndTimeAfterStartTime) {
		t.Fatalf("終了が前: err = %v", err)
	}
}

func TestNewPerformer(t *testing.T) {
	p, err := NewPerformer(3, "バンドA", "紹介", "https://example.com/a.jpg")
	if err != nil || p.BlockID != 3 || p.Name != "バンドA" || p.Detail != "紹介" || p.ThumbnailURL != "https://example.com/a.jpg" {
		t.Fatalf("正常系: %+v %v", p, err)
	}
	if _, err := NewPerformer(3, "　", "", ""); !errors.Is(err, ErrNameRequired) {
		t.Fatalf("名前なし: err = %v", err)
	}
}

func TestValidateMoveDirection(t *testing.T) {
	for _, d := range []MoveDirection{MoveUp, MoveDown} {
		if err := ValidateMoveDirection(d); err != nil {
			t.Fatalf("%s: %v", d, err)
		}
	}
	for _, d := range []MoveDirection{"", "left", "UP"} {
		if err := ValidateMoveDirection(d); !errors.Is(err, ErrInvalidDirection) {
			t.Fatalf("%q: err = %v", d, err)
		}
	}
}

func TestStageBlock_NowPlaying(t *testing.T) {
	start := time.Date(2026, 10, 31, 4, 0, 0, 0, time.UTC) // 13:00 JST
	end := start.Add(50 * time.Minute)
	three := []*Performer{{PerformOrder: 1}, {PerformOrder: 2}, {PerformOrder: 3}}
	in := start.Add(10 * time.Minute)

	tests := []struct {
		name       string
		now        time.Time
		current    int
		performers []*Performer
		want       bool
	}{
		{"時間内で1組目", in, 1, three, true},
		{"時間内で最後", in, 3, three, true},
		{"時間内でもまだ始まっていない", in, 0, three, false},
		{"時間内でも終了", in, 4, three, false},
		{"開始ちょうどは時間内", start, 2, three, true},
		{"終了ちょうどは時間外", end, 2, three, false},
		{"開始前", start.Add(-time.Minute), 1, three, false},
		{"終了後(押し忘れ)", end.Add(time.Minute), 2, three, false},
		{"出演者0人", in, 1, nil, false},
		{"JSTで渡しても同じ瞬間なら時間内", in.In(time.FixedZone("JST", 9*3600)), 1, three, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &StageBlock{StartTime: start, EndTime: end, CurrentOrder: tt.current, Performers: tt.performers}
			if got := b.NowPlaying(tt.now); got != tt.want {
				t.Fatalf("NowPlaying = %v, want %v", got, tt.want)
			}
		})
	}
}
