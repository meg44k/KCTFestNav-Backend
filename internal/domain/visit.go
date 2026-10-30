package domain

import (
	"context"
	"errors"
)

// 訪問者の番号(ブラウザが作る UUID)の形がおかしい
var ErrInvalidVisitor = errors.New("visitor must be a UUID")

// 日ごとの訪問者の数。同じ番号は 1 日 1 人と数える(Redis の HyperLogLog なので、数はおおよそ)
type VisitRepository interface {
	// その日(2006-01-02、日本時間)にその番号が来たと記録する
	Record(ctx context.Context, day, visitor string) error
	// 日ごとの数と、それらの日を通した重なりなしの数
	Count(ctx context.Context, days []string) (perDay map[string]int64, total int64, err error)
}

// 1 日分の訪問者の数
type VisitDay struct {
	Day      string `json:"day"`
	Visitors int64  `json:"visitors"`
}

// 直近の日ごとの数(古い日から順)と、それらの日を通した重なりなしの合計
type VisitSummary struct {
	Days  []VisitDay `json:"days"`
	Total int64      `json:"total"`
}
