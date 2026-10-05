package domain

import (
	"context"
	"strings"
	"time"
)

// ステージイベントは セクション(Live1 など) → ブロック(時間帯) → 出演者(順番だけ) の 3 段。
// 出演者ごとの時刻は持たず、学生会がブロックの current_order を進めて「今だれか」を決める

type StageSection struct {
	ID        int
	Name      string // セクション名(Live1、癒し系ミュージシャン など)
	Location  string // 場所(第一体育館 など)。空でもよい
	SortOrder int    // 並び順。同じなら最初のブロックの開始時刻の順
	Blocks    []*StageBlock
}

type StageBlock struct {
	ID        int
	SectionID int
	StartTime time.Time
	EndTime   time.Time
	// 0 = まだ始まっていない、1〜出演者数 = その順番の出演者が演奏中、出演者数 + 1 = 終了
	CurrentOrder int
	Performers   []*Performer
}

type Performer struct {
	ID           int
	BlockID      int
	Name         string
	Detail       string
	ThumbnailURL string
	PerformOrder int // ブロックの中での出演順(1 から)
}

// 出演順の入れ替えの向き
type MoveDirection string

const (
	MoveUp   MoveDirection = "up"
	MoveDown MoveDirection = "down"
)

func NewStageSection(name, location string, sortOrder int) (*StageSection, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrNameRequired
	}
	return &StageSection{Name: name, Location: location, SortOrder: sortOrder}, nil
}

func NewStageBlock(sectionID int, start, end time.Time) (*StageBlock, error) {
	if !end.After(start) {
		return nil, ErrEndTimeAfterStartTime
	}
	return &StageBlock{SectionID: sectionID, StartTime: start, EndTime: end}, nil
}

func NewPerformer(blockID int, name, detail, thumbnailURL string) (*Performer, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrNameRequired
	}
	return &Performer{BlockID: blockID, Name: name, Detail: detail, ThumbnailURL: thumbnailURL}, nil
}

func ValidateMoveDirection(d MoveDirection) error {
	if d != MoveUp && d != MoveDown {
		return ErrInvalidDirection
	}
	return nil
}

// 終了時刻を過ぎても演奏中と出し続ける長さ。延びたライブに備え、押し忘れはこの後に消す
const NowPlayingGrace = 30 * time.Minute

// 来場者に「演奏中」と出すか。学生会が進めて current_order が出演者を指していれば、
// 開始時刻に関係なく出す(早めの開始や当日前の確認でも見えるように)。
// 押し忘れに備え、終了時刻から NowPlayingGrace たったら消す
func (b *StageBlock) NowPlaying(now time.Time) bool {
	if !now.Before(b.EndTime.Add(NowPlayingGrace)) {
		return false
	}
	return b.CurrentOrder >= 1 && b.CurrentOrder <= len(b.Performers)
}

type StageRepository interface {
	GetSchedule(ctx context.Context) ([]*StageSection, error)
	GetSection(ctx context.Context, id int) (*StageSection, error)
	GetBlock(ctx context.Context, id int) (*StageBlock, error)
	GetPerformer(ctx context.Context, id int) (*Performer, error)
	CreateSection(ctx context.Context, s *StageSection) (int, error)
	UpdateSection(ctx context.Context, s *StageSection) error
	DeleteSection(ctx context.Context, id int) error
	CreateBlock(ctx context.Context, b *StageBlock) (int, error)
	UpdateBlock(ctx context.Context, b *StageBlock) error
	DeleteBlock(ctx context.Context, id int) error
	CreatePerformer(ctx context.Context, p *Performer) (int, error)
	UpdatePerformer(ctx context.Context, p *Performer) error
	DeletePerformer(ctx context.Context, id int) error
	MovePerformer(ctx context.Context, id int, d MoveDirection) error
	AdvanceBlock(ctx context.Context, id int) error
	RewindBlock(ctx context.Context, id int) error
}
