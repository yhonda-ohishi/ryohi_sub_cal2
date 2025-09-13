# Tasks: ルーターモジュール

**入力**: `/specs/001-router-module/` の設計ドキュメント
**前提条件**: plan-ja.md (必須), research.md, data-model.md, contracts/openapi.yaml

## 実行フロー (main)
```
1. plan-ja.mdから実装計画を読み込み
   → 技術スタック: Go 1.23.0, gorilla/mux, prometheus, gobreaker
   → 構造: src/, tests/ at repository root
2. 設計ドキュメントを読み込み:
   → data-model.md: 7つのエンティティ抽出
   → contracts/openapi.yaml: 11エンドポイント抽出
   → quickstart.md: 3つのテストシナリオ抽出
3. カテゴリ別にタスクを生成:
   → セットアップ: プロジェクト初期化、依存関係
   → テスト: 契約テスト、統合テスト
   → コア実装: モデル、サービス、CLIコマンド
   → 統合: ミドルウェア、ログ、メトリクス
   → 仕上げ: ユニットテスト、パフォーマンス、ドキュメント
4. タスクルール適用:
   → 異なるファイル = [P]で並列実行可能
   → 同じファイル = 順次実行
   → テスト優先実装 (TDD)
5. タスク番号付け (T001, T002...)
6. 依存関係グラフ生成
7. 並列実行例の作成
8. タスク完全性の検証完了
```

## フォーマット: `[ID] [P?] 説明`
- **[P]**: 並列実行可能（異なるファイル、依存関係なし）
- 説明には正確なファイルパスを含む

## パス規約
- **プロジェクト構造**: リポジトリルートに `src/`, `tests/`
- すべてのパスは絶対パスで記載

## フェーズ 3.1: セットアップ
- [ ] T001 プロジェクト構造の作成（src/, tests/, cmd/, configs/）
- [ ] T002 Go モジュールの初期化とgo.modの作成
- [ ] T003 [P] 依存関係のインストール (gorilla/mux, prometheus/client_golang, sony/gobreaker, viper, testify)
- [ ] T004 [P] Makefileとビルドスクリプトの作成
- [ ] T005 [P] .gitignoreと開発環境設定ファイルの作成
- [ ] T006 [P] golangci-lintの設定とlint設定ファイルの作成

## フェーズ 3.2: テスト優先 (TDD) ⚠️ 3.3の前に必ず完了
**重要: これらのテストは実装前に作成し、必ず失敗することを確認**

### 契約テスト
- [ ] T007 [P] GET /health エンドポイントの契約テスト作成 (tests/contract/health_test.go)
- [ ] T008 [P] GET /metrics エンドポイントの契約テスト作成 (tests/contract/metrics_test.go)
- [ ] T009 [P] GET /admin/routes エンドポイントの契約テスト作成 (tests/contract/admin_routes_test.go)
- [ ] T010 [P] POST /admin/routes エンドポイントの契約テスト作成 (tests/contract/admin_routes_post_test.go)
- [ ] T011 [P] GET /admin/routes/{id} エンドポイントの契約テスト作成 (tests/contract/admin_route_get_test.go)
- [ ] T012 [P] PUT /admin/routes/{id} エンドポイントの契約テスト作成 (tests/contract/admin_route_put_test.go)
- [ ] T013 [P] DELETE /admin/routes/{id} エンドポイントの契約テスト作成 (tests/contract/admin_route_delete_test.go)
- [ ] T014 [P] GET /admin/backends エンドポイントの契約テスト作成 (tests/contract/admin_backends_test.go)
- [ ] T015 [P] POST /admin/backends エンドポイントの契約テスト作成 (tests/contract/admin_backends_post_test.go)
- [ ] T016 [P] GET /admin/backends/{id}/health エンドポイントの契約テスト作成 (tests/contract/backend_health_test.go)
- [ ] T017 [P] POST /admin/reload エンドポイントの契約テスト作成 (tests/contract/admin_reload_test.go)

### 統合テスト
- [ ] T018 [P] 基本的なルーティング機能の統合テスト作成 (tests/integration/routing_test.go)
- [ ] T019 [P] 負荷分散機能の統合テスト作成 (tests/integration/loadbalancer_test.go)
- [ ] T020 [P] サーキットブレーカー機能の統合テスト作成 (tests/integration/circuit_breaker_test.go)
- [ ] T021 [P] ヘルスチェック機能の統合テスト作成 (tests/integration/health_check_test.go)
- [ ] T022 [P] 設定ホットリロード機能の統合テスト作成 (tests/integration/config_reload_test.go)

## フェーズ 3.3: コア実装（テストが失敗することを確認後のみ）

### データモデル
- [ ] T023 [P] RouteConfig構造体の実装 (src/models/route.go)
- [ ] T024 [P] BackendService構造体の実装 (src/models/backend.go)
- [ ] T025 [P] HealthCheckConfig構造体の実装 (src/models/health.go)
- [ ] T026 [P] CircuitBreakerConfig構造体の実装 (src/models/circuit_breaker.go)
- [ ] T027 [P] RateLimitConfig構造体の実装 (src/models/rate_limit.go)
- [ ] T028 [P] AuthConfig構造体の実装 (src/models/auth.go)
- [ ] T029 [P] Metrics構造体の実装 (src/models/metrics.go)

### サービス層
- [ ] T030 ルーターサービスの実装 (src/services/router/router.go)
- [ ] T031 プロキシサービスの実装 (src/services/proxy/proxy.go)
- [ ] T032 ヘルスチェックサービスの実装 (src/services/health/health.go)
- [ ] T033 負荷分散サービスの実装 (src/services/loadbalancer/loadbalancer.go)
- [ ] T034 サーキットブレーカーサービスの実装 (src/services/circuit/circuit.go)

### ライブラリ
- [ ] T035 [P] 設定管理ライブラリの実装 (src/lib/config/config.go)
- [ ] T036 [P] ログミドルウェアの実装 (src/lib/middleware/logging.go)
- [ ] T037 [P] 認証ミドルウェアの実装 (src/lib/middleware/auth.go)
- [ ] T038 [P] レート制限ミドルウェアの実装 (src/lib/middleware/rate_limit.go)
- [ ] T039 [P] メトリクス収集ミドルウェアの実装 (src/lib/middleware/metrics.go)
- [ ] T040 [P] CORSミドルウェアの実装 (src/lib/middleware/cors.go)

### APIハンドラー
- [ ] T041 ヘルスチェックハンドラーの実装 (src/api/health.go)
- [ ] T042 メトリクスハンドラーの実装 (src/api/metrics.go)
- [ ] T043 ルート管理ハンドラーの実装 (src/api/routes.go)
- [ ] T044 バックエンド管理ハンドラーの実装 (src/api/backends.go)
- [ ] T045 設定リロードハンドラーの実装 (src/api/reload.go)

### CLIコマンド
- [ ] T046 [P] ルーター起動コマンドの実装 (src/cli/router/start.go)
- [ ] T047 [P] 設定検証コマンドの実装 (src/cli/router/validate.go)
- [ ] T048 [P] ヘルスチェックコマンドの実装 (src/cli/router/health.go)
- [ ] T049 [P] メトリクス表示コマンドの実装 (src/cli/router/metrics.go)

### メインアプリケーション
- [ ] T050 メインエントリーポイントの実装 (cmd/router/main.go)
- [ ] T051 HTTPサーバーの初期化とルーティング設定 (src/server.go)

## フェーズ 3.4: 統合

- [ ] T052 設定ファイルローダーの統合 (src/services/config_loader.go)
- [ ] T053 Prometheusメトリクスコレクターの統合 (src/services/metrics_collector.go)
- [ ] T054 fsnotifyによる設定監視の統合 (src/services/config_watcher.go)
- [ ] T055 すべてのミドルウェアのチェーン設定 (src/middleware_chain.go)
- [ ] T056 グレースフルシャットダウンの実装 (src/shutdown.go)
- [ ] T057 エラーハンドリングとリカバリーの実装 (src/error_handler.go)

## フェーズ 3.5: 仕上げ

### ユニットテスト
- [ ] T058 [P] RouteConfig検証ロジックのユニットテスト (tests/unit/models/route_test.go)
- [ ] T059 [P] BackendService検証ロジックのユニットテスト (tests/unit/models/backend_test.go)
- [ ] T060 [P] 負荷分散アルゴリズムのユニットテスト (tests/unit/services/loadbalancer_test.go)
- [ ] T061 [P] サーキットブレーカー状態遷移のユニットテスト (tests/unit/services/circuit_test.go)
- [ ] T062 [P] レート制限ロジックのユニットテスト (tests/unit/middleware/rate_limit_test.go)

### パフォーマンステスト
- [ ] T063 10,000 req/sの負荷テスト実施 (tests/performance/load_test.go)
- [ ] T064 p99レイテンシ < 100msの検証 (tests/performance/latency_test.go)
- [ ] T065 メモリ使用量 < 100MBの検証 (tests/performance/memory_test.go)

### ドキュメント
- [ ] T066 [P] API仕様書の更新 (docs/api.md)
- [ ] T067 [P] 設定ガイドの作成 (docs/configuration.md)
- [ ] T068 [P] デプロイメントガイドの作成 (docs/deployment.md)
- [ ] T069 [P] トラブルシューティングガイドの作成 (docs/troubleshooting.md)

### 最終確認
- [ ] T070 quickstart.mdのすべてのシナリオを手動実行
- [ ] T071 コードの重複除去とリファクタリング
- [ ] T072 すべてのテストの実行と合格確認
- [ ] T073 ビルドとDockerイメージの作成
- [ ] T074 最終的なlintチェックとフォーマット

## 依存関係
- セットアップ (T001-T006) → すべてのタスク
- テスト (T007-T022) → 実装 (T023-T051)
- モデル (T023-T029) → サービス (T030-T034)
- サービス (T030-T034) → ハンドラー (T041-T045)
- ライブラリ (T035-T040) → 統合 (T052-T057)
- 実装完了 → 仕上げ (T058-T074)

## 並列実行例

### テストフェーズの並列実行
```bash
# T007-T017を同時に起動（契約テスト）:
Task: "GET /health エンドポイントの契約テスト作成 (tests/contract/health_test.go)"
Task: "GET /metrics エンドポイントの契約テスト作成 (tests/contract/metrics_test.go)"
Task: "GET /admin/routes エンドポイントの契約テスト作成 (tests/contract/admin_routes_test.go)"
# ... 他の契約テストも同様
```

### モデル実装の並列実行
```bash
# T023-T029を同時に起動:
Task: "RouteConfig構造体の実装 (src/models/route.go)"
Task: "BackendService構造体の実装 (src/models/backend.go)"
Task: "HealthCheckConfig構造体の実装 (src/models/health.go)"
# ... 他のモデルも同様
```

### ミドルウェア実装の並列実行
```bash
# T035-T040を同時に起動:
Task: "設定管理ライブラリの実装 (src/lib/config/config.go)"
Task: "ログミドルウェアの実装 (src/lib/middleware/logging.go)"
Task: "認証ミドルウェアの実装 (src/lib/middleware/auth.go)"
# ... 他のミドルウェアも同様
```

## 注意事項
- [P]タスク = 異なるファイル、依存関係なし
- テストが失敗することを確認してから実装に進む
- 各タスク完了後にコミット
- 同じファイルを変更する並列タスクは避ける
- TDDサイクル: RED (テスト失敗) → GREEN (実装) → REFACTOR (改善)

## タスク生成ルール
*main()実行中に適用済み*

1. **契約から**:
   - openapi.yaml内の11エンドポイント → 11個の契約テストタスク [P]
   - 各エンドポイント → 対応する実装タスク

2. **データモデルから**:
   - 7つのエンティティ → 7個のモデル作成タスク [P]
   - 関連性 → サービス層タスク

3. **ユーザーストーリーから**:
   - 3つのテストシナリオ → 統合テスト [P]
   - クイックスタートシナリオ → 検証タスク

4. **順序付け**:
   - セットアップ → テスト → モデル → サービス → エンドポイント → 仕上げ
   - 依存関係が並列実行をブロック

## 検証チェックリスト
*GATE: main()がリターン前にチェック済み*

- [x] すべての契約に対応するテストがある (11/11)
- [x] すべてのエンティティにモデルタスクがある (7/7)
- [x] すべてのテストが実装前に来る
- [x] 並列タスクが真に独立している
- [x] 各タスクが正確なファイルパスを指定
- [x] [P]タスクが同じファイルを変更しない

---
**タスク総数**: 74
**並列実行可能**: 43タスク ([P]マーク付き)
**推定完了時間**: 5-7日（2名の開発者で並列作業時）
## フェーズ 5: dtako_mod統合 (T075-T100)
**詳細**: [tasks-dtako-integration.md](tasks-dtako-integration.md)を参照

### 概要
- T075-T077: dtako_mod分析とマッピング
- T078-T082: 統合テスト作成 (TDD)
- T083-T092: dtako_mod実装統合
- T093-T095: 統合テストと検証
- T096-T100: ドキュメントと最終検証

**新規タスク数**: 26タスク
**並列実行可能**: 15タスク
**推定時間**: 2-3日（並列実行時）

---
*dtako_mod integration tasks added - 2025-09-12*
*Total Tasks: 100 (T001-T074 + T075-T100)*
