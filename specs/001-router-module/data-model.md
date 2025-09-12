# データモデル: ルーターモジュール

## 1. ルート設定 (RouteConfig)

```go
type RouteConfig struct {
    ID          string            `json:"id" yaml:"id"`
    Path        string            `json:"path" yaml:"path"`                 // URLパスパターン (例: /api/users/*)
    Method      []string          `json:"method" yaml:"method"`             // HTTPメソッド (GET, POST等)
    Backend     string            `json:"backend" yaml:"backend"`           // バックエンドサービスID
    Timeout     time.Duration     `json:"timeout" yaml:"timeout"`           // リクエストタイムアウト
    RateLimit   *RateLimitConfig  `json:"rate_limit" yaml:"rate_limit"`     // レート制限設定
    Auth        *AuthConfig       `json:"auth" yaml:"auth"`                 // 認証設定
    Middleware  []string          `json:"middleware" yaml:"middleware"`     // 適用するミドルウェアID
    Priority    int               `json:"priority" yaml:"priority"`         // ルート優先度
    Enabled     bool              `json:"enabled" yaml:"enabled"`           // 有効/無効フラグ
    CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}
```

### 検証ルール
- `ID`: 必須、一意
- `Path`: 必須、有効なURLパターン
- `Method`: 必須、有効なHTTPメソッド
- `Backend`: 必須、存在するバックエンドサービスID
- `Timeout`: デフォルト30秒、最大5分
- `Priority`: 0-1000の範囲

## 2. バックエンドサービス (BackendService)

```go
type BackendService struct {
    ID             string                `json:"id" yaml:"id"`
    Name           string                `json:"name" yaml:"name"`
    Endpoints      []EndpointConfig      `json:"endpoints" yaml:"endpoints"`       // サービスエンドポイント
    LoadBalancer   LoadBalancerConfig    `json:"load_balancer" yaml:"load_balancer"`
    HealthCheck    HealthCheckConfig     `json:"health_check" yaml:"health_check"`
    CircuitBreaker CircuitBreakerConfig  `json:"circuit_breaker" yaml:"circuit_breaker"`
    RetryPolicy    RetryPolicyConfig     `json:"retry_policy" yaml:"retry_policy"`
    Enabled        bool                  `json:"enabled" yaml:"enabled"`
    CreatedAt      time.Time             `json:"created_at" yaml:"created_at"`
    UpdatedAt      time.Time             `json:"updated_at" yaml:"updated_at"`
}

type EndpointConfig struct {
    URL      string  `json:"url" yaml:"url"`           // http://service:port
    Weight   int     `json:"weight" yaml:"weight"`     // 負荷分散の重み
    Healthy  bool    `json:"healthy" yaml:"healthy"`   // 健全性ステータス
    Metadata map[string]string `json:"metadata" yaml:"metadata"`
}

type LoadBalancerConfig struct {
    Algorithm string `json:"algorithm" yaml:"algorithm"` // round-robin, weighted, least-conn
    StickySession bool `json:"sticky_session" yaml:"sticky_session"`
}
```

### 検証ルール
- `ID`: 必須、一意
- `Name`: 必須、255文字以内
- `Endpoints`: 最低1つのエンドポイント必須
- `URL`: 有効なHTTP/HTTPS URL
- `Weight`: 1-100の範囲

## 3. ヘルスチェック設定 (HealthCheckConfig)

```go
type HealthCheckConfig struct {
    Enabled       bool          `json:"enabled" yaml:"enabled"`
    Path          string        `json:"path" yaml:"path"`           // ヘルスチェックパス
    Interval      time.Duration `json:"interval" yaml:"interval"`   // チェック間隔
    Timeout       time.Duration `json:"timeout" yaml:"timeout"`     // タイムアウト
    HealthyThreshold   int      `json:"healthy_threshold" yaml:"healthy_threshold"`     // 健全と判定する成功回数
    UnhealthyThreshold int      `json:"unhealthy_threshold" yaml:"unhealthy_threshold"` // 不健全と判定する失敗回数
    ExpectedStatus     []int    `json:"expected_status" yaml:"expected_status"`         // 期待するステータスコード
}
```

### デフォルト値
- `Path`: /health
- `Interval`: 30秒
- `Timeout`: 5秒
- `HealthyThreshold`: 2
- `UnhealthyThreshold`: 3
- `ExpectedStatus`: [200]

## 4. サーキットブレーカー設定 (CircuitBreakerConfig)

```go
type CircuitBreakerConfig struct {
    Enabled         bool          `json:"enabled" yaml:"enabled"`
    MaxRequests     uint32        `json:"max_requests" yaml:"max_requests"`         // ハーフオープン状態での最大リクエスト数
    Interval        time.Duration `json:"interval" yaml:"interval"`                 // カウンタリセット間隔
    Timeout         time.Duration `json:"timeout" yaml:"timeout"`                   // オープン状態のタイムアウト
    FailureRatio    float64       `json:"failure_ratio" yaml:"failure_ratio"`       // 失敗率閾値
    MinimumRequests uint32        `json:"minimum_requests" yaml:"minimum_requests"` // 判定に必要な最小リクエスト数
}
```

### 状態遷移
- **Closed** → **Open**: 失敗率が閾値を超えた場合
- **Open** → **Half-Open**: タイムアウト経過後
- **Half-Open** → **Closed**: 成功した場合
- **Half-Open** → **Open**: 失敗した場合

## 5. レート制限設定 (RateLimitConfig)

```go
type RateLimitConfig struct {
    Enabled     bool   `json:"enabled" yaml:"enabled"`
    Rate        int    `json:"rate" yaml:"rate"`               // リクエスト数
    Period      string `json:"period" yaml:"period"`           // 期間 (second, minute, hour)
    BurstSize   int    `json:"burst_size" yaml:"burst_size"`   // バーストサイズ
    KeyType     string `json:"key_type" yaml:"key_type"`       // IP, API_KEY, USER_ID
    WhiteList   []string `json:"white_list" yaml:"white_list"`   // ホワイトリスト
}
```

## 6. 認証設定 (AuthConfig)

```go
type AuthConfig struct {
    Enabled  bool     `json:"enabled" yaml:"enabled"`
    Type     string   `json:"type" yaml:"type"`       // none, basic, bearer, api-key
    Required bool     `json:"required" yaml:"required"`
    Roles    []string `json:"roles" yaml:"roles"`     // 必要なロール
}
```

## 7. メトリクス (Metrics)

```go
type RequestMetrics struct {
    RequestID    string        `json:"request_id"`
    Method       string        `json:"method"`
    Path         string        `json:"path"`
    StatusCode   int           `json:"status_code"`
    Duration     time.Duration `json:"duration"`
    BackendID    string        `json:"backend_id"`
    Error        string        `json:"error,omitempty"`
    Timestamp    time.Time     `json:"timestamp"`
}

type SystemMetrics struct {
    ActiveConnections  int64   `json:"active_connections"`
    TotalRequests      int64   `json:"total_requests"`
    FailedRequests     int64   `json:"failed_requests"`
    AverageLatency     float64 `json:"average_latency_ms"`
    P95Latency         float64 `json:"p95_latency_ms"`
    P99Latency         float64 `json:"p99_latency_ms"`
    MemoryUsage        int64   `json:"memory_usage_bytes"`
    CPUUsage           float64 `json:"cpu_usage_percent"`
    Timestamp          time.Time `json:"timestamp"`
}
```

## 8. 設定ファイル全体構造

```yaml
version: "1.0"
router:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
  max_header_bytes: 1048576

backends:
  - id: api-service
    name: "API Service"
    endpoints:
      - url: "http://api-1:8081"
        weight: 50
      - url: "http://api-2:8081"
        weight: 50
    load_balancer:
      algorithm: weighted
    health_check:
      enabled: true
      path: /health
      interval: 30s
    circuit_breaker:
      enabled: true
      failure_ratio: 0.6

routes:
  - id: api-route
    path: "/api/*"
    method: ["GET", "POST", "PUT", "DELETE"]
    backend: api-service
    timeout: 30s
    rate_limit:
      enabled: true
      rate: 100
      period: minute
    auth:
      enabled: true
      type: bearer
      required: true

middleware:
  logging:
    enabled: true
    level: info
  cors:
    enabled: true
    allowed_origins: ["*"]
  compression:
    enabled: true
    level: 5
```

## データ永続化

設定データは以下の方法で管理されます：

1. **プライマリ**: YAMLファイル（config.yaml）
2. **ランタイム**: メモリ内キャッシュ
3. **バックアップ**: 定期的なスナップショット

## 状態管理

ルーターは以下の状態を管理します：

1. **ルーティングテーブル**: アクティブなルート設定
2. **バックエンドプール**: 利用可能なバックエンドサービス
3. **ヘルスステータス**: 各エンドポイントの健全性
4. **サーキットブレーカー状態**: 各バックエンドのCB状態
5. **メトリクスバッファ**: 最近のリクエストメトリクス