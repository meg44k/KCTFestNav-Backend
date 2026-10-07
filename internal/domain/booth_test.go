package domain

import (
	"errors"
	"testing"
)

func TestBoothCongestionStatus(t *testing.T) {
	booth, _ := NewBooth(BoothParams{
		Name:      "テストブース",
		Organizer: "1-1",
		Detail:    "テストブースです",
	})
	t.Run("正常値のセットチェック", func(t *testing.T) {
		booth.SetCongestionStatus(1)
		result := booth.CongestionStatus()
		expected := CongestionStatus(1)
		if result != expected {
			t.Error("混雑度が正しくセットできてないよ")
		}
	})

	t.Run("異常値のセットバリデーションチェック", func(t *testing.T) {
		err := booth.SetCongestionStatus(3)
		if err == nil {
			t.Error("混雑度のバリデーションがうまくいってないよ")
		}
	})
}

func TestBoothFloor(t *testing.T) {
	t.Run("階を持てる(0 は屋外)", func(t *testing.T) {
		b, err := NewBooth(BoothParams{Name: "3-1 展示", Floor: 2})
		if err != nil || b.Floor != 2 {
			t.Fatalf("got %v, %v", b, err)
		}
		b, err = NewBooth(BoothParams{Name: "バザー"})
		if err != nil || b.Floor != 0 {
			t.Fatalf("got %v, %v", b, err)
		}
	})

	t.Run("負の階は作成・復元とも ErrInvalidFloor", func(t *testing.T) {
		if _, err := NewBooth(BoothParams{Name: "x", Floor: -1}); !errors.Is(err, ErrInvalidFloor) {
			t.Fatalf("NewBooth: %v", err)
		}
		if _, err := ReconstructBooth(1, BoothCongestionEmpty, BoothParams{Name: "x", Floor: -1}); !errors.Is(err, ErrInvalidFloor) {
			t.Fatalf("ReconstructBooth: %v", err)
		}
	})
}
