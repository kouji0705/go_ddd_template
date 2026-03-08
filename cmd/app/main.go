package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	clientdb "github.com/kouji/go_ddd_template/internal/common/client/db"
	"github.com/kouji/go_ddd_template/internal/common/logs"
	"github.com/kouji/go_ddd_template/internal/common/server"
	"github.com/kouji/go_ddd_template/internal/workout/adapters"
	"github.com/kouji/go_ddd_template/internal/workout/app"
	"github.com/kouji/go_ddd_template/internal/workout/ports"
)

func main() {
	// 1. ロガー初期化
	logger := logs.Init()
	logger.Info("starting application")

	// 2. DB 初期化（接続確認つき）
	dbCfg := clientdb.ConfigFromEnv()
	bunDB, err := clientdb.NewBunDB(dbCfg)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer bunDB.Close()
	logger.Info("database connected", slog.String("host", dbCfg.Host))

	// 3. リポジトリ初期化 & マイグレーション
	workoutRepo := adapters.NewWorkoutRepository(bunDB)
	ctx := context.Background()
	if err := workoutRepo.AutoMigrate(ctx); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("migrations completed")

	// 4. アプリケーション層（DI）
	workoutSvc := app.NewService(workoutRepo)

	// 5. HTTP サーバー構築
	e := server.NewEchoServer()
	workoutHandler := ports.NewHTTPHandler(workoutSvc)
	workoutHandler.RegisterRoutes(e)

	// 6. グレースフルシャットダウン
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		logger.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown gracefully", slog.String("error", err.Error()))
		}
	}()

	// 7. サーバー起動
	logger.Info("server listening", slog.String("addr", ":8080"))
	if err := e.Start(":8080"); err != nil {
		logger.Info("server stopped", slog.String("reason", err.Error()))
	}
}
