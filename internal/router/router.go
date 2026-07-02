package router

import (
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	mv "github.com/meg44k/KCTFestNav-Backend/internal/middleware"
)

func InitRoutes(e *echo.Echo, h *handler.Handlers) {
	e.Use(middleware.RequestLogger())

	// 認証系
	auth := e.Group("/auth")

	auth.POST("/login", h.User.Login) // ログイン用API
	auth.GET("/me", handler.OK)       // ログイン中のユーザ情報取得

	// ブース情報
	e.GET("/booths", h.Booth.GetAll)      // 全ブースの情報を取得
	e.GET("/booths/:id", h.Booth.GetByID) // 特定のブースIDの情報を取得

	// ライブイベント
	e.GET("/lives", h.Live.GetAll)                 // 全ライブイベントの情報を取得
	e.GET("/lives/:id", h.Live.GetByID)            // 特定のライブIDの情報を取得
	e.GET("/lives/current", h.Live.GetCurrentLive) // 現在進行中のライブ情報を取得

	// アナウンス
	e.GET("/announcements", handler.OK)         // お知らせの一覧を表示
	e.GET("/announcements/current", handler.OK) // 現在のお知らせを表示

	// 管理者系
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	manage := e.Group("/manage")
	manage.Use(mv.JWTAuth(jwtSecret))

	manage.POST("/booths", h.Booth.Create)                           // ブースの追加
	manage.PUT("/booths/:id", h.Booth.Update)                        // ブースの更新
	manage.PATCH("/booths/:id/congestion", h.Booth.UpdateCongestion) // ブースの混雑度の変更
	manage.DELETE("/booths/:id", h.Booth.Delete)                     // ブースの削除

	manage.GET("/users", handler.OK)            // 全ユーザの取得
	manage.POST("/users", h.User.Create)        // ユーザの追加
	manage.DELETE("/users/:userId", handler.OK) // ユーザの削除

	manage.POST("/lives", h.Live.Create)
	manage.PUT("/lives/:id", h.Live.Update)
	manage.PATCH("/lives/:id/status", h.Live.UpdateLiveStatus)
	manage.DELETE("/lives/:id", h.Live.Delete)
}
