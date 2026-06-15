package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
)

// setupRedis はテスト用のRedisクライアントを作成します
func setupRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("テスト用Redisが起動していないためスキップします: %v", err)
	}
	// テストの独立性を保つため、開始時に current などのキーを消しておく
	rdb.Del(context.Background(), "lives:current")
	return rdb
}

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

	_, _ = db.Exec("TRUNCATE TABLE lives")
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
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

func TestLiveRepository_GetCurrentLive(t *testing.T) {
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
	ctx := context.Background()

	t.Run("進行中のライブが1件取得できること", func(t *testing.T) {
		_, _ = db.Exec("TRUNCATE TABLE lives")

		// 1. ダミーデータを2件用意（Create時は全て未開催ステータスになる想定）
		live1, _ := domain.NewLive("終わったライブ", "詳細", "", time.Now(), time.Now().Add(time.Hour), 1)
		live2, _ := domain.NewLive("進行中ライブ", "詳細", "", time.Now().Add(time.Hour), time.Now().Add(2*time.Hour), 2)
		
		_ = repo.Create(ctx, live1)
		_ = repo.Create(ctx, live2)

		// 2. 進行中のもの（live2）だけ、テスト用DB上でステータスを直接進行中（1）に書き換える
		_, err = db.Exec("UPDATE lives SET status = 1 WHERE name = '進行中ライブ'")
		assert.NoError(t, err)

		// 3. GetCurrentLive を実行
		currentLive, err := repo.GetCurrentLive(ctx)

		// 4. 検証
		assert.NoError(t, err)
		if assert.NotNil(t, currentLive) {
			assert.Equal(t, "進行中ライブ", currentLive.Name)
			assert.Equal(t, int8(1), int8(currentLive.Status()))
		}
	})

	t.Run("進行中のライブがない場合はエラーになること", func(t *testing.T) {
		_, _ = db.Exec("TRUNCATE TABLE lives")
		rdb.Del(ctx, "lives:current") // キャッシュもクリアする
		// データが空の状態で取得を試みる
		currentLive, err := repo.GetCurrentLive(ctx)

		assert.Error(t, err)
		assert.Nil(t, currentLive)
	})
}

func TestLiveRepository_UpdateLiveStatus(t *testing.T) {
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
	rdb := setupRedis(t)
	defer rdb.Close()
	repo := NewLiveRepository(db, rdb)
	ctx := context.Background()

	t.Run("ステータスのみを正常に更新し、Redisキャッシュを削除できる", func(t *testing.T) {
		live, _ := domain.ReconstructLive(0, "ステータス変更用", "詳細", "url", time.Now(), time.Now().Add(time.Hour), 1, 0)
		err := repo.Create(ctx, live)
		assert.NoError(t, err)

		var id int
		err = db.QueryRow("SELECT id FROM lives LIMIT 1").Scan(&id)
		assert.NoError(t, err)

		rdb.Set(ctx, "lives:current", "dummy_data", 0)

		err = repo.UpdateLiveStatus(ctx, id, domain.LiveStatus(1))
		assert.NoError(t, err)

		var newStatus int
		err = db.QueryRow("SELECT status FROM lives WHERE id = ?", id).Scan(&newStatus)
		assert.NoError(t, err)
		assert.Equal(t, 1, newStatus)

		val, err := rdb.Get(ctx, "lives:current").Result()
		assert.Equal(t, redis.Nil, err)
		assert.Empty(t, val)
	})
}
