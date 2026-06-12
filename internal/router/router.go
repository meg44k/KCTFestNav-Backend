package router

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
)

func InitRoutes(e *echo.Echo, h *handler.Handlers) {
	e.Use(middleware.RequestLogger())

	// 認証系
	auth := e.Group("/auth")

	auth.POST("/login", handler.OK) // ログイン用API
	auth.GET("/me", handler.OK)     // ログイン中のユーザ情報取得

	// ブース情報
	e.GET("/booth", handler.OK)           // 全ブースの情報を取得
	e.GET("/booths/:boothId", handler.OK) // 特定のブースIDの情報を取得

	// ライブイベント
	e.GET("/lives", h.Live.GetAll)                 // 全ライブイベントの情報を取得
	e.GET("/lives/:id", h.Live.GetByID)            // 特定のライブIDの情報を取得
	e.GET("/lives/current", h.Live.GetCurrentLive) // 現在進行中のライブ情報を取得

	// アナウンス
	e.GET("/announcements", handler.OK)         // お知らせの一覧を表示
	e.GET("/announcements/current", handler.OK) // 現在のお知らせを表示

	// 管理者系
	manage := e.Group("/manage")

	manage.POST("/booths", handler.OK)            // ブースの追加
	manage.PUT("/booths/:boothId", handler.OK)    // ブースの更新
	manage.DELETE("/booths/:boothId", handler.OK) // ブースの削除

	manage.GET("/users", handler.OK)            // 全ユーザの取得
	manage.POST("/users", handler.OK)           // ユーザの追加
	manage.DELETE("/users/:userId", handler.OK) // ユーザの削除

	manage.POST("/lives", h.Live.Create)
	manage.PUT("/lives/:id", h.Live.Update)
	manage.DELETE("/lives/:id", h.Live.Delete)
}
