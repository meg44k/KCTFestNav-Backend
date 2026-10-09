package usecase

import (
	"context"
	"log"
	"strings"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// ブース・出演者の写真を上げる。権限はそれぞれを更新できる人と同じ
type ImageUsecase struct {
	store  domain.ImageStore
	booths domain.BoothRepository
	stage  domain.StageRepository
}

func NewImageUsecase(store domain.ImageStore, booths domain.BoothRepository, stage domain.StageRepository) *ImageUsecase {
	return &ImageUsecase{store: store, booths: booths, stage: stage}
}

// 確かめて置き、来場者に配る URL を返す。ブースや出演者の画像はまだ変えない(フォームの保存で変わる)
func (u *ImageUsecase) Upload(ctx context.Context, target string, data []byte) (string, error) {
	t, err := domain.ParseImageTarget(target)
	if err != nil {
		return "", err
	}
	if err := u.canEdit(ctx, t); err != nil {
		return "", err
	}
	ext, contentType, err := domain.CheckImage(data)
	if err != nil {
		return "", err
	}
	return u.store.Put(ctx, domain.ImageKey(t, ext), contentType, data)
}

// 更新と同じ権限。ブース: 管理者・学生会・その担当、出演者: 管理者。あわせて相手があるかも確かめる
func (u *ImageUsecase) canEdit(ctx context.Context, t domain.ImageTarget) error {
	reqUser, ok := ctx.Value(ContextRequestUserKey).(RequestUser)
	if !ok {
		return ErrForbidden
	}
	switch t.Kind {
	case "booth":
		switch {
		case reqUser.Role == domain.RoleAdmin, reqUser.Role == domain.RoleGakuseikai:
		case reqUser.Role == domain.RoleStudent && reqUser.AssignedBoothID == t.ID:
		default:
			return ErrForbidden
		}
		_, err := u.booths.GetByID(ctx, t.ID)
		return err
	default:
		if reqUser.Role != domain.RoleAdmin {
			return ErrForbidden
		}
		_, err := u.stage.GetPerformer(ctx, t.ID)
		return err
	}
}

// 保存しようとしている画像を確かめる。自分の置き場所の画像なら、その物の置き場所のもので、まだあること。
// 外の URL や空はそのまま通す
func checkImageURL(ctx context.Context, store domain.ImageStore, t domain.ImageTarget, url string) error {
	if store == nil || url == "" {
		return nil
	}
	key, ok := store.KeyOf(url)
	if !ok {
		return nil
	}
	// ほかのブース・出演者の写真は使えない(使えると、外したときにその写真を消してしまう)
	if !strings.HasPrefix(key, domain.ImagePrefix(t)) {
		return domain.ErrInvalidImage
	}
	// フォームを開いている間にほかの人が写真を変え、前の写真が消えていた
	exists, err := store.Exists(ctx, url)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrImageChanged
	}
	return nil
}

// 画像が変わったら、前の画像(その物の置き場所のものだけ)を消す。消せなくても更新は成功のまま
func removeReplacedImage(ctx context.Context, store domain.ImageStore, t domain.ImageTarget, before, after string) {
	if store == nil || before == "" || before == after {
		return
	}
	key, ok := store.KeyOf(before)
	if !ok || !strings.HasPrefix(key, domain.ImagePrefix(t)) {
		return
	}
	if err := store.Delete(ctx, before); err != nil {
		log.Printf("古い画像を消せませんでした %s: %v", before, err)
	}
}
