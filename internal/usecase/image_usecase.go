package usecase

import (
	"context"
	"log"

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

// 画像が変わったら、前の画像(自分の置き場所のものだけ)を消す。消せなくても更新は成功のまま
func removeReplacedImage(ctx context.Context, store domain.ImageStore, before, after string) {
	if store == nil || before == "" || before == after || !store.Owns(before) {
		return
	}
	if err := store.Delete(ctx, before); err != nil {
		log.Printf("古い画像を消せませんでした %s: %v", before, err)
	}
}
