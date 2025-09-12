# 研究文書: ルーターモジュール

## 1. gorilla/mux ベストプラクティス

**決定**: gorilla/muxを主要ルーティングライブラリとして使用
**根拠**: 
- 成熟したライブラリで大規模プロダクションで実績あり
- 正規表現ベースのルーティングサポート
- ミドルウェアチェーンの優れたサポート
- 標準のnet/httpと完全互換

**検討された代替案**:
- chi: より軽量だが、gorilla/muxほど機能が豊富でない
- gin: パフォーマンスは高いが、カスタムインターフェースが多い
- 標準library: 複雑なルーティングには不十分

## 2. サーキットブレーカー実装

**決定**: sony/gobreaker ライブラリを使用
**根拠**:
- シンプルで理解しやすいAPI
- 本番環境での実績
- カスタマイズ可能な閾値設定
- Go標準のcontext対応

**実装パターン**:
```go
// 各バックエンドサービスごとにサーキットブレーカーを作成
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "backend-service",
    MaxRequests: 3,
    Interval:    time.Minute,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 3 && failureRatio >= 0.6
    },
})
```

**検討された代替案**:
- 自前実装: 複雑でバグが入りやすい
- hystrix-go: Netflixプロジェクトだがメンテナンスが停止

## 3. 設定ホットリロード

**決定**: fsnotifyとviperの組み合わせ
**根拠**:
- viperは設定管理のデファクトスタンダード
- fsnotifyでファイル変更を効率的に監視
- YAML/JSON/TOML等複数フォーマット対応
- 環境変数との統合が容易

**実装パターン**:
```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    // 新しい設定を適用
    reloadRoutes()
})
```

**検討された代替案**:
- ポーリング: リソース消費が大きい
- シグナルベース: ファイル変更の自動検出なし
- etcd/consul: 小規模プロジェクトには過剰

## 4. Prometheusメトリクス統合

**決定**: prometheus/client_golang公式ライブラリ
**根拠**:
- Prometheusプロジェクト公式
- 標準的なメトリクスタイプ完全サポート
- HTTPハンドラーが組み込み済み
- ヒストグラムとサマリーの効率的な実装

**メトリクス設計**:
```go
// 主要メトリクス
- http_requests_total (Counter): リクエスト総数
- http_request_duration_seconds (Histogram): レスポンス時間
- http_requests_in_flight (Gauge): 処理中リクエスト数
- backend_health_status (Gauge): バックエンド健全性
```

**検討された代替案**:
- OpenTelemetry: より汎用的だが設定が複雑
- StatsD: Prometheusエコシステムとの統合が弱い

## 5. プロキシ実装

**決定**: httputil.ReverseProxyを基盤として使用
**根拠**:
- Go標準ライブラリで安定性が高い
- カスタマイズ可能なDirector関数
- エラーハンドリングのフック
- WebSocketサポート

**カスタマイズポイント**:
```go
proxy := &httputil.ReverseProxy{
    Director: customDirector,
    ModifyResponse: addHeaders,
    ErrorHandler: handleProxyError,
}
```

**検討された代替案**:
- 完全自前実装: 車輪の再発明
- サードパーティプロキシライブラリ: 標準ライブラリで十分

## 6. 負荷分散アルゴリズム

**決定**: ラウンドロビンとWeighted Round Robin実装
**根拠**:
- シンプルで予測可能
- 状態管理が最小限
- 重み付けで柔軟性を確保

**実装アプローチ**:
```go
type LoadBalancer interface {
    Next() *Backend
    MarkHealthy(backend *Backend)
    MarkUnhealthy(backend *Backend)
}
```

**検討された代替案**:
- Least Connections: 接続追跡のオーバーヘッド
- IP Hash: セッション永続性が不要な場合は過剰
- Random: 分散が不均等になる可能性

## 7. ログとトレーシング

**決定**: 構造化ログにslog（Go 1.21+標準）を使用
**根拠**:
- Go標準ライブラリ
- 構造化ログネイティブサポート
- パフォーマンスが優秀
- コンテキスト伝播が容易

**相関ID実装**:
```go
// ミドルウェアでリクエストIDを生成
requestID := uuid.New().String()
ctx := context.WithValue(r.Context(), "request_id", requestID)
logger := slog.With("request_id", requestID)
```

**検討された代替案**:
- zerolog: 外部依存
- zap: 設定が複雑
- logrus: メンテナンスモード

## まとめ

すべての技術選択は以下の原則に基づいています：
1. **シンプリシティ**: 可能な限り標準ライブラリを使用
2. **信頼性**: 本番環境で実績のあるライブラリを選択
3. **保守性**: アクティブにメンテナンスされているプロジェクト
4. **パフォーマンス**: 高スループット要件を満たす実装

これらの決定により、堅牢で拡張可能なルーターモジュールの構築が可能になります。