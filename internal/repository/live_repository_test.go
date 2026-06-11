package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

func TestLiveRepository_Create(t *testing.T) {
	// テスト用DBのDSN（※ご自身の環境のパスワードやDB名に合わせて変更してください）
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}
	defer db.Close()

	// DBが実際に起動していない場合はテストをスキップしてコケないようにする
	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	// 毎回テスト前に lives テーブルを空にして、まっさらな状態からテストする
	_, _ = db.Exec("TRUNCATE TABLE lives")

	// リポジトリの初期化（Redisは今回は使わないので nil）
	repo := NewLiveRepository(db, nil)
	ctx := context.Background()

	t.Run("正常にライブイベントを保存できる", func(t *testing.T) {
		// 1. 保存するデータを準備
		// ※domain.NewLiveの引数は実際のものに合わせて調整してください
		live, err := domain.NewLive(
			"テストライブ",
			"テストの詳細です",
			"https://example.com/thumb.jpg",
			time.Now(),
			time.Now().Add(1*time.Hour),
			1,
		)
		assert.NoError(t, err)

		// 2. Create メソッドを実行
		err = repo.Create(ctx, live)

		// 3. エラーが出ずに成功したか検証
		assert.NoError(t, err)

		// 4. 本当にDBに保存されているか、レコードの数を数えて検証する
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM lives").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "livesテーブルのレコード数が1件になっていること")
	})
}

func TestLiveRepository_GetByID(t *testing.T) {
	dsn := "root:@tcp(127.0.0.1:3306)/kctfest_test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("DBの初期化エラー: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("テスト用DBが起動していないためスキップします: %v", err)
	}

	_, _ = db.Exec("TRUNCATE TABLE lives")
	repo := NewLiveRepository(db, nil)
	ctx := context.Background()

	t.Run("存在するIDを指定した場合、正常に取得できる", func(t *testing.T) {
		// 1. テスト用にダミーデータを1件保存しておく
		live, _ := domain.NewLive(
			"取得テストライブ",
			"取得のテスト詳細",
			"https://example.com/get.jpg",
			time.Now().Round(time.Second), // DB保存時にミリ秒以下が丸められる対策
			time.Now().Add(1*time.Hour).Round(time.Second),
			2,
		)
		err := repo.Create(ctx, live)
		assert.NoError(t, err)

		// 2. DBから実際に採番されたIDを取得する（ハードコードを避けて堅牢にする）
		var insertedID int
		err = db.QueryRow("SELECT id FROM lives LIMIT 1").Scan(&insertedID)
		assert.NoError(t, err)

		// 3. 取得した実際のIDを使って GetByID を実行する
		fetchedLive, err := repo.GetByID(ctx, insertedID)

		// 4. 検証（エラーが出ず、名前が一致しているか）
		assert.NoError(t, err)
		if assert.NotNil(t, fetchedLive) {
			assert.Equal(t, "取得テストライブ", fetchedLive.Name)
			// assert.Equal(t, 2, fetchedLive.SessionNumber()) // メソッド名に合わせて追加で検証も可能
		}
	})

	t.Run("存在しないIDを指定した場合、エラーが返る", func(t *testing.T) {
		// 存在しない適当なID（999など）を指定
		fetchedLive, err := repo.GetByID(ctx, 999)

		// エラーが返ってきており、データは取得できていない(nil)ことを検証
		assert.Error(t, err)
		assert.Nil(t, fetchedLive)
	})
}
