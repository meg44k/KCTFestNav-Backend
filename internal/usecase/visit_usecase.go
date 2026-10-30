package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 来場者は日本の文化祭なので、日の区切りは日本時間
var jst = time.FixedZone("JST", 9*3600)

// 何人のブラウザが来たかを日ごとに数える
type VisitUsecase struct {
	visits domain.VisitRepository
	now    func() time.Time
}

func NewVisitUsecase(visits domain.VisitRepository, now func() time.Time) *VisitUsecase {
	return &VisitUsecase{visits: visits, now: now}
}

func dayOf(t time.Time) string { return t.In(jst).Format("2006-01-02") }

// 今日(日本時間)その番号が来たと記録する。日付はブラウザの言う日ではなく、ここで決める
func (u *VisitUsecase) Record(ctx context.Context, visitor string) error {
	id, err := uuid.Parse(visitor)
	if err != nil || id.String() != visitor {
		return domain.ErrInvalidVisitor
	}
	return u.visits.Record(ctx, dayOf(u.now()), visitor)
}

// 今日までの days 日分(古い日から順)と、その間の重なりなしの合計
func (u *VisitUsecase) Count(ctx context.Context, days int) (domain.VisitSummary, error) {
	if days < 1 || days > 60 {
		return domain.VisitSummary{}, domain.ErrInvalidRange
	}
	today := u.now()
	keys := make([]string, days)
	for i := range days {
		keys[i] = dayOf(today.AddDate(0, 0, i-days+1))
	}
	per, total, err := u.visits.Count(ctx, keys)
	if err != nil {
		return domain.VisitSummary{}, err
	}
	out := domain.VisitSummary{Days: make([]domain.VisitDay, days), Total: total}
	for i, k := range keys {
		out.Days[i] = domain.VisitDay{Day: k, Visitors: per[k]}
	}
	return out, nil
}
