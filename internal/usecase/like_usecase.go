package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/meg44k/KCTFestNav-Backend/internal/auth"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// いいねを 10 分ごとにまとめる
const likeBucket = 10 * time.Minute

type LikeUsecase struct {
	likes       domain.LikeRepository
	booths      domain.BoothRepository
	voterSecret []byte
	now         func() time.Time
}

func NewLikeUsecase(likes domain.LikeRepository, booths domain.BoothRepository, voterSecret []byte, now func() time.Time) *LikeUsecase {
	return &LikeUsecase{likes: likes, booths: booths, voterSecret: voterSecret, now: now}
}

// 新しい投票者番号を作る。サーバー全体の上限を超えたら ErrTooMany
func (u *LikeUsecase) IssueVoter(ctx context.Context) (string, error) {
	ok, err := u.likes.AllowNewVoter(ctx)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrTooMany
	}
	token, _ := auth.NewVoterToken(u.voterSecret)
	return token, nil
}

// その番号で押したブース。番号が不正なら空
func (u *LikeUsecase) Mine(ctx context.Context, token string) ([]int, error) {
	voter, err := auth.ParseVoterToken(token, u.voterSecret)
	if err != nil {
		return []int{}, nil
	}
	return u.likes.Mine(ctx, voter)
}

func (u *LikeUsecase) Like(ctx context.Context, token string, boothID int) error {
	voter, err := u.voterFor(ctx, token)
	if err != nil {
		return err
	}
	booth, err := u.booths.GetByID(ctx, boothID)
	if err != nil {
		return err
	}
	if !domain.IsClassOrganizer(booth.Organizer) {
		return domain.ErrNotClassBooth
	}
	return u.likes.Like(ctx, boothID, voter, u.now())
}

func (u *LikeUsecase) Unlike(ctx context.Context, token string, boothID int) error {
	voter, err := u.voterFor(ctx, token)
	if err != nil {
		return err
	}
	return u.likes.Unlike(ctx, boothID, voter)
}

// 番号を確かめて、切り替えの上限を数える
func (u *LikeUsecase) voterFor(ctx context.Context, token string) (string, error) {
	voter, err := auth.ParseVoterToken(token, u.voterSecret)
	if err != nil {
		return "", ErrUnauthorized
	}
	ok, err := u.likes.AllowToggle(ctx, voter)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrTooMany
	}
	return voter, nil
}

// クラス展示ごとの数と 10 分ごとの推移(多い順)。管理者と学生会だけ
func (u *LikeUsecase) Summary(ctx context.Context) ([]domain.BoothLikes, error) {
	if !hasRole(ctx, domain.RoleAdmin, domain.RoleGakuseikai) {
		return nil, ErrForbidden
	}
	booths, err := u.booths.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	times, err := u.likes.AllLikeTimes(ctx)
	if err != nil {
		return nil, err
	}

	// 全ブースで同じ時間の軸にする(グラフを並べて比べられるように)
	var first, last time.Time
	for _, ts := range times {
		for _, t := range ts {
			t = t.UTC().Truncate(likeBucket)
			if first.IsZero() || t.Before(first) {
				first = t
			}
			if t.After(last) {
				last = t
			}
		}
	}

	out := []domain.BoothLikes{}
	for _, b := range booths {
		if !domain.IsClassOrganizer(b.Organizer) {
			continue
		}
		var buckets []domain.LikeBucket
		if !first.IsZero() {
			for s := first; !s.After(last); s = s.Add(likeBucket) {
				buckets = append(buckets, domain.LikeBucket{Start: s})
			}
			for _, t := range times[b.ID] {
				i := int(t.UTC().Truncate(likeBucket).Sub(first) / likeBucket)
				buckets[i].Count++
			}
		}
		buckets = domain.MarkBursts(buckets)
		bl := domain.BoothLikes{BoothID: b.ID, Name: b.Name, Organizer: b.Organizer, Total: len(times[b.ID]), Buckets: buckets}
		for _, x := range buckets {
			bl.Burst = bl.Burst || x.Burst
		}
		out = append(out, bl)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Total > out[j].Total })
	return out, nil
}

// そのブースの from 以上 to 未満のいいねを消す。管理者と学生会だけ
func (u *LikeUsecase) RemoveRange(ctx context.Context, boothID int, from, to time.Time) (int, error) {
	if !hasRole(ctx, domain.RoleAdmin, domain.RoleGakuseikai) {
		return 0, ErrForbidden
	}
	if !from.Before(to) {
		return 0, domain.ErrInvalidRange
	}
	return u.likes.RemoveRange(ctx, boothID, from, to, requestUserID(ctx), u.now())
}

// 全部消す(文化祭の前の試しの票を消す用)。管理者だけ
func (u *LikeUsecase) RemoveAll(ctx context.Context) (int, error) {
	if !isAdmin(ctx) {
		return 0, ErrForbidden
	}
	return u.likes.RemoveAll(ctx, requestUserID(ctx), u.now())
}

func requestUserID(ctx context.Context) string {
	reqUser, _ := ctx.Value(ContextRequestUserKey).(RequestUser)
	return reqUser.ID.String()
}
