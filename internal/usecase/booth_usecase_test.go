package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBoothRepository struct {
	createFn           func(ctx context.Context, booth *domain.Booth) error
	updateFn           func(ctx context.Context, booth *domain.Booth) error
	deleteFn           func(ctx context.Context, id int) error
	getByIDFn          func(ctx context.Context, id int) (*domain.Booth, error)
	getAllFn           func(ctx context.Context) ([]*domain.Booth, error)
	updateCongestionFn func(ctx context.Context, id int, congestionLevel domain.CongestionStatus) error
}

func (m *mockBoothRepository) Create(ctx context.Context, booth *domain.Booth) error {
	if m.createFn != nil {
		return m.createFn(ctx, booth)
	}
	return nil
}

func (m *mockBoothRepository) Update(ctx context.Context, booth *domain.Booth) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, booth)
	}
	return nil
}

func (m *mockBoothRepository) Delete(ctx context.Context, id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockBoothRepository) GetByID(ctx context.Context, id int) (*domain.Booth, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockBoothRepository) GetAll(ctx context.Context) ([]*domain.Booth, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx)
	}
	return nil, nil
}

func (m *mockBoothRepository) UpdateCongestion(ctx context.Context, id int, congestionLevel domain.CongestionStatus) error {
	if m.updateCongestionFn != nil {
		return m.updateCongestionFn(ctx, id, congestionLevel)
	}
	return nil
}

// 権限チェック用のヘルパーは live_usecase_test.go の ctxWithRole を利用できるが、


func boothCtxWithRequestUser(role domain.Role, assignedBoothID int) context.Context {
	reqUser := RequestUser{
		Role:            role,
		AssignedBoothID: assignedBoothID,
	}
	return context.WithValue(context.Background(), ContextRequestUserKey, reqUser)
}

func TestBoothUsecase_Create(t *testing.T) {
	t.Run("正常系: ブースを作成できる(Admin)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleAdmin, 0)

		repo := &mockBoothRepository{
			createFn: func(ctx context.Context, booth *domain.Booth) error {
				assert.Equal(t, "ブースA", booth.Name)
				assert.Equal(t, float32(10.5), booth.X)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Create(ctx, domain.BoothParams{
			Name:      "ブースA",
			Organizer: "1-1",
			Detail:    "詳細",
			X:         10.5,
			Y:         20.5,
		})

		require.NoError(t, err)
	})

	t.Run("異常系: 権限不足(Gakuseikai)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleGakuseikai, 0)
		uc := NewBoothUsecase(&mockBoothRepository{})

		err := uc.Create(ctx, domain.BoothParams{
			Name:      "ブースA",
			Organizer: "1-1",
			Detail:    "詳細",
			X:         10.5,
			Y:         20.5,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("異常系: DB保存に失敗", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleAdmin, 0)
		repo := &mockBoothRepository{
			createFn: func(ctx context.Context, booth *domain.Booth) error {
				return errors.New("db error")
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Create(ctx, domain.BoothParams{
			Name:      "ブースA",
			Organizer: "1-1",
			Detail:    "詳細",
			X:         10.5,
			Y:         20.5,
		})

		require.Error(t, err)
		assert.EqualError(t, err, "db error")
	})
}

func TestBoothUsecase_Update(t *testing.T) {
	t.Run("正常系: ブースを更新できる(Gakuseikai)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleGakuseikai, 1)

		repo := &mockBoothRepository{
			updateFn: func(ctx context.Context, booth *domain.Booth) error {
				assert.Equal(t, 1, booth.ID)
				assert.Equal(t, "ブース更新", booth.Name)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Update(ctx, 1, domain.CongestionStatus(1), domain.BoothParams{
			Name:      "ブース更新",
			Organizer: "1-1",
			Detail:    "詳細",
			X:         10.5,
			Y:         20.5,
		})

		require.NoError(t, err)
	})

	t.Run("異常系: 権限不足(Student)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleStudent, 2) // assignされているIDと更新対象(1)が違うのでエラーになる
		uc := NewBoothUsecase(&mockBoothRepository{})

		err := uc.Update(ctx, 1, domain.CongestionStatus(1), domain.BoothParams{
			Name:      "ブース更新",
			Organizer: "1-1",
			Detail:    "詳細",
			X:         10.5,
			Y:         20.5,
		})

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})
}

func TestBoothUsecase_Delete(t *testing.T) {
	t.Run("正常系: ブースを削除できる(Admin)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleAdmin, 0)

		repo := &mockBoothRepository{
			deleteFn: func(ctx context.Context, id int) error {
				assert.Equal(t, 1, id)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Delete(ctx, 1)

		require.NoError(t, err)
	})

	t.Run("異常系: 権限不足(Gakuseikai)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleGakuseikai, 0)
		uc := NewBoothUsecase(&mockBoothRepository{})

		err := uc.Delete(ctx, 1)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})
}

func TestBoothUsecase_GetByID(t *testing.T) {
	t.Run("正常系: ブースを取得できる", func(t *testing.T) {
		expectedBooth, _ := domain.ReconstructBooth(1, domain.BoothCongestionEmpty, domain.BoothParams{Name: "A", Organizer: "O", Detail: "D"})
		repo := &mockBoothRepository{
			getByIDFn: func(ctx context.Context, id int) (*domain.Booth, error) {
				return expectedBooth, nil
			},
		}

		uc := NewBoothUsecase(repo)
		booth, err := uc.GetByID(context.Background(), 1)

		require.NoError(t, err)
		assert.Equal(t, expectedBooth, booth)
	})
}

func TestBoothUsecase_GetAll(t *testing.T) {
	t.Run("正常系: ブース一覧を取得できる", func(t *testing.T) {
		expectedBooth, _ := domain.ReconstructBooth(1, domain.BoothCongestionEmpty, domain.BoothParams{Name: "A", Organizer: "O", Detail: "D"})
		repo := &mockBoothRepository{
			getAllFn: func(ctx context.Context) ([]*domain.Booth, error) {
				return []*domain.Booth{expectedBooth}, nil
			},
		}

		uc := NewBoothUsecase(repo)
		booths, err := uc.GetAll(context.Background())

		require.NoError(t, err)
		assert.Len(t, booths, 1)
	})
}

func TestBoothUsecase_UpdateCongestion(t *testing.T) {
	t.Run("正常系: 混雑状況を更新できる(Gakuseikai)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleGakuseikai, 1)

		repo := &mockBoothRepository{
			updateCongestionFn: func(ctx context.Context, id int, level domain.CongestionStatus) error {
				assert.Equal(t, 1, id)
				assert.Equal(t, domain.CongestionStatus(2), level)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.UpdateCongestion(ctx, 1, 2)

		require.NoError(t, err)
	})

	t.Run("異常系: 不正な混雑度(3)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleGakuseikai, 1)
		uc := NewBoothUsecase(&mockBoothRepository{})

		err := uc.UpdateCongestion(ctx, 1, 3)

		require.Error(t, err)
		// ドメインバリデーションエラーになるはず
	})

	t.Run("正常系: 混雑状況を更新できる(Student)", func(t *testing.T) {
		ctx := boothCtxWithRequestUser(domain.RoleStudent, 1) // assignされているIDと更新対象(1)が一致するのでOK

		repo := &mockBoothRepository{
			updateCongestionFn: func(ctx context.Context, id int, level domain.CongestionStatus) error {
				assert.Equal(t, 1, id)
				assert.Equal(t, domain.CongestionStatus(1), level)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.UpdateCongestion(ctx, 1, 1)

		require.NoError(t, err)
	})
}
