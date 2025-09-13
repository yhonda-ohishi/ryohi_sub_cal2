# DTako Module ルートパス修正指示書

## 問題
現在、DTako moduleの`RegisterRoutes`関数が`/dtako`プレフィックスを含んでいるため、Ryohi Routerで`/dtako`にマウントすると、実際のパスが`/dtako/dtako/rows`のように二重になってしまいます。

## 現在の状況

### DTako module (routes.go)
```go
// 現在の実装（問題あり）
func RegisterRoutes(r chi.Router) {
    // ...
    r.Route("/dtako", func(r chi.Router) {  // ← ここが問題
        r.Route("/rows", func(r chi.Router) {
            r.Get("/", rowsHandler.List)
            // ...
        })
    })
}
```

### Ryohi Router側
```go
// /dtakoにマウント
AdaptChiToMux(router, "/dtako", func(r chi.Router) {
    dtako_mod.RegisterRoutes(r)
})
```

### 結果
- 期待されるパス: `/dtako/rows`
- 実際のパス: `/dtako/dtako/rows` ❌

## 必要な修正

### 方法1: DTako moduleから`/dtako`プレフィックスを削除（推奨）

**routes.goの修正:**
```go
func RegisterRoutes(r chi.Router) {
    // Initialize handlers
    rowsHandler := handlers.NewDtakoRowsHandler()
    eventsHandler := handlers.NewDtakoEventsHandler()
    ferryRowsHandler := handlers.NewDtakoFerryRowsHandler()

    // Register routes WITHOUT /dtako prefix
    // dtako_rows endpoints
    r.Route("/rows", func(r chi.Router) {
        r.Get("/", rowsHandler.List)
        r.Post("/import", rowsHandler.Import)
        r.Get("/{id}", rowsHandler.GetByID)
    })

    // dtako_events endpoints
    r.Route("/events", func(r chi.Router) {
        r.Get("/", eventsHandler.List)
        r.Post("/import", eventsHandler.Import)
        r.Get("/{id}", eventsHandler.GetByID)
    })

    // dtako_ferry_rows endpoints
    r.Route("/ferry_rows", func(r chi.Router) {
        r.Get("/", ferryRowsHandler.List)
        r.Post("/import", ferryRowsHandler.Import)
        r.Get("/{id}", ferryRowsHandler.GetByID)
    })
}
```

### 方法2: RegisterRoutesWithPrefix関数を追加

別の関数を追加して、プレフィックスの有無を選択できるようにする:

```go
// RegisterRoutes registers routes with /dtako prefix (backward compatibility)
func RegisterRoutes(r chi.Router) {
    r.Route("/dtako", func(r chi.Router) {
        RegisterRoutesWithoutPrefix(r)
    })
}

// RegisterRoutesWithoutPrefix registers routes without prefix
func RegisterRoutesWithoutPrefix(r chi.Router) {
    // Initialize handlers
    rowsHandler := handlers.NewDtakoRowsHandler()
    eventsHandler := handlers.NewDtakoEventsHandler()
    ferryRowsHandler := handlers.NewDtakoFerryRowsHandler()

    // Register routes
    r.Route("/rows", func(r chi.Router) {
        r.Get("/", rowsHandler.List)
        r.Post("/import", rowsHandler.Import)
        r.Get("/{id}", rowsHandler.GetByID)
    })

    // ... other routes
}
```

Ryohi Router側では:
```go
AdaptChiToMux(router, "/dtako", func(r chi.Router) {
    dtako_mod.RegisterRoutesWithoutPrefix(r)
})
```

## Swaggerアノテーションの更新

ルートパスを変更した場合、Swaggerアノテーションも更新が必要です:

```go
// @Router /rows [get]  // プレフィックスなし
// または
// @Router /dtako/rows [get]  // プレフィックスあり
```

## 推奨アプローチ

**方法1（プレフィックスを削除）を推奨します。**

理由:
- シンプルで分かりやすい
- DTako moduleは独立したモジュールとして、どこにでもマウント可能
- マウント先のアプリケーションがパスを決定できる
- Swagger BasePath設定でパスを管理できる

## テスト手順

1. routes.goを修正
2. `swag init`でSwaggerドキュメントを再生成
3. コミット&プッシュ
4. Ryohi Router側で最新版を取得
5. 以下のエンドポイントで動作確認:
   - `/dtako/rows`
   - `/dtako/events`
   - `/dtako/ferry_rows`

## 確認項目

- [ ] routes.goから`/dtako`プレフィックスを削除
- [ ] Swaggerアノテーションを更新
- [ ] swag initを実行
- [ ] テストを実行して動作確認
- [ ] git push