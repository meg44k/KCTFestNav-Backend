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
	updateCongestionFn func(ctx context.Context, id int, congestionLevel int8) error
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

func (m *mockBoothRepository) UpdateCongestion(ctx context.Context, id int, congestionLevel int8) error {
	if m.updateCongestionFn != nil {
		return m.updateCongestionFn(ctx, id, congestionLevel)
	}
	return nil
}

func TestBoothUsecase_GetByID(t *testing.T) {
	t.Run("正常系: ブースを取得できる", func(t *testing.T) {
		expectedBooth, _ := domain.ReconstructBooth(1, "ブースA", "主催者A", "詳細A", int8(1), 10.0, 20.0, 30.0)
		repo := &mockBoothRepository{
			getByIDFn: func(ctx context.Context, id int) (*domain.Booth, error) {
				assert.Equal(t, 1, id)
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
		expectedBooth, _ := domain.ReconstructBooth(1, "ブースA", "主催者A", "詳細A", int8(1), 10.0, 20.0, 30.0)
		repo := &mockBoothRepository{
			getAllFn: func(ctx context.Context) ([]*domain.Booth, error) {
				return []*domain.Booth{expectedBooth}, nil
			},
		}

		uc := NewBoothUsecase(repo)
		booths, err := uc.GetAll(context.Background())

		require.NoError(t, err)
		assert.Len(t, booths, 1)
		assert.Equal(t, expectedBooth, booths[0])
	})
}

func boothCtxWithRole(role domain.Role) context.Context {
	return context.WithValue(context.Background(), domain.ContextUserRoleKey, role)
}

func TestBoothUsecase_Create(t *testing.T) {
	t.Run("正常系: ブースを作成できる(Admin)", func(t *testing.T) {
		ctx := boothCtxWithRole(domain.RoleAdmin)

		repo := &mockBoothRepository{
			createFn: func(ctx context.Context, booth *domain.Booth) error {
				assert.Equal(t, "ブースA", booth.Name)
				assert.Equal(t, float32(10.5), booth.X)
				return nil
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Create(ctx, "ブースA", "主催A", "詳細", 0, 10.5, 20.5, 0.0)

		require.NoError(t, err)
	})

	t.Run("異常系: 権限不足(Gakuseikai)", func(t *testing.T) {
		ctx := boothCtxWithRole(domain.RoleGakuseikai)
		uc := NewBoothUsecase(&mockBoothRepository{})
		
		err := uc.Create(ctx, "ブースA", "主催A", "詳細", 0, 10.5, 20.5, 0.0)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("異常系: DB保存に失敗した場合はエラーが返る", func(t *testing.T) {
		ctx := boothCtxWithRole(domain.RoleAdmin)
		repo := &mockBoothRepository{
			createFn: func(ctx context.Context, booth *domain.Booth) error {
				return errors.New("db error")
			},
		}

		uc := NewBoothUsecase(repo)
		err := uc.Create(ctx, "ブースA", "主催A", "詳細", 0, 10.5, 20.5, 0.0)

		require.Error(t, err)
		assert.EqualError(t, err, "db error")
	})

	t.Run("異常系: 空の名前はドメイン生成のバリデーションエラーになる", func(t *testing.T) {
		ctx := boothCtxWithRole(domain.RoleAdmin)
		uc := NewBoothUsecase(&mockBoothRepository{})
		
		err := uc.Create(ctx, "", "主催A", "詳細", 0, 10.5, 20.5, 0.0)

		require.Error(t, err)
	})
}
