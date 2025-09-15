# ryohi_sub_cal2 Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-09-12

## Active Technologies
- Go 1.23.0 (001-router-module)
- gorilla/mux - HTTPルーティング
- prometheus/client_golang - メトリクス収集
- sony/gobreaker - サーキットブレーカー
- viper - 設定管理
- fsnotify - ファイル監視

## Project Structure
```
src/
├── models/          # データモデル定義
├── services/        # ビジネスロジック
│   ├── router/      # ルーティングサービス
│   ├── proxy/       # プロキシサービス
│   └── health/      # ヘルスチェックサービス
├── cli/             # CLIコマンド
└── lib/             # 共有ライブラリ
    ├── middleware/  # HTTPミドルウェア
    └── config/      # 設定管理

tests/
├── contract/        # API契約テスト
├── integration/     # 統合テスト
└── unit/           # ユニットテスト
```

## Commands
```bash
# ビルド
go build -o router cmd/router/main.go

# テスト実行
go test ./...

# 開発サーバー起動
go run cmd/router/main.go
```

## Code Style
Go: 標準のgofmt、golangci-lint使用

## Recent Changes
- 001-router-module: ルーターモジュール初期実装

<!-- MANUAL ADDITIONS START -->

## 外部モジュール統合

### DTako Module (v1.4.0)
- 自動Swagger統合済み
- `/dtako`プレフィックスでルーティング
- `src/services/dtako/`でサービス管理

### ETC Meisai Module (v0.0.3)
- ETC利用明細の自動取得・管理
- `/etc_meisai`プレフィックスでルーティング
- `src/services/etc_meisai/`でサービス管理

## モジュール統合手順

新しい外部モジュールを統合する場合：

1. **サービス作成** (`src/services/{module_name}/`)
   ```go
   // {module_name}_service.go
   type {ModuleName}Service struct {
       enabled bool
   }

   func (s *{ModuleName}Service) RegisterRoutes(router *mux.Router) {
       adapters.AdaptChiToMux(router, "/{module_name}", func(r chi.Router) {
           // モジュールのルート登録
       })
   }
   ```

2. **Swagger統合** (`src/lib/swagger/merger.go`)
   ```go
   var integratedModules = []ModuleConfig{
       // 既存モジュール...
       {
           Name:       "{module_name}",
           SwaggerURL: "https://raw.githubusercontent.com/{org}/{repo}/master/docs/swagger.json",
           PathPrefix: "/{module_name}",
       },
   }
   ```

3. **サーバー統合** (`src/server/server.go`)
   - サービスインスタンスをServer構造体に追加
   - New関数で初期化
   - setupMainRouterでルート登録

4. **バージョン管理**
   - 可能な場合、モジュールバージョンの自動取得機能を実装
   - Swaggerドキュメントに反映

## 注意事項

- 外部モジュールのSwagger URLが404の場合は手動で定義作成が必要
- データベース接続が必要なモジュールは環境変数設定を追加
- モジュール固有の認証要件を確認

<!-- MANUAL ADDITIONS END -->