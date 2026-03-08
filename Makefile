# ============================================================
#  Go DDD Template — Makefile
# ============================================================

# ── 変数 ─────────────────────────────────────────────────────
APP_NAME   := go_ddd_template
BINARY     := ./bin/$(APP_NAME)
MAIN       := ./cmd/app/main.go
MODULE     := github.com/kouji/go_ddd_template

GO         := go
GOTEST     := $(GO) test
GOBUILD    := $(GO) build
GOVET      := $(GO) vet
GOFMT      := gofmt

DC         := docker compose

# ── デフォルトターゲット ──────────────────────────────────────
.DEFAULT_GOAL := help

# ── ヘルプ ───────────────────────────────────────────────────
.PHONY: help
help: ## このヘルプを表示する
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================
#  ローカル開発
# ============================================================

.PHONY: build
build: ## バイナリをビルドする（./bin/go_ddd_template）
	@mkdir -p ./bin
	$(GOBUILD) -o $(BINARY) $(MAIN)

.PHONY: run
run: ## ローカルでサーバーを起動する（要: DB起動済み）
	$(GO) run $(MAIN)

.PHONY: fmt
fmt: ## コードをフォーマットする（gofmt）
	$(GOFMT) -w .

.PHONY: vet
vet: ## 静的解析を実行する（go vet）
	$(GOVET) ./...

.PHONY: lint
lint: fmt vet ## fmt + vet をまとめて実行する

.PHONY: test
test: ## テストを実行する
	$(GOTEST) -v -race ./...

.PHONY: test-cover
test-cover: ## カバレッジ付きでテストを実行し、HTMLレポートを開く
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: tidy
tidy: ## go mod tidy を実行する
	$(GO) mod tidy

.PHONY: clean
clean: ## ビルド成果物・カバレッジファイルを削除する
	rm -rf ./bin coverage.out coverage.html

# ============================================================
#  Docker Compose
# ============================================================

.PHONY: up
up: ## コンテナをビルドして起動する（ホットリロード有効）
	$(DC) up --build

.PHONY: up-d
up-d: ## コンテナをバックグラウンドで起動する
	$(DC) up --build -d

.PHONY: down
down: ## コンテナを停止・削除する
	$(DC) down

.PHONY: down-v
down-v: ## コンテナを停止・削除し、ボリュームも削除する（DB初期化）
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
#  DB
# ============================================================

.PHONY: db-shell
db-shell: ## PostgreSQL の psql シェルに接続する
	$(DC) exec db psql -U user -d dbname

.PHONY: db-reset
db-reset: ## DB ボリュームを削除してコンテナを再作成する（データ全消去）
	$(DC) down -v
	$(DC) up --build -d
