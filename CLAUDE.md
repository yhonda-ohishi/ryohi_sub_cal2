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
<!-- MANUAL ADDITIONS END -->