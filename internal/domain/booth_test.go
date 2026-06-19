package domain

import (
	"testing"
)

func TestBoothCongestionStatus(t *testing.T) {
	booth, _ := NewBooth("テストブース", "1-1", "テストブースです", 0, 0, 0)
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
