// Package logs はアプリケーション全体で使用するロガーを初期化します。
// Docker の標準出力に JSON 形式で出力し、Logdy がリアルタイムで収集します。
package logs

import (
	"log/slog"
	"os"
)

// Init はグローバルな構造化ロガー（JSON形式）を初期化します。
func Init() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	return logger
}
