package domain

import (
	"testing"
	"time"
)

func TestNewLive(t *testing.T) {
	startTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)
	endTime := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.Local)

	// 実際の関数（コンストラクタ）を呼び出してテストする
	live, err := NewLive("testName", "詳細", "https://example.com/thumb.png", startTime, endTime, 1)

	if err != nil {
		t.Fatalf("NewLive() で予期せぬエラーが発生しました: %v", err)
	}

	// 新規作成時は ID が 0 であることの確認
	if live.ID != 0 {
		t.Errorf("NewLive() ID = %v, want 0 (自動採番される前は0であるべき)", live.ID)
	}
	// 値が正しく入っているかの確認
	if live.Name != "testName" {
		t.Errorf("NewLive() Name = %v, want testName", live.Name)
	}
	// 新規作成時は初期ステータスが必ず LiveStatusUpcoming であることの確認
	if live.Status() != LiveStatusUpcoming {
		t.Errorf("NewLive() Status = %v, want %v (初期状態は未開演であるべき)", live.Status(), LiveStatusUpcoming)
	}
}

func TestReconstructLive(t *testing.T) {
	startTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)
	endTime := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.Local)

	// DBから取得したという想定で、既存のID(99)やステータス(LiveStatusOngoing)を渡して復元する
	live, err := ReconstructLive(99, "testName", "詳細", "https://example.com/thumb.png", startTime, endTime, 1, LiveStatusOngoing)

	if err != nil {
		t.Fatalf("ReconstructLive() で予期せぬエラーが発生しました: %v", err)
	}

	// 渡したIDがそのまま維持されているかの確認
	if live.ID != 99 {
		t.Errorf("ReconstructLive() ID = %v, want 99", live.ID)
	}
	// 値が正しく入っているかの確認
	if live.Name != "testName" {
		t.Errorf("ReconstructLive() Name = %v, want testName", live.Name)
	}
	// 渡したステータスがそのまま維持されているかの確認
	if live.Status() != LiveStatusOngoing {
		t.Errorf("ReconstructLive() Status = %v, want %v", live.Status(), LiveStatusOngoing)
	}
}

func TestLive_SetStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  LiveStatus
		wantErr bool
	}{
		{"Upcoming", LiveStatusUpcoming, false},
		{"Ongoing", LiveStatusOngoing, false},
		{"Finished", LiveStatusFinished, false},
		{"InvalidValue", LiveStatus(4), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Live{}
			err := l.SetStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && l.Status() != tt.status {
				t.Errorf("ステータスが更新されていません:got %v, want %v", l.Status(), tt.status)
			}
		})
	}
}
