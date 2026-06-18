package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

func setupBoothTestDB(t *testing.T) (*sql.DB, *redis.Client) {
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
	_, err = db.Exec("TRUNCATE TABLE booths")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
	if err != nil {
		t.Fatalf("Failed to truncate booths table: %v", err)
	}

	return db, rdb
}

func TestBoothRepository_GetByID(t *testing.T) {
	db, rdb := setupBoothTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := NewBoothRepository(db, rdb)
	ctx := context.Background()

	// 事前にデータを1件INSERTしておく
	res, err := db.ExecContext(ctx, `
		INSERT INTO booths (name, organizer, detail, congestion_status, x, y, z)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, "テストブース", "学生会", "詳細テキスト", 1, 10.5, 20.5, 30.5)
	assert.NoError(t, err)

	insertedID, err := res.LastInsertId()
	assert.NoError(t, err)

	t.Run("正常系: 存在するIDでブースを取得できる", func(t *testing.T) {
		booth, err := repo.GetByID(ctx, int(insertedID))

		assert.NoError(t, err)
		assert.NotNil(t, booth)
		assert.Equal(t, int(insertedID), booth.ID)
		assert.Equal(t, "テストブース", booth.Name)
		assert.Equal(t, "学生会", booth.Organizer)
		assert.Equal(t, "詳細テキスト", booth.Detail)
		assert.Equal(t, int8(1), booth.CongestionStatus())
		assert.Equal(t, float32(10.5), booth.X)
	})

	t.Run("異常系: 存在しないIDを指定するとエラーになる", func(t *testing.T) {
		booth, err := repo.GetByID(ctx, 9999)

		assert.Error(t, err)
		assert.Nil(t, booth)
	})
}

func TestBoothRepository_GetAll(t *testing.T) {
	db, rdb := setupBoothTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := NewBoothRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: 0件の場合は空のスライスが返る", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, err := db.Exec("TRUNCATE TABLE booths")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		assert.NoError(t, err)

		booths, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, booths, 0)
	})

	t.Run("正常系: 複数件のブースを取得できる", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, err := db.Exec("TRUNCATE TABLE booths")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		assert.NoError(t, err)

		// 2件INSERT
		_, err = db.ExecContext(ctx, `
			INSERT INTO booths (name, organizer, detail, congestion_status, x, y, z)
			VALUES 
			(?, ?, ?, ?, ?, ?, ?),
			(?, ?, ?, ?, ?, ?, ?)
		`,
			"ブースA", "主催A", "詳細A", 0, 1.0, 2.0, 3.0,
			"ブースB", "主催B", "詳細B", 2, 4.0, 5.0, 6.0,
		)
		assert.NoError(t, err)

		booths, err := repo.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, booths, 2)

		assert.Equal(t, "ブースA", booths[0].Name)
		assert.Equal(t, "主催A", booths[0].Organizer)

		assert.Equal(t, "ブースB", booths[1].Name)
		assert.Equal(t, int8(2), booths[1].CongestionStatus())
	})
}

func TestBoothRepository_Create(t *testing.T) {
	db, rdb := setupBoothTestDB(t)
	defer db.Close()
	defer rdb.Close()

	repo := NewBoothRepository(db, rdb)
	ctx := context.Background()

	t.Run("正常系: 新規ブースを作成できる", func(t *testing.T) {
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 0")
		_, err := db.Exec("TRUNCATE TABLE booths")
		_, _ = db.Exec("SET FOREIGN_KEY_CHECKS = 1")
		assert.NoError(t, err)

		booth, _ := domain.NewBooth("新規ブース", "学生会", "詳細", 10.5, 20.5, 30.5)

		err = repo.Create(ctx, booth)
		assert.NoError(t, err)

		// 実際にDBに保存されたか検証
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM booths").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)

		var name, organizer string
		err = db.QueryRow("SELECT name, organizer FROM booths LIMIT 1").Scan(&name, &organizer)
		assert.NoError(t, err)
		assert.Equal(t, "新規ブース", name)
		assert.Equal(t, "学生会", organizer)
	})
}
