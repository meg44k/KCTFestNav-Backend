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
	Location         string
	X                float32 // X座標
	Y                float32 // Y座標
	Z                float32 // Z座標
	Latitude         float64 // 緯度
	Longitude        float64 // 経度
}

type CongestionStatus int8

const (
	BoothCongestionEmpty           CongestionStatus = 0
	BoothCongestionSlightlyCrowded CongestionStatus = 1
	BoothCongestionVeryCrowded     CongestionStatus = 2
)

type BoothParams struct {
	Name      string
	Organizer string
	Detail    string
	Location  string
	X         float32
	Y         float32
	Z         float32
	Latitude  float64
	Longitude float64
}

// 新規ブース作成用コンストラクタ
// IDがDB側で採番されるため、デフォルトではID=0となっている
func NewBooth(p BoothParams) (*Booth, error) {
	if strings.TrimSpace(p.Name) == "" {
		return nil, ErrNameRequired

	}
	return &Booth{
		ID:               0,
		Name:             p.Name,
		Organizer:        p.Organizer,
		Detail:           p.Detail,
		Location:         p.Location,
		congestionStatus: BoothCongestionEmpty,
		X:                p.X,
		Y:                p.Y,
		Z:                p.Z,
		Latitude:         p.Latitude,
		Longitude:        p.Longitude,
	}, nil
}

// DBからの復元用コンストラクタ
// IDがDB側から採択されたものがIDに入っている
func ReconstructBooth(
	id int,
	congestionStatus CongestionStatus,
	p BoothParams,
) (*Booth, error) {
	return &Booth{
		ID:               id,
		Name:             p.Name,
		Organizer:        p.Organizer,
		Detail:           p.Detail,
		Location:         p.Location,
		congestionStatus: congestionStatus,
		X:                p.X,
		Y:                p.Y,
		Z:                p.Z,
		Latitude:         p.Latitude,
		Longitude:        p.Longitude,
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
