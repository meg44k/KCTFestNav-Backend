package repository_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupUserTestDB(t *testing.T) (*sql.DB, *redis.Client) {
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}

	// テーブル初期化
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
	_, err = db.Exec("TRUNCATE TABLE users")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		t.Fatalf("Failed to truncate users table: %v", err)
	}

	return db, rdb
}

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	db, rdb := setupUserTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := repository.NewUserRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: ユーザーを作成してGetByIDで取得できること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		targetID := uuid.New()
		user, _ := domain.ReconstructUser(targetID, "リポジトリテスト", "repo_test", []byte("hashed"), 0, domain.RoleStudent)

		// Create実行
		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		// DBに保存されているか確認 (GetByID)
		fetchedUser, err := repo.GetByID(ctx, targetID)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedUser)

		assert.Equal(t, targetID, fetchedUser.ID)
		assert.Equal(t, "リポジトリテスト", fetchedUser.Name)
		assert.Equal(t, "repo_test", fetchedUser.LoginID)
		assert.Equal(t, 0, fetchedUser.AssignedBoothID)
		assert.Equal(t, domain.RoleStudent, fetchedUser.Role)
	})

	t.Run("異常系: 存在しないIDを指定するとエラーになること", func(t *testing.T) {
		fetchedUser, err := repo.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, fetchedUser)
	})
}

func TestUserRepository_GetByLoginID(t *testing.T) {
	db, rdb := setupUserTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := repository.NewUserRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: LoginIDでユーザーを取得できること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		targetID := uuid.New()
		user, _ := domain.ReconstructUser(targetID, "リポジトリテスト", "login_test", []byte("hashed"), 0, domain.RoleStudent)
		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		fetchedUser, err := repo.GetByLoginID(ctx, "login_test")
		assert.NoError(t, err)
		assert.NotNil(t, fetchedUser)
		assert.Equal(t, "login_test", fetchedUser.LoginID)
	})

	t.Run("異常系: 存在しないLoginIDを指定すると ErrNotFound が返ること", func(t *testing.T) {
		fetchedUser, err := repo.GetByLoginID(ctx, "ghost_login_id")
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, fetchedUser)
	})
}

func TestUserRepository_GetAll(t *testing.T) {
	db, rdb := setupUserTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := repository.NewUserRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: 0件の場合は空のリストが返ること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		users, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 0)
	})

	t.Run("正常系: 複数件のユーザーを取得できること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		user1, _ := domain.ReconstructUser(uuid.New(), "テスト1", "test_1", []byte("hash"), 0, domain.RoleStudent)
		user2, _ := domain.ReconstructUser(uuid.New(), "テスト2", "test_2", []byte("hash"), 0, domain.RoleAdmin)
		err := repo.Create(ctx, user1)
		assert.NoError(t, err)
		err = repo.Create(ctx, user2)
		assert.NoError(t, err)

		users, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		// DBからの取得順序は保証されていないため、要素の存在確認をする
		var foundUser1, foundUser2 bool
		for _, u := range users {
			if u.Name == "テスト1" {
				assert.Equal(t, user1.ID, u.ID)
				assert.Equal(t, "test_1", u.LoginID)
				assert.Equal(t, domain.RoleStudent, u.Role)
				foundUser1 = true
			}
			if u.Name == "テスト2" {
				assert.Equal(t, user2.ID, u.ID)
				assert.Equal(t, "test_2", u.LoginID)
				assert.Equal(t, domain.RoleAdmin, u.Role)
				foundUser2 = true
			}
		}
		assert.True(t, foundUser1)
		assert.True(t, foundUser2)
	})

	t.Run("正常系: 所属ブース(AssignedBoothID)が正しく保存・復元されること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("TRUNCATE TABLE booths")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		// 事前にブースを作成しておく
		_, err := db.Exec("INSERT INTO booths (id, name, organizer, detail, x, y, z) VALUES (999, 'テストブース', 'テスト主催者', '詳細', 0, 0, 0)")
		assert.NoError(t, err)

		user, _ := domain.ReconstructUser(uuid.New(), "ブース所属テスト", "booth_test", []byte("hash"), 999, domain.RoleStudent)
		err = repo.Create(ctx, user)
		assert.NoError(t, err)

		users, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		assert.Equal(t, 999, users[0].AssignedBoothID)
		assert.Equal(t, "ブース所属テスト", users[0].Name)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, rdb := setupUserTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := repository.NewUserRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: ユーザー情報を更新できること", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, _ = db.Exec("TRUNCATE TABLE users")
		_, _ = db.Exec("TRUNCATE TABLE booths")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")

		// 更新用のブースを用意しておく
		_, _ = db.Exec("INSERT INTO booths (id, name, organizer, detail, x, y, z) VALUES (999, 'テストブース', '主催', '詳細', 0, 0, 0)")

		targetID := uuid.New()
		user, _ := domain.ReconstructUser(targetID, "元の名前", "login_id", []byte("hash"), 0, domain.RoleStudent)

		// Create
		err := repo.Create(ctx, user)
		assert.NoError(t, err)

		// 情報を書き換えたドメインモデルを用意 (ブースIDも999に変更)
		updatedUser, _ := domain.ReconstructUser(targetID, "更新後の名前", "updated_login", []byte("new_hash"), 999, domain.RoleAdmin)

		// Update
		err = repo.Update(ctx, updatedUser)
		assert.NoError(t, err)

		// DBから取得して更新されているか確認
		fetchedUser, err := repo.GetByID(ctx, targetID)
		assert.NoError(t, err)
		assert.Equal(t, "更新後の名前", fetchedUser.Name)
		assert.Equal(t, "updated_login", fetchedUser.LoginID)
		assert.Equal(t, domain.RoleAdmin, fetchedUser.Role)
		assert.Equal(t, 999, fetchedUser.AssignedBoothID)
	})
}
