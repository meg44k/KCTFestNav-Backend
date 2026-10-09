package domain

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// 画像を上げる先。booth:3 なら 3 番のブースの写真
type ImageTarget struct {
	Kind string // "booth" / "performer"
	ID   int
}

// 1 枚の上限。ブラウザで 1080px に縮めてから送るので、ふつうは 200KB ほど
const maxImageBytes = 2 << 20

func ParseImageTarget(s string) (ImageTarget, error) {
	kind, idStr, ok := strings.Cut(s, ":")
	if !ok || (kind != "booth" && kind != "performer") {
		return ImageTarget{}, ErrInvalidImageTarget
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return ImageTarget{}, ErrInvalidImageTarget
	}
	return ImageTarget{Kind: kind, ID: id}, nil
}

// 中身が JPEG・PNG・WebP で 2MB 以内か確かめる(ヘッダーや拡張子ではなく中身を見る)
func CheckImage(data []byte) (ext string, contentType string, err error) {
	if len(data) == 0 || len(data) > maxImageBytes {
		return "", "", ErrInvalidImage
	}
	switch ct := http.DetectContentType(data); ct {
	case "image/jpeg":
		return "jpg", ct, nil
	case "image/png":
		return "png", ct, nil
	case "image/webp":
		return "webp", ct, nil
	default:
		return "", "", ErrInvalidImage
	}
}

// 置き場所の名前。booths/3/<UUID>.webp
func ImageKey(t ImageTarget, ext string) string {
	return t.Kind + "s/" + strconv.Itoa(t.ID) + "/" + uuid.NewString() + "." + ext
}

// 画像の置き場所(本番は R2、手元はフォルダ)
type ImageStore interface {
	// 置いて、来場者に配る URL を返す
	Put(ctx context.Context, key, contentType string, data []byte) (string, error)
	// 自分の置き場所の URL なら消す(無くてもエラーにしない)。それ以外は何もしない
	Delete(ctx context.Context, url string) error
	// この URL が自分の置き場所のものか
	Owns(url string) bool
}
