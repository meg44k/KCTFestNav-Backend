package domain

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var classOrganizer = regexp.MustCompile(`^(\d+)\s*-\s*\d+$`)

// クラス展示の主催者("1-1" のようなクラス表記、学年 1〜5)か。フロントの parseGrade と同じ決まり
func IsClassOrganizer(organizer string) bool {
	m := classOrganizer.FindStringSubmatch(strings.TrimSpace(organizer))
	if m == nil {
		return false
	}
	g, _ := strconv.Atoi(m[1])
	return g >= 1 && g <= 5
}

// 10 分ごとのいいねの数
type LikeBucket struct {
	Start time.Time
	Count int
	Burst bool
}

// 管理画面に出す、ブースごとのいいね
type BoothLikes struct {
	BoothID   int
	Name      string
	Organizer string
	Total     int
	Buckets   []LikeBucket
	Burst     bool
}

const (
	burstMin   = 30
	burstRatio = 5
)

// 10 分の数が 30 件以上で、そのブースの 10 分ごとの数(0 を除く)の中央値の 5 倍以上なら ⚠
func MarkBursts(buckets []LikeBucket) []LikeBucket {
	out := make([]LikeBucket, len(buckets))
	copy(out, buckets)
	var nonzero []int
	for _, x := range out {
		if x.Count > 0 {
			nonzero = append(nonzero, x.Count)
		}
	}
	if len(nonzero) == 0 {
		return out
	}
	sort.Ints(nonzero)
	mid := len(nonzero) / 2
	median := float64(nonzero[mid])
	if len(nonzero)%2 == 0 {
		median = float64(nonzero[mid-1]+nonzero[mid]) / 2
	}
	for i := range out {
		out[i].Burst = out[i].Count >= burstMin && float64(out[i].Count) >= median*burstRatio
	}
	return out
}

type LikeRepository interface {
	Like(ctx context.Context, boothID int, voterID string, at time.Time) error
	Unlike(ctx context.Context, boothID int, voterID string) error
	Mine(ctx context.Context, voterID string) ([]int, error)
	// 全ブースのいいねの時刻(10 分ごとにまとめるのは usecase)
	AllLikeTimes(ctx context.Context) (map[int][]time.Time, error)
	// from 以上 to 未満を消し、記録を残す。消した件数を返す
	RemoveRange(ctx context.Context, boothID int, from, to time.Time, by string, at time.Time) (int, error)
	RemoveAll(ctx context.Context, by string, at time.Time) (int, error)
	// 回数の上限。超えていなければ true
	AllowNewVoter(ctx context.Context) (bool, error)
	AllowToggle(ctx context.Context, voterID string) (bool, error)
}
