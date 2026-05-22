package domain

import (
	"context"
	"time"
)

type Live struct {
	ID string; // ライブID
	Name string; // バンドorライブ名
	Detail string; // 説明文
	Thumbnail []byte; // サムネイル画像 画像パスに変えるかも
	StartTime time.Time; // 始まる時間
	EndTime time.Time; // 終わる時間
	SessionNumber int8; // 何番目にライブがあるか ここに関しては少し変えるかも
	Status int8; // ライブの状況 0: まだ始まっていない 1: 開演中 2: 終了済み
}

type LiveStatus int8

const(
	LiveStatusUpcoming LiveStatus = 0;
	LiveStatusOngoing LiveStatus = 1;
	LiveStatusFinished LiveStatus = 2;
)

func NewLive(
	id string,
	name string,
	detail string,
	thumbnail []byte,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8,
	status int8) (*Live, error) {
		// TODO: 実際の実装をここに書く
	return &Live{
		ID: id,
		Name: name,
		Detail: detail,
		Thumbnail: thumbnail,
		StartTime: startTime,
		EndTime: endTime,
		SessionNumber: sessionNumber,
		Status: status,
	}, nil
}

type LiveRepository interface {
	Create(ctx context.Context, live *Live) error
	Update(ctx context.Context, live *Live) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Live, error)
	GetAll(ctx context.Context) ([]*Live, error)
} 
