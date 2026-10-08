package router

import (
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	mv "github.com/meg44k/KCTFestNav-Backend/internal/middleware"
)

func InitRoutes(e *echo.Echo, h *handler.Handlers) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	e.Use(middleware.RequestLogger())

	// 認証系
	auth := e.Group("/auth")

	auth.POST("/login", h.User.Login)                    // ログイン用API
	auth.GET("/me", h.User.GetMe, mv.JWTAuth(jwtSecret)) // ログイン中のユーザ情報取得

	// ブース情報
	e.GET("/booths", h.Booth.GetAll)      // 全ブースの情報を取得
	e.GET("/booths/:id", h.Booth.GetByID) // 特定のブースIDの情報を取得

	// ステージイベント(セクション・ブロック・出演者)は StageRoutes で登録する

	// アナウンス
	e.GET("/announcements", h.Announcement.Get) // 現在のお知らせを表示

	// 管理者系
	manage := e.Group("/manage")
	manage.Use(mv.JWTAuth(jwtSecret))

	manage.POST("/booths", h.Booth.Create)                           // ブースの追加
	manage.PUT("/booths/:id", h.Booth.Update)                        // ブースの更新
	manage.PATCH("/booths/:id/congestion", h.Booth.UpdateCongestion) // ブースの混雑度の変更
	manage.DELETE("/booths/:id", h.Booth.Delete)                     // ブースの削除

	manage.PUT("/announcements", h.Announcement.Update) // アナウンスの更新

	manage.GET("/users/:id", h.User.GetByID)   // 特定ユーザの取得
	manage.GET("/users", h.User.GetAll)        // 全ユーザの取得
	manage.POST("/users", h.User.Create)       // ユーザの追加
	manage.PUT("/users/:id", h.User.Update)    // ユーザの更新
	manage.DELETE("/users/:id", h.User.Delete) // ユーザの削除

	StageRoutes(e, manage, h.Stage)
}

// ステージイベントのルート。e2e テストでも同じものを使う
func StageRoutes(e *echo.Echo, manage *echo.Group, h *handler.StageHandler) {
	e.GET("/stage", h.GetSchedule) // 全セクション・ブロック・出演者と演奏中

	// 当日の操作(学生会・管理者)
	manage.POST("/stage/blocks/:id/next", h.AdvanceBlock) // 次の出演者へ
	manage.POST("/stage/blocks/:id/prev", h.RewindBlock)  // 1 つ戻す

	// 番組表の編集(管理者)
	manage.POST("/stage/sections", h.CreateSection)
	manage.PUT("/stage/sections/:id", h.UpdateSection)
	manage.DELETE("/stage/sections/:id", h.DeleteSection)
	manage.POST("/stage/sections/:id/blocks", h.CreateBlock)
	manage.PUT("/stage/blocks/:id", h.UpdateBlock)
	manage.DELETE("/stage/blocks/:id", h.DeleteBlock)
	manage.POST("/stage/blocks/:id/performers", h.CreatePerformer)
	manage.PUT("/stage/performers/:id", h.UpdatePerformer)
	manage.DELETE("/stage/performers/:id", h.DeletePerformer)
	manage.POST("/stage/performers/:id/move", h.MovePerformer) // 出演順を上下に入れ替える
}
