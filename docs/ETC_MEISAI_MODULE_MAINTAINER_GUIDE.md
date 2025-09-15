# ETC_MEISAI モジュール メンテナー向けガイド

## 概要
このドキュメントは、`github.com/yhonda-ohishi/etc_meisai`モジュールのメンテナー向けに、
`ryohi_sub_cal2`プロジェクトとの統合を円滑にするための要件と推奨事項をまとめたものです。

## 必須要件

### 1. ルート登録関数の実装
`ryohi_sub_cal2`との統合のため、以下の関数を実装してください：

```go
package etc_meisai

import (
    "github.com/go-chi/chi/v5"
)

// RegisterRoutes registers all etc_meisai routes with the given Chi router
// この関数は ryohi_sub_cal2 から呼び出されます
func RegisterRoutes(r chi.Router) {
    // ヘルスチェック
    r.Get("/health", HealthHandler)

    // ETC明細インポート
    r.Post("/import", ImportHandler)
    r.Post("/import/csv", ImportCSVHandler)

    // 明細取得
    r.Get("/details", ListDetailsHandler)
    r.Get("/details/{id}", GetDetailHandler)

    // 集計
    r.Get("/summary", GetSummaryHandler)
    r.Get("/summary/monthly", GetMonthlySummaryHandler)

    // スクレイピング
    r.Post("/scrape", ScrapeHandler)
    r.Get("/scrape/status", GetScrapeStatusHandler)
}
```

### 2. Swagger/OpenAPI ドキュメント

#### 2.1 ファイル配置
- **必須パス**: `docs/swagger.json`
- **形式**: OpenAPI 3.0 または Swagger 2.0
- **アクセス**: GitHubのmasterブランチから直接アクセス可能にする

#### 2.2 Swagger定義例
```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "ETC Meisai API",
    "version": "0.0.1",
    "description": "ETC利用明細の自動取得・管理API"
  },
  "servers": [
    {
      "url": "/etc_meisai",
      "description": "Base path for ETC Meisai API"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "summary": "ヘルスチェック",
        "operationId": "getHealth",
        "tags": ["Health"],
        "responses": {
          "200": {
            "description": "サービス正常",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "/import": {
      "post": {
        "summary": "ETC明細データインポート",
        "operationId": "importData",
        "tags": ["Import"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/ImportRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "インポート成功",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ImportResponse"
                }
              }
            }
          }
        }
      }
    },
    "/details": {
      "get": {
        "summary": "明細一覧取得",
        "operationId": "listDetails",
        "tags": ["Details"],
        "parameters": [
          {
            "name": "from",
            "in": "query",
            "schema": {
              "type": "string",
              "format": "date"
            },
            "description": "開始日 (YYYY-MM-DD)"
          },
          {
            "name": "to",
            "in": "query",
            "schema": {
              "type": "string",
              "format": "date"
            },
            "description": "終了日 (YYYY-MM-DD)"
          },
          {
            "name": "trip_number",
            "in": "query",
            "schema": {
              "type": "string"
            },
            "description": "出張番号"
          }
        ],
        "responses": {
          "200": {
            "description": "明細リスト",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/ETCDetail"
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "example": "healthy"
          },
          "version": {
            "type": "string",
            "example": "0.0.1"
          },
          "database": {
            "type": "string",
            "example": "connected"
          }
        }
      },
      "ImportRequest": {
        "type": "object",
        "required": ["data"],
        "properties": {
          "data": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/ETCDetail"
            }
          },
          "overwrite": {
            "type": "boolean",
            "default": false,
            "description": "既存データを上書きするか"
          }
        }
      },
      "ImportResponse": {
        "type": "object",
        "properties": {
          "success": {
            "type": "boolean"
          },
          "imported_count": {
            "type": "integer"
          },
          "skipped_count": {
            "type": "integer"
          },
          "message": {
            "type": "string"
          }
        }
      },
      "ETCDetail": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid"
          },
          "date": {
            "type": "string",
            "format": "date-time"
          },
          "entrance": {
            "type": "string",
            "description": "入口IC"
          },
          "exit": {
            "type": "string",
            "description": "出口IC"
          },
          "amount": {
            "type": "integer",
            "description": "料金（円）"
          },
          "vehicle_number": {
            "type": "string",
            "description": "車両番号"
          },
          "trip_number": {
            "type": "string",
            "description": "出張番号"
          },
          "purpose": {
            "type": "string",
            "description": "利用目的"
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          },
          "updated_at": {
            "type": "string",
            "format": "date-time"
          }
        }
      }
    }
  },
  "tags": [
    {
      "name": "Health",
      "description": "ヘルスチェック"
    },
    {
      "name": "Import",
      "description": "データインポート"
    },
    {
      "name": "Details",
      "description": "明細管理"
    }
  ]
}
```

### 3. エクスポート関数

以下の関数を公開（大文字始まり）にしてください：

```go
// ハンドラー関数（HTTPハンドラーとして使用）
func HealthHandler(w http.ResponseWriter, r *http.Request)
func ImportHandler(w http.ResponseWriter, r *http.Request)
func ImportCSVHandler(w http.ResponseWriter, r *http.Request)
func ListDetailsHandler(w http.ResponseWriter, r *http.Request)
func GetDetailHandler(w http.ResponseWriter, r *http.Request)
func GetSummaryHandler(w http.ResponseWriter, r *http.Request)
func GetMonthlySummaryHandler(w http.ResponseWriter, r *http.Request)
func ScrapeHandler(w http.ResponseWriter, r *http.Request)
func GetScrapeStatusHandler(w http.ResponseWriter, r *http.Request)

// モデル構造体（JSONレスポンス用）
type ETCDetail struct {
    ID            string    `json:"id"`
    Date          time.Time `json:"date"`
    Entrance      string    `json:"entrance"`
    Exit          string    `json:"exit"`
    Amount        int       `json:"amount"`
    VehicleNumber string    `json:"vehicle_number"`
    TripNumber    string    `json:"trip_number,omitempty"`
    Purpose       string    `json:"purpose,omitempty"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}

type ImportRequest struct {
    Data      []ETCDetail `json:"data"`
    Overwrite bool        `json:"overwrite"`
}

type ImportResponse struct {
    Success      bool   `json:"success"`
    ImportedCount int    `json:"imported_count"`
    SkippedCount  int    `json:"skipped_count"`
    Message      string `json:"message"`
}
```

### 4. バージョン管理

#### 4.1 セマンティックバージョニング
- メジャー: 破壊的変更
- マイナー: 後方互換性のある機能追加
- パッチ: バグ修正

#### 4.2 バージョン取得関数（推奨）
```go
// GetVersion returns the current module version
func GetVersion() string {
    return "0.0.1"  // or read from embedded variable
}
```

### 5. データベース設定

#### 5.1 環境変数
以下の環境変数をサポートしてください：

```bash
# データベース接続
ETC_MEISAI_DB_HOST=localhost
ETC_MEISAI_DB_PORT=3306
ETC_MEISAI_DB_NAME=etc_meisai
ETC_MEISAI_DB_USER=root
ETC_MEISAI_DB_PASSWORD=password

# スクレイピング設定（オプション）
ETC_MEISAI_SCRAPE_ENABLED=true
ETC_MEISAI_SCRAPE_USER=user@example.com
ETC_MEISAI_SCRAPE_PASSWORD=password
```

#### 5.2 初期化関数（推奨）
```go
// Initialize sets up database connection and configurations
func Initialize(config *Config) error {
    // データベース接続
    // マイグレーション実行
    // 設定の検証
    return nil
}

type Config struct {
    DBHost     string
    DBPort     int
    DBName     string
    DBUser     string
    DBPassword string
    // その他の設定
}
```

### 6. エラーハンドリング

#### 6.1 標準エラーレスポンス
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}
```

#### 6.2 HTTPステータスコード
- 200: 成功
- 400: リクエスト不正
- 401: 認証エラー
- 404: リソース未発見
- 500: サーバーエラー

### 7. テスト

#### 7.1 単体テスト
```bash
go test ./...
```

#### 7.2 統合テストのサポート
モックモード or テストデータベースのサポート：
```go
// SetTestMode enables test mode with mock data
func SetTestMode(enabled bool) {
    // テストモードの設定
}
```

## 推奨事項

### 1. ログ出力
構造化ログ（JSON形式）の使用を推奨：
```go
log.Printf(`{"level":"info","module":"etc_meisai","message":"Import completed","count":%d}`, count)
```

### 2. メトリクス
Prometheusメトリクスのエクスポート（オプション）：
- `etc_meisai_import_total` - インポート総数
- `etc_meisai_scrape_duration_seconds` - スクレイピング所要時間
- `etc_meisai_db_connections` - DB接続数

### 3. ヘルスチェック
`/health`エンドポイントで以下を返す：
- サービス状態
- データベース接続状態
- 最終スクレイピング時刻（該当する場合）

### 4. ドキュメント
READMEに以下を記載：
- API仕様へのリンク
- 環境変数一覧
- セットアップ手順
- 統合例

## 統合確認チェックリスト

- [ ] `RegisterRoutes`関数が実装されている
- [ ] `docs/swagger.json`が存在する
- [ ] 主要なハンドラー関数が公開されている
- [ ] 環境変数による設定が可能
- [ ] エラーレスポンスが標準化されている
- [ ] ヘルスチェックエンドポイントが動作する
- [ ] テストが通る

## 連絡先

統合に関する質問は以下までお願いします：
- プロジェクト: ryohi_sub_cal2
- 統合担当: [担当者名]
- 連絡方法: [メール/Slack/GitHub Issues]

## 変更履歴

- 2025-09-15: 初版作成
- モジュールバージョン v0.0.1 対応
- 2025-09-15: v0.0.2へアップデート