# ============================================================
#  Go DDD Template — Makefile
# ============================================================

# ── 変数 ─────────────────────────────────────────────────────
APP_NAME := go_ddd_template
BINARY   := ./bin/$(APP_NAME)
MAIN     := ./cmd/app/main.go

GO      := go
GOBUILD := $(GO) build
GOTEST  := $(GO) test
GOVET   := $(GO) vet
GOFMT   := gofmt

DC      := docker compose
DC_TEST := docker compose -f docker-compose.test.yml

# ── 環境変数ファイルの読み込み ────────────────────────────────
# .env が存在する場合は開発用の環境変数を読み込む
-include .env
# .env.test が存在する場合はテスト用の環境変数を読み込む
-include .env.test

# フォールバック（.env が存在しない場合のデフォルト値）
DB_HOST       ?= localhost
DB_PORT       ?= 5432
DB_USER       ?= user
DB_PASSWORD   ?= password
DB_NAME       ?= dbname

TEST_DB_HOST     ?= localhost
TEST_DB_PORT     ?= 5433
TEST_DB_USER     ?= test
TEST_DB_PASSWORD ?= test
TEST_DB_NAME     ?= test_db

# ── migrate CLI ───────────────────────────────────────────────
MIGRATE         := migrate
MIGRATIONS_DIR  := ./db/migrations
# 開発 DB の接続 URL
DATABASE_URL    ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
# テスト DB の接続 URL
TEST_DATABASE_URL ?= postgres://$(TEST_DB_USER):$(TEST_DB_PASSWORD)@$(TEST_DB_HOST):$(TEST_DB_PORT)/$(TEST_DB_NAME)?sslmode=disable

# ── デフォルトターゲット ──────────────────────────────────────
.DEFAULT_GOAL := help

# ── ヘルプ ───────────────────────────────────────────────────
.PHONY: help
help: ## このヘルプを表示する
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

# ============================================================
#  開発フロー
# ============================================================

.PHONY: start-dev
start-dev: ## 【開発】DB コンテナを起動して準備する
	$(DC) up --build -d db
	@echo "⏳ DB の起動を待機しています..."
	@$(DC) exec db sh -c 'until pg_isready -U user -d dbname; do sleep 1; done'
	@echo "✅ DB の準備が完了しました。'make run' でサーバーを起動できます。"

.PHONY: run
run: ## 【開発】API サーバーをローカルで起動する（make start-dev が前提）
	DB_HOST=$(DB_HOST) \
	DB_PORT=$(DB_PORT) \
	DB_USER=$(DB_USER) \
	DB_PASSWORD=$(DB_PASSWORD) \
	DB_NAME=$(DB_NAME) \
	$(GO) run $(MAIN)

# ============================================================
#  テストフロー
# ============================================================

.PHONY: start-test
start-test: ## 【テスト】テスト用 DB コンテナを起動して準備する
	$(DC_TEST) up -d --wait
	@echo "✅ テスト用 DB の準備が完了しました。'make test' でテストを実行できます。"

.PHONY: test
test: ## 【テスト】テストを実行する（make start-test が前提）
	DB_HOST=$(TEST_DB_HOST) \
	DB_PORT=$(TEST_DB_PORT) \
	DB_USER=$(TEST_DB_USER) \
	DB_PASSWORD=$(TEST_DB_PASSWORD) \
	DB_NAME=$(TEST_DB_NAME) \
	$(GOTEST) -v -race ./...

.PHONY: test-cover
test-cover: ## 【テスト】カバレッジ付きでテストを実行し HTML レポートを生成する
	DB_HOST=$(TEST_DB_HOST) \
	DB_PORT=$(TEST_DB_PORT) \
	DB_USER=$(TEST_DB_USER) \
	DB_PASSWORD=$(TEST_DB_PASSWORD) \
	DB_NAME=$(TEST_DB_NAME) \
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

.PHONY: stop-test
stop-test: ## 【テスト】テスト用 DB コンテナを停止・削除する
	$(DC_TEST) down -v
	@echo "🗑  テスト用 DB を削除しました。"

# ============================================================
#  コード品質
# ============================================================

.PHONY: fmt
fmt: ## コードをフォーマットする（gofmt）
	$(GOFMT) -w .

.PHONY: vet
vet: ## 静的解析を実行する（go vet）
	$(GOVET) ./...

.PHONY: lint
lint: fmt vet ## fmt + vet をまとめて実行する

.PHONY: tidy
tidy: ## go mod tidy を実行する
	$(GO) mod tidy

# ============================================================
#  ビルド / クリーン
# ============================================================

.PHONY: build
build: ## バイナリをビルドする（./bin/go_ddd_template）
	@mkdir -p ./bin
	$(GOBUILD) -o $(BINARY) $(MAIN)

.PHONY: clean
clean: ## ビルド成果物・カバレッジファイルを削除する
	rm -rf ./bin coverage.out coverage.html

# ============================================================
#  Docker Compose（開発環境フル操作）
# ============================================================

.PHONY: up
up: ## 全コンテナをビルドして起動する（ホットリロード有効）
	$(DC) up --build

.PHONY: down
down: ## 全コンテナを停止・削除する
	$(DC) down

.PHONY: down-v
down-v: ## 全コンテナを停止・削除し、ボリュームも削除する（DB 初期化）
	$(DC) down -v

.PHONY: logs
logs: ## 全コンテナのログをフォローする
	$(DC) logs -f

.PHONY: logs-app
logs-app: ## app コンテナのログをフォローする
	$(DC) logs -f app

.PHONY: ps
ps: ## コンテナの状態を表示する
	$(DC) ps

.PHONY: restart
restart: ## app コンテナだけ再起動する
	$(DC) restart app

# ============================================================
#  DB 操作
# ============================================================

.PHONY: db-shell
db-shell: ## 開発 DB の psql シェルに接続する
	$(DC) exec db psql -U $(DB_USER) -d $(DB_NAME)

.PHONY: db-shell-test
db-shell-test: ## テスト DB の psql シェルに接続する
	$(DC_TEST) exec db-test psql -U $(TEST_DB_USER) -d $(TEST_DB_NAME)

.PHONY: db-reset
db-reset: ## 開発 DB を完全リセットしてコンテナを再作成する
	$(DC) down -v
	$(DC) up --build -d

# ============================================================
#  マイグレーション（golang-migrate）
# ============================================================

# migrate CLI のインストール確認（未インストールならインストール）
.PHONY: migrate-install
migrate-install: ## migrate CLI をインストールする（go install）
	@which migrate > /dev/null 2>&1 || \
	  (echo "⬇️  migrate CLI をインストールしています..." && \
	   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
	   echo "✅ migrate CLI のインストールが完了しました。")

.PHONY: migrate-create
migrate-create: ## マイグレーションファイルを作成する（例: make migrate-create NAME=create_users_table）
	@test -n "$(NAME)" || (echo "❌ NAME を指定してください。例: make migrate-create NAME=create_users_table" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "✅ マイグレーションファイルを作成しました: $(MIGRATIONS_DIR)"

.PHONY: migrate-up
migrate-up: ## 【開発】未適用のマイグレーションをすべて適用する
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "✅ マイグレーション（up）が完了しました。"

.PHONY: migrate-down
migrate-down: ## 【開発】最新のマイグレーションを1つ戻す
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "✅ マイグレーション（down 1）が完了しました。"

.PHONY: migrate-down-all
migrate-down-all: ## 【開発】すべてのマイグレーションを戻す
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all
	@echo "✅ すべてのマイグレーション（down all）が完了しました。"

.PHONY: migrate-version
migrate-version: ## 【開発】現在のマイグレーションバージョンを表示する
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: migrate-force
migrate-force: ## 【開発】マイグレーションバージョンを強制設定する（例: make migrate-force VERSION=1）
	@test -n "$(VERSION)" || (echo "❌ VERSION を指定してください。例: make migrate-force VERSION=1" && exit 1)
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)

# ── テスト DB 向け ────────────────────────────────────────────

.PHONY: migrate-test-up
migrate-test-up: ## 【テスト】未適用のマイグレーションをすべて適用する
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(TEST_DATABASE_URL)" up
	@echo "✅ テスト DB のマイグレーション（up）が完了しました。"

.PHONY: migrate-test-down
migrate-test-down: ## 【テスト】最新のマイグレーションを1つ戻す
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(TEST_DATABASE_URL)" down 1
	@echo "✅ テスト DB のマイグレーション（down 1）が完了しました。"
