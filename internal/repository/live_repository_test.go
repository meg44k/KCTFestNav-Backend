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

func TestLiveRepository_GetAll(t *testing.T) {
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

	t.Run("複数件のデータが正常に取得できること", func(t *testing.T) {
		// 1. テスト用にダミーデータを2件保存する
		live1, _ := domain.NewLive(
			"ライブA",
			"詳細A",
			"",
			time.Now().Round(time.Second),
			time.Now().Add(1*time.Hour).Round(time.Second),
			1,
		)
		live2, _ := domain.NewLive(
			"ライブB",
			"詳細B",
			"",
			time.Now().Round(time.Second),
			time.Now().Add(2*time.Hour).Round(time.Second),
			2,
		)
		
		err := repo.Create(ctx, live1)
		assert.NoError(t, err, "1件目のセットアップ失敗")
		
		err = repo.Create(ctx, live2)
		assert.NoError(t, err, "2件目のセットアップ失敗")

		// 2. GetAll を実行
		lives, err := repo.GetAll(ctx)

		// 3. 検証（エラーが出ず、2件取得できているか）
		assert.NoError(t, err)
		if assert.NotNil(t, lives) {
			assert.Len(t, lives, 2, "2件のライブデータが取得できるはず")
			
			// 取り出したデータの中身が合っているか確認
			// ※ 取得順（ORDER BY）を指定していない場合は順不同で返る可能性があるため、
			// 名前で検索するか、簡単なチェックにとどめるのが無難です
			names := []string{lives[0].Name, lives[1].Name}
			assert.Contains(t, names, "ライブA")
			assert.Contains(t, names, "ライブB")
		}
	})

	t.Run("データが1件もない場合は空の配列が返ること", func(t *testing.T) {
		// テーブルを空にする
		_, _ = db.Exec("TRUNCATE TABLE lives")

		lives, err := repo.GetAll(ctx)

		// エラーにはならず、長さ0のスライスが返ってくるはず
		assert.NoError(t, err)
		assert.NotNil(t, lives)
		assert.Len(t, lives, 0)
	})
}

func TestLiveRepository_Update(t *testing.T) {
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

	t.Run("正常にデータを更新できること", func(t *testing.T) {
		// 1. テスト用にダミーデータを1件保存する
		live, _ := domain.NewLive(
			"更新前のライブ",
			"古い詳細",
			"",
			time.Now().Round(time.Second),
			time.Now().Add(1*time.Hour).Round(time.Second),
			1,
		)
		err := repo.Create(ctx, live)
		assert.NoError(t, err, "セットアップ失敗")

		// 2. 実際に採番されたIDを取得する
		var insertedID int
		err = db.QueryRow("SELECT id FROM lives LIMIT 1").Scan(&insertedID)
		assert.NoError(t, err)

		// 3. ドメインモデルの値を更新する
		// ※ Createしただけの `live` 変数はIDが0のままなので、取得したIDをセットしてあげます
		live.ID = insertedID
		live.Name = "更新されたスーパーライブ！"
		live.Detail = "詳細も新しくなりました"

		// 4. Update を実行
		err = repo.Update(ctx, live)

		// 5. 検証（エラーが出ずに更新できたか）
		assert.NoError(t, err)

		// 6. 本当にDBの中身が変わっているか、GetByID で取り直して確認する
		updatedLive, err := repo.GetByID(ctx, insertedID)
		assert.NoError(t, err)
		if assert.NotNil(t, updatedLive) {
			assert.Equal(t, "更新されたスーパーライブ！", updatedLive.Name)
			assert.Equal(t, "詳細も新しくなりました", updatedLive.Detail)
		}
	})
}

func TestLiveRepository_Delete(t *testing.T) {
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

	t.Run("正常にデータを削除できること", func(t *testing.T) {
		// 1. テスト用にダミーデータを1件保存する
		live, _ := domain.NewLive(
			"削除される運命のライブ",
			"詳細",
			"",
			time.Now().Round(time.Second),
			time.Now().Add(1*time.Hour).Round(time.Second),
			1,
		)
		err := repo.Create(ctx, live)
		assert.NoError(t, err, "セットアップ失敗")

		// 2. 実際に採番されたIDを取得する
		var insertedID int
		err = db.QueryRow("SELECT id FROM lives LIMIT 1").Scan(&insertedID)
		assert.NoError(t, err)

		// 3. Delete を実行
		err = repo.Delete(ctx, insertedID)

		// 4. 検証（エラーが出ずに削除できたか）
		assert.NoError(t, err)

		// 5. 本当にDBから消えているか、GetByID で取り直して確認する
		deletedLive, err := repo.GetByID(ctx, insertedID)
		
		// 削除されているので「見つかりません」というエラーが返ってくるはず
		assert.Error(t, err)
		assert.Nil(t, deletedLive, "削除されたデータは取得できないこと")
	})
}
