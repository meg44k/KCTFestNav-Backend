package domain

import (
	"testing"
	"time"
)

func TestNewLive(t *testing.T) {
	// テスト用のLiveモデル（モックデータ）を作成
	mockLive := &Live{
		ID:            1,
		Name:          "testName",
		Detail:        "とてもかっこいいバンドです",
		ThumbnailURL:  "https://example.com/thumbnail.png",
		StartTime:     time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local), // 13:00開始
		EndTime:       time.Date(2026, time.May, 29, 14, 0, 0, 0, time.Local), // 14:00終了
		SessionNumber: 1,
	}

	// 例: データが正しく入っているかの簡単な確認
	if mockLive.Name != "testName" {
		t.Errorf("期待したバンド名は 'testName' ですが、 '%s' でした", mockLive.Name)
	}

	err := mockLive.SetStatus(LiveStatusFinished)
	if err != nil {
		t.Errorf("正しいステータスなのにエラーが返されました: %v", err)
	}
	if mockLive.Status() != LiveStatusFinished {
		t.Error("ステータスが正しく更新されていません")
	}

	err = mockLive.SetStatus(LiveStatusOngoing)
	if err != nil {
		t.Errorf("正しいステータスなのにエラーが返されました: %v", err)
	}
	if mockLive.Status() != LiveStatusOngoing {
		t.Error("ステータスが正しく更新されていません")
	}

	const FakeLiveStatus LiveStatus = 4
	err = mockLive.SetStatus(FakeLiveStatus)
	if err == nil {
		t.Error("不正なステータスを設定したのにエラーが返されませんでした")
	}
	if mockLive.Status() == FakeLiveStatus {
		t.Errorf("不正な値 (%d) がセットされてしまいました", FakeLiveStatus)
	}

}

func TestReconstructLive(t *testing.T) {
	// テスト用のLiveモデル（モックデータ）を作る

	mockLiveStruct := &Live{
		ID:            1,
		Name:          "testName",
		Detail:        "とてもかっこいいバンドです",
		ThumbnailURL:  "https://example.com/thumbnail.png",
		StartTime:     time.Date(2026, time.May, 29, 13, 0, 0, 0, time.Local), // 13:00開始
		EndTime:       time.Date(2026, time.May, 29, 14, 0, 0, 0, time.Local), // 14:00終了
		SessionNumber: 1,
		status:        LiveStatusUpcoming,
	}

	mockLive, err := ReconstructLive(mockLiveStruct.ID, mockLiveStruct.Name, mockLiveStruct.Detail, mockLiveStruct.ThumbnailURL, mockLiveStruct.StartTime, mockLiveStruct.EndTime, mockLiveStruct.SessionNumber, mockLiveStruct.status)

	// 例: データが正しく入っているかの簡単な確認
	if mockLive.Name != "testName" {
		t.Errorf("期待したバンド名は 'testName' ですが、 '%s' でした", mockLive.Name)
	}

	err = mockLive.SetStatus(LiveStatusFinished)
	if err != nil {
		t.Errorf("正しいステータスなのにエラーが返されました: %v", err)
	}
	if mockLive.Status() != LiveStatusFinished {
		t.Error("ステータスが正しく更新されていません")
	}

	err = mockLive.SetStatus(LiveStatusOngoing)
	if err != nil {
		t.Errorf("正しいステータスなのにエラーが返されました: %v", err)
	}
	if mockLive.Status() != LiveStatusOngoing {
		t.Error("ステータスが正しく更新されていません")
	}

	const FakeLiveStatus LiveStatus = 4
	err = mockLive.SetStatus(FakeLiveStatus)
	if err == nil {
		t.Error("不正なステータスを設定したのにエラーが返されませんでした")
	}
	if mockLive.Status() == FakeLiveStatus {
		t.Errorf("不正な値 (%d) がセットされてしまいました", FakeLiveStatus)
	}

}
