# Go DDD Template

このプロジェクトは、Go の**ドメイン駆動設計（DDD）**テンプレートです。  
`wild-workouts-go-ddd-example` を参考に、3回の Review & Refactoring サイクルで設計を磨き上げました。

---

## 技術スタック

| 役割 | ライブラリ |
|---|---|
| **Webフレームワーク** | [Echo v4](https://echo.labstack.com/) |
| **ORM** | [Bun](https://bun.uptrace.dev/) |
| **DB** | PostgreSQL 15 |
| **ロガー** | [slog](https://pkg.go.dev/log/slog)（JSON出力）+ [Logdy](https://logdy.dev/)（WebUI） |
| **ホットリロード** | [Air](https://github.com/air-verse/air) |

---

## ディレクトリ構造

```text
.
├── cmd/app/main.go              # エントリーポイント（ワイヤリング / DI）
├── internal/
│   ├── common/
│   │   ├── client/db/           # DB接続・設定（Bun + PostgreSQL）
│   │   ├── logs/                # ロガー初期化（slog JSON）
│   │   └── server/              # Echo + ミドルウェア設定
│   └── workout/                 # ワークアウトドメイン（境界づけられたコンテキスト）
│       ├── domain/              # エンティティ・VO・ドメインエラー（外部依存ゼロ）
│       ├── repository/          # Repository インターフェース定義
│       ├── command/             # CommandService（書き込み系ユースケース）
│       ├── query/               # QueryService（読み取り系ユースケース）
│       ├── infrastructure/      # Repository 実装（永続化モデル分離・Bun）
│       └── controller/          # HTTP Controller（リクエスト解析 / レスポンス変換）
└── pkg/                         # 外部公開ライブラリ（今後用）
```

### データフロー

```
HTTP Request
    ↓
controller/  （リクエストのバインド・バリデーション・レスポンス変換）
    ↓ Command         ↓ Query
command/          query/
（書き込み系UC）    （読み取り系UC）
    ↓                  ↓
repository/WorkoutRepository（インターフェース）
    ↓
infrastructure/  （Bun ORM・永続化モデル）
    ↓
PostgreSQL
```

---

## DDDの設計ポイント（3回のReview＆Refactoring）

### Rev 1 — ドメイン層の強化
- **Value Object** を導入（`WorkoutID`, `WorkoutName`, `Calories`, `Duration`）
- ドメインモデルのフィールドを**非公開**にし、不変条件をファクトリで保証
- `NewWorkout()` / `RestoreWorkout()` でファクトリパターンを明示
- **ドメインエラー**を `errors.New` で型付き定義

### Rev 2 — アプリケーション層の改善（CQRS）
- `CreateWorkoutCommand` を Command オブジェクトとして定義
- `WorkoutDTO` でドメインオブジェクトを外部に漏らさない
- `WorkoutService` / `WorkoutCreator` / `WorkoutReader` のインターフェース分離
- サービスは具象型ではなく**インターフェース**を返す（テスト容易性）

### Rev 3 — インフラ・ポート層の整備
- アダプター層で**永続化モデル（`workoutModel`）とドメインモデルを完全分離**
- `httpError()` でドメインエラー → HTTPステータスの集中マッピング
- レスポンス型（`workoutResponse`）を外部APIスキーマとして分離
- DB接続に `ConfigFromEnv` + `PingContext` による起動時確認を追加
- `db`サービスに `healthcheck` を追加し、`app`が依存順序を保証
- **グレースフルシャットダウン**対応（`SIGINT` / `SIGTERM`）

---

## 起動方法

### Docker（推奨）

```bash
make up        # ビルド＋起動（ホットリロード有効）
make up-d      # バックグラウンド起動
make down      # 停止・削除
make down-v    # 停止・削除＋DBボリューム初期化
```

### ローカル直接起動（DB は別途起動済みが前提）

```bash
make run
```

### その他の便利なコマンド

```bash
make help        # 全コマンド一覧を表示
make lint        # gofmt + go vet
make test        # テスト実行
make test-cover  # カバレッジレポート生成
make logs-app    # app コンテナのログをフォロー
make db-shell    # psql シェルに接続
make db-reset    # DB を完全リセット
```

---

## API

| メソッド | パス | 説明 |
|---|---|---|
| `POST` | `/workouts` | ワークアウト作成 |
| `GET`  | `/workouts` | 一覧取得 |
| `GET`  | `/workouts/:id` | 1件取得 |

### 例

```bash
# 作成
curl -X POST http://localhost:8080/workouts \
  -H "Content-Type: application/json" \
  -d '{"name": "Morning Run", "calories": 300, "duration": 30}'

# 一覧
curl http://localhost:8080/workouts

# 1件取得
curl http://localhost:8080/workouts/<id>
```

---

## ログ確認 (Logdy)

[http://localhost:8081](http://localhost:8081) にアクセスするとリアルタイムログをブラウザで確認できます。  
アプリは JSON 形式（`slog`）で標準出力に書き出し、Docker がキャプチャします。
