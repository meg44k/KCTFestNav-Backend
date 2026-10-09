package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"

	"github.com/meg44k/KCTFestNav-Backend/internal/config"
	"github.com/meg44k/KCTFestNav-Backend/internal/domain"
	"github.com/meg44k/KCTFestNav-Backend/internal/handler"
	"github.com/meg44k/KCTFestNav-Backend/internal/repository"
	"github.com/meg44k/KCTFestNav-Backend/internal/router"
	"github.com/meg44k/KCTFestNav-Backend/internal/storage"
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

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}

	// MySQL設定
	db, err := sql.Open("mysql", cfg.MySQL.FormatDSN())
	if err != nil {
		log.Printf("%v", err)
	}
	// Cloud SQL の小さい台はつなげる数が少ないので、1 台あたりの数を絞る
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(5 * time.Minute)
	// Redis設定
	rdb := redis.NewClient(cfg.Redis)

	// 画像の置き場所。R2 の設定が無ければ手元のフォルダに置いて /images/* で配る
	var images domain.ImageStore
	if cfg.Images.R2Bucket != "" {
		images = storage.NewR2(cfg.Images.R2AccountID, cfg.Images.R2Bucket, cfg.Images.R2KeyID, cfg.Images.R2Secret, cfg.Images.BaseURL)
	} else {
		images = storage.NewLocal(cfg.Images.UploadDir, cfg.Images.BaseURL)
		e.Static("/images", cfg.Images.UploadDir)
		log.Printf("画像は %s に置き、%s で配ります", cfg.Images.UploadDir, cfg.Images.BaseURL)
	}

	// 依存関係注入
	stageRepo := repository.NewStageRepository(db)
	stageUsecase := usecase.NewStageUsecase(stageRepo).WithImages(images)
	stageHandler := handler.NewStageHandler(stageUsecase, time.Now)

	boothRepo := repository.NewBoothRepository(db, rdb)
	boothUsecase := usecase.NewBoothUsecase(boothRepo).WithImages(images)
	boothHandler := handler.NewBoothHandler(boothUsecase)

	userRepo := repository.NewUserRepository(db, rdb)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewUserHandler(userUsecase, []byte(os.Getenv("JWT_SECRET")))
	announceRepo := repository.NewAnnouncementRepository()
	announceUsecase := usecase.NewAnnouncementUsecase(announceRepo)
	announceHandler := handler.NewAnnouncementHandler(announceUsecase)

	likeUsecase := usecase.NewLikeUsecase(repository.NewLikeRepository(db, rdb), boothRepo, cfg.VoterSecret, time.Now)
	likeHandler := handler.NewLikeHandler(likeUsecase)
	imageHandler := handler.NewImageHandler(usecase.NewImageUsecase(images, boothRepo, stageRepo))

	// ハンドラをまとめてルーターに渡す(ここもっと良くなるかも)
	handlers := &handler.Handlers{
		Stage:        stageHandler,
		Booth:        boothHandler,
		User:         userHandler,
		Announcement: announceHandler,
		Like:         likeHandler,
		Image:        imageHandler,
	}

	router.InitRoutes(e, handlers)

	// adminユーザの追加
	adminID := os.Getenv("INIT_ADMIN_ID")
	adminPassword := os.Getenv("INIT_ADMIN_PASSWORD")
	if adminID != "" && adminPassword != "" {
		if err := userUsecase.InitAdminUser(context.Background(), adminID, []byte(adminPassword)); err != nil {
			log.Printf("Failed to initialize admin user: %v", err)
		} else {
			log.Println("Admin initialization check completed.")
		}
	}

	// サーバーの起動
	if err := e.Start(":" + cfg.Port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
