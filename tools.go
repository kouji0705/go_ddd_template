//go:build tools

package tools

import (
	// golang-migrate CLI ツールをモジュール管理に含めるための blank import
	// postgres ドライバー
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// ファイルソースドライバー
	_ "github.com/golang-migrate/migrate/v4/source/file"
)
