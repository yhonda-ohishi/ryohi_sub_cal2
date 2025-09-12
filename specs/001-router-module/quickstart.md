# クイックスタートガイド: ルーターモジュール

## 前提条件

- Go 1.23.0以上がインストールされていること
- Dockerがインストールされていること（オプション）
- curlまたはPostmanなどのHTTPクライアント

## 1. インストール

### リポジトリのクローン
```bash
git clone https://github.com/your-org/ryohi-router.git
cd ryohi-router
```

### 依存関係のインストール
```bash
go mod download
```

## 2. 設定

### 基本設定ファイルの作成
```bash
cp config.example.yaml config.yaml
```

### 最小限の設定例
```yaml
# config.yaml
version: "1.0"
router:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

backends:
  - id: example-backend
    name: "Example Backend"
    endpoints:
      - url: "http://httpbin.org"
        weight: 100
    health_check:
      enabled: true
      path: /status/200
      interval: 30s

routes:
  - id: example-route
    path: "/test/*"
    method: ["GET", "POST"]
    backend: example-backend
    timeout: 10s
```

## 3. 起動

### 開発モード
```bash
go run cmd/router/main.go
```

### ビルドして実行
```bash
go build -o router cmd/router/main.go
./router
```

### Dockerで実行
```bash
docker build -t ryohi-router .
docker run -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml ryohi-router
```

## 4. 動作確認

### ヘルスチェック
```bash
curl http://localhost:8080/health
```

期待されるレスポンス:
```json
{
  "status": "healthy",
  "timestamp": "2025-09-12T10:00:00Z",
  "services": {
    "example-backend": {
      "status": "healthy",
      "message": "All endpoints operational"
    }
  }
}
```

### テストリクエスト
```bash
# プロキシ経由でリクエスト送信
curl http://localhost:8080/test/get
```

### メトリクス確認
```bash
curl http://localhost:8080/metrics
```

## 5. 管理API操作

### APIキーの設定
```bash
export API_KEY="your-secret-api-key"
```

### ルート一覧取得
```bash
curl -H "X-API-Key: $API_KEY" http://localhost:8080/admin/routes
```

### 新しいルート追加
```bash
curl -X POST -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "new-route",
    "path": "/api/v2/*",
    "method": ["GET"],
    "backend": "example-backend",
    "timeout": "15s"
  }' \
  http://localhost:8080/admin/routes
```

### バックエンドヘルス確認
```bash
curl -H "X-API-Key: $API_KEY" \
  http://localhost:8080/admin/backends/example-backend/health
```

### 設定リロード
```bash
curl -X POST -H "X-API-Key: $API_KEY" \
  http://localhost:8080/admin/reload
```

## 6. 高度な設定

### レート制限の追加
```yaml
routes:
  - id: rate-limited-route
    path: "/api/limited/*"
    method: ["GET"]
    backend: example-backend
    rate_limit:
      enabled: true
      rate: 10
      period: minute
      key_type: IP
```

### サーキットブレーカーの設定
```yaml
backends:
  - id: protected-backend
    name: "Protected Backend"
    endpoints:
      - url: "http://api.example.com"
    circuit_breaker:
      enabled: true
      failure_ratio: 0.5
      minimum_requests: 10
      timeout: 30s
```

### 認証の有効化
```yaml
routes:
  - id: secure-route
    path: "/secure/*"
    method: ["GET", "POST"]
    backend: example-backend
    auth:
      enabled: true
      type: bearer
      required: true
```

## 7. テストシナリオ

### シナリオ1: 基本的なルーティング
```bash
# 1. ルーターを起動
./router

# 2. テストリクエストを送信
curl http://localhost:8080/test/get

# 3. レスポンスが正しくプロキシされることを確認
```

### シナリオ2: 負荷分散
```bash
# 1. 複数エンドポイントを設定
# config.yamlを編集して複数のendpointsを追加

# 2. 複数のリクエストを送信
for i in {1..10}; do
  curl http://localhost:8080/test/get
done

# 3. メトリクスで分散を確認
curl http://localhost:8080/metrics | grep backend_requests_total
```

### シナリオ3: サーキットブレーカー
```bash
# 1. 故障するバックエンドを設定
# 2. 複数の失敗リクエストを送信
# 3. サーキットがオープンすることを確認
curl -H "X-API-Key: $API_KEY" \
  http://localhost:8080/admin/backends/example-backend/health
```

## 8. トラブルシューティング

### ポートが使用中の場合
```bash
# 別のポートを指定
PORT=8081 ./router
```

### 設定エラーの場合
```bash
# 設定検証モードで実行
./router --validate-config
```

### ログレベルの変更
```bash
# デバッグログを有効化
LOG_LEVEL=debug ./router
```

## 9. パフォーマンステスト

### Apache Benchでの負荷テスト
```bash
ab -n 10000 -c 100 http://localhost:8080/test/get
```

### wrkでの負荷テスト
```bash
wrk -t12 -c400 -d30s http://localhost:8080/test/get
```

## 10. 本番環境へのデプロイ

### systemdサービスとして
```ini
# /etc/systemd/system/ryohi-router.service
[Unit]
Description=Ryohi Router
After=network.target

[Service]
Type=simple
User=router
WorkingDirectory=/opt/ryohi-router
ExecStart=/opt/ryohi-router/router
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Kubernetesへのデプロイ
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ryohi-router
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ryohi-router
  template:
    metadata:
      labels:
        app: ryohi-router
    spec:
      containers:
      - name: router
        image: ryohi-router:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## まとめ

このクイックスタートガイドで、ルーターモジュールの基本的な使い方を学びました。より詳細な設定や高度な機能については、完全なドキュメントを参照してください。