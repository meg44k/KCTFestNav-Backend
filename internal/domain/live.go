package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

type LiveStatus int8

type Live struct {
	ID            int        // ライブID
	Name          string     // バンドorライブ名
	Detail        string     // 説明文
	ThumbnailURL  string     // サムネイル画像URL
	StartTime     time.Time  // 始まる時間
	EndTime       time.Time  // 終わる時間
	sessionNumber int8       // 何番目にライブがあるか ここに関しては少し変えるかも
	status        LiveStatus // ライブの状況 0: まだ始まっていない 1: 開演中 2: 終了済み 学生会員が手動で状況を変える
}

const (
	LiveStatusUpcoming LiveStatus = 0 // 開演前
	LiveStatusOngoing  LiveStatus = 1 // 公演中
	LiveStatusFinished LiveStatus = 2 // 終了済
)

var ErrLiveNameRequired = errors.New("name is required")
var ErrEndTimeAfterStartTime = errors.New("end time must be after start time")
var ErrSessionNumberLessThanOne = errors.New("session number must be at least 1")
var ErrInvalidLiveStatus = errors.New("invalid live status value")

// 新規作成用コンストラクタ
// DBでIDが採番されるためデフォルトではID=0
func NewLive(
	name string,
	detail string,
	thumbnailURL string,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8,
) (*Live, error) {
	err := validateLive(name, startTime, endTime, sessionNumber)
	if err != nil {
		return nil, err
	}
	return &Live{
		ID:            0,
		Name:          name,
		Detail:        detail,
		ThumbnailURL:  thumbnailURL,
		StartTime:     startTime,
		EndTime:       endTime,
		sessionNumber: sessionNumber,
		status:        LiveStatusUpcoming,
	}, nil
}

// DBからの復元用コンストラクタ
func ReconstructLive(
	id int,
	name string,
	detail string,
	thumbnailURL string,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8,
	status LiveStatus,
) (*Live, error) {
	err := validateLive(name, startTime, endTime, sessionNumber)
	if err != nil {
		return nil, err
	}
	return &Live{
		ID:            id,
		Name:          name,
		Detail:        detail,
		ThumbnailURL:  thumbnailURL,
		StartTime:     startTime,
		EndTime:       endTime,
		sessionNumber: sessionNumber,
		status:        status,
	}, nil
}

type LiveRepository interface {
	Create(ctx context.Context, live *Live) error
	Update(ctx context.Context, live *Live) error
	Delete(ctx context.Context, id int) error
	GetByID(ctx context.Context, id int) (*Live, error)
	GetAll(ctx context.Context) ([]*Live, error)
}

func validateLive(
	name string,
	startTime time.Time,
	endTime time.Time,
	sessionNumber int8) error {
	if strings.TrimSpace(name) == "" {
		return ErrLiveNameRequired
	}

	if !endTime.After(startTime) {
		return ErrEndTimeAfterStartTime
	}

	if sessionNumber < 1 {
		return ErrSessionNumberLessThanOne
	}

	return nil
}

func (l *Live) SetStatus(status LiveStatus) error {
	switch status {
	case LiveStatusUpcoming, LiveStatusOngoing, LiveStatusFinished:
		l.status = status
		return nil
	default:
		return ErrInvalidLiveStatus
	}
}

func (l *Live) Status() LiveStatus {
	return l.status
}

func (l *Live) SetSessionNumber(sessionNumber int8) error {
	if sessionNumber < 1 {
		return ErrSessionNumberLessThanOne
	}
	l.sessionNumber = sessionNumber
	return nil
}

func (l *Live) SessionNumber() int8 {
	return l.sessionNumber
}
