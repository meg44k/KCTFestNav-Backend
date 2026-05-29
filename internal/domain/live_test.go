package domain

import (
	"testing"
	"time"
)

func TestNewLive(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)

	tests := []struct {
		name        string
		inputName   string
		inputDetail string
		wantErr     bool
	}{
		{"正常系", "testName", "詳細", false},
		{"異常系：名前が空文字", "", "詳細", true},
		{"異常系：名前が半角スペースのみ", " ", "詳細", true},
		{"異常系：名前が全角スペースのみ", "　", "詳細", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			live, err := NewLive(tt.inputName, tt.inputDetail, "https://example.com/thumb.png", baseTime, baseTime, 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewLive() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if live.ID != 0 {
					t.Errorf("NewLive() ID = %v, want 0", live.ID)
				}
				if live.Name != tt.inputName {
					t.Errorf("NewLive() Name = %v, want %v", live.Name, tt.inputName)
				}
				if live.Status() != LiveStatusUpcoming {
					t.Errorf("NewLive() Status = %v, want %v", live.Status(), LiveStatusUpcoming)
				}
			}
		})
	}
}

func TestReconstructLive(t *testing.T) {
	baseTime := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local)

	tests := []struct {
		name        string
		inputID     int
		inputName   string
		inputDetail string
		inputStatus LiveStatus
		wantErr     bool
	}{
		{"正常系", 99, "testName", "詳細", LiveStatusOngoing, false},
		{"異常系：名前が空文字", 99, "", "詳細", LiveStatusOngoing, true},
		{"異常系：名前が半角スペースのみ", 99, " ", "詳細", LiveStatusOngoing, true},
		{"異常系：名前が全角スペースのみ", 99, "　", "詳細", LiveStatusOngoing, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			live, err := ReconstructLive(tt.inputID, tt.inputName, tt.inputDetail, "https://example.com/thumb.png", baseTime, baseTime, 1, tt.inputStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReconstructLive() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if live.ID != tt.inputID {
					t.Errorf("ReconstructLive() ID = %v, want %v", live.ID, tt.inputID)
				}
				if live.Name != tt.inputName {
					t.Errorf("ReconstructLive() Name = %v, want %v", live.Name, tt.inputName)
				}
				if live.Status() != tt.inputStatus {
					t.Errorf("ReconstructLive() Status = %v, want %v", live.Status(), tt.inputStatus)
				}
			}
		})
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

func TestLive_SetSessionNumber(t *testing.T) {
	tests := []struct {
		name          string
		sessionNumber int8
		wantErr       bool
	}{
		{"正常系: 正しい番号", 1, false},
		{"正常系: 正しい番号", 5, false},
		{"異常系: 0は無効", 0, true},
		{"異常系: マイナスは無効", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Live{}
			err := l.SetSessionNumber(tt.sessionNumber)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSessionNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && l.SessionNumber() != tt.sessionNumber {
				t.Errorf("SetSessionNumber() sessionNumberが更新されていません: got %v, want %v", l.SessionNumber(), tt.sessionNumber)
			}
		})
	}
}
