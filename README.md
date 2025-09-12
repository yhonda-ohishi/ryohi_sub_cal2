# Ryohi Router

高性能なHTTPリバースプロキシ/ロードバランサー

## 特徴

- 🚀 **高性能ルーティング** - gorilla/muxベースの高速ルーティング
- ⚖️ **ロードバランシング** - ラウンドロビン、重み付け、最少接続アルゴリズム
- 🔄 **サーキットブレーカー** - 障害のあるバックエンドを自動的に隔離
- 📊 **Prometheusメトリクス** - 詳細な監視とアラート
- 🔒 **認証/認可** - API Key、JWT、OAuth2サポート
- ⚡ **レート制限** - トークンバケットアルゴリズムによる制限
- 🔧 **設定の自動リロード** - fsnotifyによる設定変更の自動検出
- 🏥 **ヘルスチェック** - バックエンドの定期的な健康診断
- 🛡️ **ミドルウェア** - CORS、圧縮、セキュリティヘッダー

## クイックスタート

### インストール

```bash
go get github.com/your-org/ryohi-router
```

### 設定

`configs/config.yaml`を作成:

```yaml
version: "1.0"

router:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

backends:
  - id: api-service
    name: API Service
    endpoints:
      - url: http://localhost:3000
        weight: 1
    load_balancer:
      algorithm: round-robin

routes:
  - id: api-route
    path: /api
    method: ["GET", "POST", "PUT", "DELETE"]
    backend: api-service
    enabled: true
```

### 実行

```bash
go run cmd/router/main.go
```

または、ビルドして実行:

```bash
go build -o router cmd/router/main.go
./router
```

## API

### メインルーター (ポート8080)

- `GET /health` - ヘルスチェック

### 管理API (ポート8081)

認証: `X-API-Key`ヘッダーが必要

- `GET /admin/routes` - すべてのルートを取得
- `POST /admin/routes` - 新しいルートを作成
- `GET /admin/routes/{id}` - 特定のルートを取得
- `PUT /admin/routes/{id}` - ルートを更新
- `DELETE /admin/routes/{id}` - ルートを削除
- `POST /admin/reload` - 設定をリロード

### メトリクス (ポート9090)

- `GET /metrics` - Prometheusメトリクス

## 開発

### 必要条件

- Go 1.23以上
- Make (オプション)

### ビルド

```bash
make build
```

### テスト

```bash
make test
```

### リント

```bash
make lint
```

## アーキテクチャ

```
src/
├── api/          # APIハンドラー
├── lib/          # 共有ライブラリ
│   ├── config/   # 設定管理
│   └── middleware/ # HTTPミドルウェア
├── models/       # データモデル
├── server/       # HTTPサーバー
└── services/     # ビジネスロジック
    ├── health/   # ヘルスチェック
    ├── loadbalancer/ # ロードバランサー
    └── router/   # ルーティング
```

## 設定オプション

### ロードバランサーアルゴリズム

- `round-robin` - ラウンドロビン
- `weighted` - 重み付けラウンドロビン
- `least-connections` - 最少接続
- `ip-hash` - IPハッシュ

### レート制限

```yaml
rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200
```

### サーキットブレーカー

```yaml
circuit_breaker:
  enabled: true
  max_requests: 3
  failure_ratio: 0.6
  timeout: 30s
```

## 環境変数

- `ROUTER_PORT` - ルーターポート (デフォルト: 8080)
- `ADMIN_API_KEY` - 管理APIキー
- `LOG_LEVEL` - ログレベル (debug, info, warn, error)

## ライセンス

MIT

## 貢献

プルリクエストを歓迎します。大きな変更の場合は、まずissueを開いて変更内容を議論してください。