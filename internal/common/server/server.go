// Package server はEchoサーバーの初期化とミドルウェア設定を提供します。
package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// ミドルウェア
	e.Use(middleware.RequestID())                         // X-Request-ID の自動付与
	e.Use(middleware.Logger())                            // アクセスログ（構造化）
	e.Use(middleware.Recover())                           // パニックリカバリー
	e.Use(middleware.CORS())                              // CORS
	e.Use(middleware.Secure())                            // セキュリティヘッダー

	return e
}
