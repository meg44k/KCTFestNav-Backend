package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// 画像の置き場所の作りもの。https://img/ で始まる URL を自分のものとする。missing に入れた URL は「もう無い」
type fakeImageStore struct {
	put       []string
	deleted   []string
	deleteErr error
	missing   map[string]bool
}

func (s *fakeImageStore) Put(_ context.Context, key, _ string, _ []byte) (string, error) {
	s.put = append(s.put, key)
	return "https://img/" + key, nil
}
func (s *fakeImageStore) Delete(_ context.Context, url string) error {
	s.deleted = append(s.deleted, url)
	return s.deleteErr
}
func (s *fakeImageStore) KeyOf(url string) (string, bool) {
	k, ok := strings.CutPrefix(url, "https://img/")
	return k, ok && k != ""
}
func (s *fakeImageStore) Exists(_ context.Context, url string) (bool, error) {
	return !s.missing[url], nil
}

// 出演者だけ持つステージのリポジトリ
type imageStageRepo struct {
	fakeStageRepo
	performers map[int]*domain.Performer
	updated    *domain.Performer
}

func (r *imageStageRepo) GetPerformer(_ context.Context, id int) (*domain.Performer, error) {
	if p, ok := r.performers[id]; ok {
		return p, nil
	}
	return nil, sql.ErrNoRows
}
func (r *imageStageRepo) UpdatePerformer(_ context.Context, p *domain.Performer) error {
	r.updated = p
	return nil
}

var pngBytes = []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 32))

func imageBooths(imageURL string) *mockBoothRepository {
	return &mockBoothRepository{getByIDFn: func(_ context.Context, id int) (*domain.Booth, error) {
		if id == 3 || id == 4 {
			return &domain.Booth{ID: id, Name: "b", ImageURL: imageURL}, nil
		}
		return nil, sql.ErrNoRows
	}}
}

func asRole(role domain.Role, booth int) context.Context {
	return context.WithValue(context.Background(), ContextRequestUserKey, RequestUser{ID: uuid.New(), Role: role, AssignedBoothID: booth})
}

func TestImageUsecase_ブースの権限(t *testing.T) {
	store := &fakeImageStore{}
	uc := NewImageUsecase(store, imageBooths(""), &imageStageRepo{})

	for _, ctx := range []context.Context{asRole(domain.RoleAdmin, 0), asRole(domain.RoleGakuseikai, 0), asRole(domain.RoleStudent, 3)} {
		url, err := uc.Upload(ctx, "booth:3", pngBytes)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(url, "https://img/booths/3/"), url)
	}
	_, err := uc.Upload(asRole(domain.RoleStudent, 4), "booth:3", pngBytes)
	assert.ErrorIs(t, err, ErrForbidden, "ほかのブースの担当")
	_, err = uc.Upload(asRole(domain.RoleMember, 0), "booth:3", pngBytes)
	assert.ErrorIs(t, err, ErrForbidden)
	_, err = uc.Upload(context.Background(), "booth:3", pngBytes)
	assert.ErrorIs(t, err, ErrForbidden, "ログインなし")
}

func TestImageUsecase_出演者は管理者だけ(t *testing.T) {
	stage := &imageStageRepo{performers: map[int]*domain.Performer{12: {ID: 12}}}
	uc := NewImageUsecase(&fakeImageStore{}, imageBooths(""), stage)
	url, err := uc.Upload(asRole(domain.RoleAdmin, 0), "performer:12", pngBytes)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(url, "https://img/performers/12/"))
	_, err = uc.Upload(asRole(domain.RoleGakuseikai, 0), "performer:12", pngBytes)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestImageUsecase_弾くもの(t *testing.T) {
	store := &fakeImageStore{}
	uc := NewImageUsecase(store, imageBooths(""), &imageStageRepo{})
	admin := asRole(domain.RoleAdmin, 0)
	_, err := uc.Upload(admin, "booth:99", pngBytes)
	assert.ErrorIs(t, err, sql.ErrNoRows, "無いブース")
	_, err = uc.Upload(admin, "performer:99", pngBytes)
	assert.ErrorIs(t, err, sql.ErrNoRows, "無い出演者")
	_, err = uc.Upload(admin, "user:1", pngBytes)
	assert.ErrorIs(t, err, domain.ErrInvalidImageTarget)
	_, err = uc.Upload(admin, "booth:3", []byte("<html>"))
	assert.ErrorIs(t, err, domain.ErrInvalidImage)
	assert.Empty(t, store.put, "どれも置かない")
}

func boothParams(imageURL string) domain.BoothParams {
	return domain.BoothParams{Name: "b", Organizer: "1-1", ImageURL: imageURL}
}

func TestBoothUsecase_差し替えたら古い画像を消す(t *testing.T) {
	cases := []struct {
		name, before, after string
		deleted             []string
	}{
		{"自分の置き場所から別の画像", "https://img/booths/3/old.webp", "https://img/booths/3/new.webp", []string{"https://img/booths/3/old.webp"}},
		{"自分の置き場所から外す", "https://img/booths/3/old.webp", "", []string{"https://img/booths/3/old.webp"}},
		{"同じ画像", "https://img/booths/3/old.webp", "https://img/booths/3/old.webp", nil},
		{"外の URL から", "https://example.com/a.png", "https://img/booths/3/new.webp", nil},
		{"空から", "", "https://img/booths/3/new.webp", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &fakeImageStore{}
			uc := NewBoothUsecase(imageBooths(c.before)).WithImages(store)
			require.NoError(t, uc.Update(asRole(domain.RoleAdmin, 0), 3, domain.BoothCongestionEmpty, boothParams(c.after)))
			assert.Equal(t, c.deleted, store.deleted)
		})
	}
}

func TestBoothUsecase_古い画像が消せなくても更新は成功(t *testing.T) {
	store := &fakeImageStore{deleteErr: errors.New("r2 down")}
	uc := NewBoothUsecase(imageBooths("https://img/booths/3/old.webp")).WithImages(store)
	assert.NoError(t, uc.Update(asRole(domain.RoleAdmin, 0), 3, domain.BoothCongestionEmpty, boothParams("")))
	assert.Len(t, store.deleted, 1)
}

func TestStageUsecase_出演者の画像を差し替えたら古い画像を消す(t *testing.T) {
	stage := &imageStageRepo{performers: map[int]*domain.Performer{12: {ID: 12, Name: "p", ThumbnailURL: "https://img/performers/12/old.webp"}}}
	store := &fakeImageStore{}
	uc := NewStageUsecase(stage).WithImages(store)
	require.NoError(t, uc.UpdatePerformer(asRole(domain.RoleAdmin, 0), 12, PerformerInput{Name: "p", ThumbnailURL: "https://img/performers/12/new.webp"}))
	assert.Equal(t, []string{"https://img/performers/12/old.webp"}, store.deleted)
	assert.Equal(t, "https://img/performers/12/new.webp", stage.updated.ThumbnailURL)
}

func TestBoothUsecase_ほかのブースの写真は使えず消さない(t *testing.T) {
	store := &fakeImageStore{}
	uc := NewBoothUsecase(imageBooths("https://img/booths/3/mine.webp")).WithImages(store)
	// ブース 4 の担当が、ブース 3 の写真の URL を入れる
	err := uc.Update(asRole(domain.RoleStudent, 4), 4, domain.BoothCongestionEmpty, boothParams("https://img/booths/3/mine.webp"))
	assert.ErrorIs(t, err, domain.ErrInvalidImage)
	assert.Empty(t, store.deleted)

	// 前の画像がほかのブースの置き場所のものなら、外しても消さない
	uc = NewBoothUsecase(imageBooths("https://img/booths/3/other.webp")).WithImages(store)
	require.NoError(t, uc.Update(asRole(domain.RoleAdmin, 0), 4, domain.BoothCongestionEmpty, boothParams("")))
	assert.Empty(t, store.deleted)
}

func TestBoothUsecase_開いている間に写真が変えられていたら断る(t *testing.T) {
	// 開いたときは X。その間にほかの人が Y にして X は消えた。古いフォームが X のまま保存する
	store := &fakeImageStore{missing: map[string]bool{"https://img/booths/3/x.webp": true}}
	updated := false
	repo := imageBooths("https://img/booths/3/y.webp")
	repo.updateFn = func(context.Context, *domain.Booth) error { updated = true; return nil }
	uc := NewBoothUsecase(repo).WithImages(store)
	err := uc.Update(asRole(domain.RoleAdmin, 0), 3, domain.BoothCongestionEmpty, boothParams("https://img/booths/3/x.webp"))
	assert.ErrorIs(t, err, domain.ErrImageChanged)
	assert.False(t, updated, "保存しない")
	assert.Empty(t, store.deleted, "Y を消さない")
}

func TestStageUsecase_ほかの出演者の写真は使えない(t *testing.T) {
	stage := &imageStageRepo{performers: map[int]*domain.Performer{12: {ID: 12, Name: "p"}}}
	uc := NewStageUsecase(stage).WithImages(&fakeImageStore{})
	err := uc.UpdatePerformer(asRole(domain.RoleAdmin, 0), 12, PerformerInput{Name: "p", ThumbnailURL: "https://img/performers/99/a.webp"})
	assert.ErrorIs(t, err, domain.ErrInvalidImage)
}
