package domain

import (
	"context"
	"errors"
	"strings"
)

// クラス展示、部活の出店などをあらわす型。
// 混雑度は専用メソッドから読み取り、書き込みを行う必要があります。
type Booth struct {
	ID               int              // ブースID
	Name             string           // ブース名
	Organizer        string           // ブースの主催者(ex. 1-1, 陸上部...)
	Detail           string           // ブースの説明
	congestionStatus CongestionStatus // 0: 空き 1: 少し混雑している 2: かなり混雑している]
	X                float32          // X座標
	Y                float32          // Y座標
	Z                float32          // Z座標
}

type CongestionStatus int8

const (
	BoothCongestionEmpty           CongestionStatus = 0
	BoothCongestionSlightlyCrowded CongestionStatus = 1
	BoothCongestionVeryCrowded     CongestionStatus = 2
)

// 新規ブース作成用コンストラクタ
// IDがDB側で採番されるため、デフォルトではID=0となっている
func NewBooth(
	name string,
	organizer string,
	detail string,
	X float32,
	Y float32,
	Z float32,

) (*Booth, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrNameRequired

	}
	return &Booth{
		ID:               0,
		Name:             name,
		Organizer:        organizer,
		Detail:           detail,
		congestionStatus: BoothCongestionEmpty,
		X:                X,
		Y:                Y,
		Z:                Z,
	}, nil
}

// DBからの復元用コンストラクタ
// IDがDB側から採択されたものがIDに入っている
func ReconstructBooth(
	id int,
	name string,
	organizer string,
	detail string,
	congestionStatus CongestionStatus,
	x float32,
	y float32,
	z float32,

) (*Booth, error) {
	return &Booth{
		ID:               id,
		Name:             name,
		Organizer:        organizer,
		Detail:           detail,
		congestionStatus: congestionStatus,
		X:                x,
		Y:                y,
		Z:                z,
	}, nil
}

type BoothRepository interface {
	Create(ctx context.Context, booth *Booth) error
	GetByID(ctx context.Context, id int) (*Booth, error)
	Update(ctx context.Context, booth *Booth) error
	UpdateCongestion(ctx context.Context, id int, congestionStatus CongestionStatus) error
	GetAll(ctx context.Context) ([]*Booth, error)
	Delete(ctx context.Context, id int) error
}

func (b *Booth) CongestionStatus() CongestionStatus {
	return b.congestionStatus
}

func (b *Booth) SetCongestionStatus(congestionStatus CongestionStatus) error {
	if err := ValidateCongestionLevel(congestionStatus); err != nil {
		return err
	}
	b.congestionStatus = congestionStatus
	return nil
}

func ValidateCongestionLevel(congestionLevel CongestionStatus) error {
	if 0 > congestionLevel || congestionLevel > 2 {
		return errors.New("Congestion level must be between 0 and 2")
	}
	return nil
}
