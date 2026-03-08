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
├── cmd/app/main.go          # エントリーポイント（ワイヤリング）
├── internal/
│   ├── common/
│   │   ├── client/db/       # DB接続・設定
│   │   ├── logs/            # ロガー初期化（slog JSON）
│   │   └── server/          # Echo + ミドルウェア設定
│   └── workout/             # ワークアウトドメイン（境界づけられたコンテキスト）
│       ├── domain/          # ★ エンティティ・VO・リポジトリIF・ドメインエラー
│       ├── app/             # ★ ユースケース（Command/Query分離・Service IF）
│       ├── adapters/        # リポジトリ実装（永続化モデル分離）
│       └── ports/           # HTTPハンドラー（エラーマッピング・レスポンス型）
└── pkg/                     # 外部公開ライブラリ（今後用）
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

```bash
docker-compose up
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
