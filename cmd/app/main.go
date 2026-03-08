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
	"github.com/kouji/go_ddd_template/internal/workout/command"
	"github.com/kouji/go_ddd_template/internal/workout/controller"
	"github.com/kouji/go_ddd_template/internal/workout/infrastructure"
	"github.com/kouji/go_ddd_template/internal/workout/query"
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

	// 3. Repository（infrastructure 層）初期化 & マイグレーション
	workoutRepo := infrastructure.NewWorkoutRepository(bunDB)
	ctx := context.Background()
	if err := workoutRepo.AutoMigrate(ctx); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("migrations completed")

	// 4. Service 層 DI
	//    CommandService: 書き込み系ユースケース
	//    QueryService:   読み取り系ユースケース
	workoutCommandSvc := command.NewWorkoutCommandService(workoutRepo)
	workoutQuerySvc := query.NewWorkoutQueryService(workoutRepo)

	// 5. Controller 層 DI & ルート登録
	e := server.NewEchoServer()
	workoutCtrl := controller.NewWorkoutController(workoutCommandSvc, workoutQuerySvc)
	workoutCtrl.RegisterRoutes(e)

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
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "1323"
	}
	addr := ":" + port
	logger.Info("server listening", slog.String("addr", addr))
	if err := e.Start(addr); err != nil {
		logger.Info("server stopped", slog.String("reason", err.Error()))
	}
}
