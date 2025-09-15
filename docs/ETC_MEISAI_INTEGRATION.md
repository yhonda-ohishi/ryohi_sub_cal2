# ETC_MEISAI モジュール統合ガイド

## 概要
`etc_meisai` モジュールは、ETC利用明細の自動取得と管理を行うGoモジュールです。
DTakoモジュールと同様の方法でSwagger統合と自動ルーティングを実装します。

## Claude向け保守指示

### 1. モジュール情報
- **パッケージ**: `github.com/yhonda-ohishi/etc_meisai`
- **現在バージョン**: v0.0.3
- **用途**: ETC利用明細の自動取得・管理・検索

### 2. 統合作業チェックリスト

#### 2.1 サービス作成
```bash
# ファイル作成
src/services/etc_meisai/etc_meisai_service.go
```

```go
package etc_meisai

import (
    "github.com/go-chi/chi/v5"
    "github.com/gorilla/mux"
    "github.com/yhonda-ohishi/etc_meisai"
    "github.com/your-org/ryohi-router/src/lib/adapters"
)

type EtcMeisaiService struct {
    enabled bool
}

func NewEtcMeisaiService(enabled bool) *EtcMeisaiService {
    return &EtcMeisaiService{
        enabled: enabled,
    }
}

func (s *EtcMeisaiService) RegisterRoutes(router *mux.Router) {
    if !s.enabled {
        return
    }

    adapters.AdaptChiToMux(router, "/etc_meisai", func(r chi.Router) {
        // etc_meisai.RegisterRoutes(r) // モジュール側にこの関数があることを確認
    })
}
```

#### 2.2 Swagger統合設定
`src/lib/swagger/merger.go` に追加:
```go
var integratedModules = []ModuleConfig{
    {
        Name:       "dtako",
        SwaggerURL: "https://raw.githubusercontent.com/yhonda-ohishi/dtako_mod/master/docs/swagger.json",
        PathPrefix: "/dtako",
    },
    // 追加
    {
        Name:       "etc_meisai",
        SwaggerURL: "https://raw.githubusercontent.com/yhonda-ohishi/etc_meisai/master/docs/swagger.json",
        PathPrefix: "/etc_meisai",
    },
}
```

#### 2.3 サーバー統合
`src/server/server.go` に追加:
```go
// インポート追加
import (
    "github.com/your-org/ryohi-router/src/services/etc_meisai"
)

// 構造体にフィールド追加
type Server struct {
    // ...既存フィールド
    etcMeisaiService *etc_meisai.EtcMeisaiService
}

// New関数内で初期化
func New(cfg *config.Config, logger *slog.Logger) (*Server, error) {
    // ...既存コード

    // ETC Meisaiサービス初期化
    s.etcMeisaiService = etc_meisai.NewEtcMeisaiService(true)

    // ...
}

// setupMainRouter内でルート登録
func (s *Server) setupMainRouter() http.Handler {
    // ...既存コード

    // ETC Meisaiルート登録（DTakoの後）
    s.etcMeisaiService.RegisterRoutes(r)

    // ...
}
```

#### 2.4 バージョン表示機能
`src/lib/etc_meisai/version.go` を作成:
```go
package etc_meisai

import (
    "runtime/debug"
    "strings"
)

func GetEtcMeisaiVersion() (string, error) {
    info, ok := debug.ReadBuildInfo()
    if !ok {
        return "", fmt.Errorf("build info not available")
    }

    for _, dep := range info.Deps {
        if strings.Contains(dep.Path, "github.com/yhonda-ohishi/etc_meisai") {
            return dep.Version, nil
        }
    }

    return "", fmt.Errorf("etc_meisai module not found")
}
```

### 3. 確認事項

#### 3.1 必須確認
- [ ] `etc_meisai`モジュールに`RegisterRoutes`関数が存在するか
- [ ] Swagger定義ファイル（`docs/swagger.json`）が存在するか
- [ ] データベース接続設定が必要か

#### 3.2 オプション確認
- [ ] 認証が必要か
- [ ] 特別な環境変数設定が必要か
- [ ] ヘルスチェックエンドポイントが必要か

### 4. テスト作成
```go
// tests/integration/etc_meisai_integration_test.go
package integration

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestEtcMeisaiEndpoints(t *testing.T) {
    // /etc_meisai/health
    // /etc_meisai/import
    // /etc_meisai/details
}
```

### 5. 環境変数設定
`.env.example` に追加:
```env
# ETC Meisai設定
ETC_MEISAI_ENABLED=true
ETC_MEISAI_DB_HOST=localhost
ETC_MEISAI_DB_PORT=3306
ETC_MEISAI_DB_NAME=etc_meisai
ETC_MEISAI_DB_USER=root
ETC_MEISAI_DB_PASSWORD=
```

### 6. ドキュメント更新
README.mdに追加:
```markdown
## 統合モジュール
- DTako Module (v1.4.0) - D'Tako関連機能
- ETC Meisai Module (v0.0.1) - ETC利用明細管理
```

## トラブルシューティング

### Swagger URLが404の場合
1. GitHubリポジトリの`docs/swagger.json`を確認
2. 別のブランチ（main/develop）を確認
3. 手動でSwagger定義を作成

### RegisterRoutes関数が存在しない場合
1. モジュールのソースコードを確認
2. 代替の登録方法を探す
3. ラッパー関数を作成

### データベース接続エラー
1. 環境変数を確認
2. データベースが起動しているか確認
3. テーブルマイグレーションを実行

## 注意事項
- etc_meisaiモジュールは「内部使用専用」とされている
- 日本語ドキュメントが主体
- Shift-JISエンコーディングのサポートが必要