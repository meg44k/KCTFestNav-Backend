package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/usecase"
)

func main() {
	// 環境変数読み込み
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// echoフレームワーク初期化
	e := echo.New()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	// MySQL設定
	cfg := mysql.Config{
		User:      os.Getenv("DB_USER"),
		Passwd:    os.Getenv("DB_PASS"),
		Net:       "tcp",
		Addr:      os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_NAME"),
		ParseTime: true,
	}
	db, err := sql.Open("mysql", cfg.FormatDSN())
	// Redis設定
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// 依存関係注入
	liveRepo := repository.NewLiveRepository(db, rdb)
	liveUsecase := usecase.NewLiveUsecase(liveRepo)
	liveHandler := handler.NewLiveHandler(liveUsecase)

	// ハンドラをまとめてルーターに渡す(ここもっと良くなるかも)
	handlers := &handler.Handlers{
		Live: liveHandler,
	}

	router.InitRoutes(e, handlers)

	if err := e.Start(":1323"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
