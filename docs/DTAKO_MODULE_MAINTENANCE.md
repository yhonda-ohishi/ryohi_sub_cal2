# DTako Module Maintenance Instructions

## 概要
このドキュメントは、github.com/yhonda-ohishi/dtako_mod パッケージをメンテナンスするClaude向けの指示書です。

## リポジトリ情報
- **リポジトリ**: https://github.com/yhonda-ohishi/dtako_mod
- **パッケージパス**: github.com/yhonda-ohishi/dtako_mod
- **用途**: デジタルタコグラフデータ（車両運行記録）の管理API

## 主要コンポーネント

### 1. ハンドラー (handlers/)
以下の3つの主要ハンドラーが存在します：
- `DtakoRowsHandler` - 車両運行データの管理
- `DtakoEventsHandler` - イベントデータ（位置情報付き）の管理
- `DtakoFerryHandler` - フェリー運航データの管理

各ハンドラーは以下のメソッドを実装する必要があります：
- `List(w http.ResponseWriter, r *http.Request)` - 一覧取得
- `GetByID(w http.ResponseWriter, r *http.Request)` - ID指定取得
- `Import(w http.ResponseWriter, r *http.Request)` - 本番DBからのインポート

## Swagger統合のための必須作業

### 1. Swagger アノテーションの追加
各ハンドラーメソッドに以下の形式でSwaggerアノテーションを追加してください：

```go
// List lists dtako rows
// @Summary      List Dtako Rows
// @Description  Get vehicle operation data with optional date filtering
// @Tags         dtako
// @Accept       json
// @Produce      json
// @Param        from    query     string  false  "Start date (YYYY-MM-DD)"
// @Param        to      query     string  false  "End date (YYYY-MM-DD)"
// @Success      200     {array}   DtakoRow  "List of dtako rows"
// @Failure      400     {object}  ErrorResponse  "Invalid request parameters"
// @Router       /dtako/rows [get]
func (h *DtakoRowsHandler) List(w http.ResponseWriter, r *http.Request) {
    // 実装
}
```

### 2. モデル定義
レスポンスモデルを明確に定義し、Swaggerタグを追加：

```go
type DtakoRow struct {
    ID        string    `json:"id" example:"row-123"`
    VehicleID string    `json:"vehicle_id" example:"vehicle-001"`
    Timestamp time.Time `json:"timestamp" example:"2025-01-13T15:04:05Z"`
    // 他のフィールド
}

type ErrorResponse struct {
    Code    int    `json:"code" example:"400"`
    Message string `json:"message" example:"Invalid request parameters"`
}
```

### 3. ルート登録関数
Chi routerへのルート登録を行う関数を提供：

```go
// RegisterRoutes registers all dtako routes with the given Chi router
func RegisterRoutes(r chi.Router) {
    // Rows endpoints
    rowsHandler := NewDtakoRowsHandler()
    r.Get("/rows", rowsHandler.List)
    r.Get("/rows/{id}", rowsHandler.GetByID)
    r.Post("/rows/import", rowsHandler.Import)
    
    // Events endpoints
    eventsHandler := NewDtakoEventsHandler()
    r.Get("/events", eventsHandler.List)
    r.Get("/events/{id}", eventsHandler.GetByID)
    r.Post("/events/import", eventsHandler.Import)
    
    // Ferry endpoints
    ferryHandler := NewDtakoFerryHandler()
    r.Get("/ferry", ferryHandler.List)
    r.Get("/ferry/{id}", ferryHandler.GetByID)
    r.Post("/ferry/import", ferryHandler.Import)
}
```

### 4. Swagger Docs生成
以下のステップでSwaggerドキュメントを生成：

1. swagツールのインストール（まだの場合）:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. main.goまたはdoc.goファイルにベース情報を追加:
```go
// @title           DTako API
// @version         1.0.0
// @description     Digital tachograph data management API
// @host            localhost:8080
// @BasePath        /dtako
package dtako_mod
```

3. Swaggerドキュメント生成:
```bash
swag init -g doc.go -o docs/
```

4. 生成されたdocsディレクトリをコミット

## 統合チェックリスト

### 必須項目
- [ ] 各ハンドラーメソッドにSwaggerアノテーションを追加
- [ ] レスポンスモデルを定義し、適切なJSONタグとexampleタグを追加
- [ ] エラーレスポンスモデルを定義
- [ ] RegisterRoutes関数を実装
- [ ] swag initでドキュメントを生成
- [ ] docs/ディレクトリをリポジトリにコミット

### 推奨項目
- [ ] リクエストボディモデルを定義（Import用）
- [ ] ページネーションパラメータをサポート
- [ ] ソートパラメータをサポート
- [ ] フィルタリング条件を拡張（車両ID、イベントタイプなど）

## Ryohi Routerとの統合

DTako moduleは以下の方法でRyohi Routerに統合されます：

1. **サービス層での統合** (src/services/dtako/dtako_service.go):
```go
func (s *DtakoService) RegisterRoutes(router *mux.Router) {
    adapters.AdaptChiToMux(router, "/dtako", func(r chi.Router) {
        dtako_mod.RegisterRoutes(r)
    })
}
```

2. **Swaggerドキュメントの統合** (cmd/router/main.go):
```go
import (
    _ "github.com/yhonda-ohishi/dtako_mod/docs"
)
```

## テスト要件

### 単体テスト
各ハンドラーメソッドに対して以下をテスト：
- 正常系レスポンス
- エラーハンドリング
- パラメータバリデーション

### 統合テスト
- エンドポイントの疎通確認
- Swaggerドキュメントの生成確認
- Ryohi Routerとの統合動作確認

## トラブルシューティング

### "no required module provides package" エラー
```bash
go get github.com/yhonda-ohishi/dtako_mod@latest
go mod tidy
```

### Swaggerドキュメントが表示されない
1. `swag init`を実行してdocs/を生成
2. main.goで`_ "github.com/yhonda-ohishi/dtako_mod/docs"`をインポート
3. サーバーを再起動

### Chi RouterとGorilla Muxの互換性問題
AdaptChiToMuxアダプターが正しく実装されていることを確認。パスパラメータの形式が異なることに注意：
- Chi: `/rows/{id}`
- Mux: `/rows/{id}`（同じだが内部処理が異なる）

## 連絡先
問題が発生した場合は、Ryohi Routerプロジェクトのメンテナーに連絡してください。