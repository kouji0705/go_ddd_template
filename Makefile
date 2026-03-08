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

# テスト用 DB 接続情報（docker-compose.test.yml と合わせる）
TEST_DB_HOST     := localhost
TEST_DB_PORT     := 5433
TEST_DB_USER     := test
TEST_DB_PASSWORD := test
TEST_DB_NAME     := test_db

# ── デフォルトターゲット ──────────────────────────────────────
.DEFAULT_GOAL := help

# ── ヘルプ ───────────────────────────────────────────────────
.PHONY: help
help: ## このヘルプを表示する
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================
#  開発フロー
# ============================================================

.PHONY: start-dev
start-dev: ## 【開発】DB + Logdy コンテナを起動して準備する
	$(DC) up --build -d db logdy
	@echo "⏳ DB の起動を待機しています..."
	@$(DC) exec db sh -c 'until pg_isready -U user -d dbname; do sleep 1; done'
	@echo "✅ DB の準備が完了しました。'make run' でサーバーを起動できます。"

.PHONY: run
run: ## 【開発】API サーバーをローカルで起動する（make start-dev が前提）
	DB_HOST=$(TEST_DB_HOST) \
	DB_PORT=5432 \
	DB_USER=user \
	DB_PASSWORD=password \
	DB_NAME=dbname \
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
	$(DC) exec db psql -U user -d dbname

.PHONY: db-shell-test
db-shell-test: ## テスト DB の psql シェルに接続する
	$(DC_TEST) exec db-test psql -U test -d test_db

.PHONY: db-reset
db-reset: ## 開発 DB を完全リセットしてコンテナを再作成する
	$(DC) down -v
	$(DC) up --build -d
