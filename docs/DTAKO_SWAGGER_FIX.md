# DTako Module Swagger名前衝突の修正方法

## 問題
現在、DTako moduleとRyohi Router両方が同じ名前 "swagger" でSwaggerインスタンスを登録しているため、以下のエラーが発生します：
```
panic: Register called twice for swag: swagger
```

## 解決方法

### DTako module側で修正する内容

1. **docs/docs.goを編集**
   
   現在のコード:
   ```go
   func init() {
       swag.Register(swag.Name, &s{})
   }
   ```
   
   修正後:
   ```go
   func init() {
       swag.Register("dtako", &s{})
   }
   ```

2. **Swaggerドキュメントを再生成**
   ```bash
   swag init -g doc.go -o docs/ --instanceName dtako
   ```

   このコマンドで以下が生成されます：
   - docs/docs.go (instanceName: "dtako"で生成)
   - docs/swagger.json
   - docs/swagger.yaml

3. **変更をコミット&プッシュ**
   ```bash
   git add docs/
   git commit -m "fix: Change swagger instance name to avoid conflict with router"
   git push
   ```

## Ryohi Router側での統合方法

DTako moduleが修正された後、Ryohi Router側で以下を実行：

1. **DTako moduleを更新**
   ```bash
   go get -u github.com/yhonda-ohishi/dtako_mod@latest
   ```

2. **main.goでDTako docsをインポート**
   ```go
   import (
       _ "github.com/your-org/ryohi-router/docs"
       _ "github.com/yhonda-ohishi/dtako_mod/docs"  // 別名で登録されるため衝突しない
   )
   ```

3. **複数のSwaggerインスタンスを扱う場合のハンドラー設定**
   
   src/server/server.goを修正:
   ```go
   import (
       httpSwagger "github.com/swaggo/http-swagger"
       _ "github.com/your-org/ryohi-router/docs"
       dtakoDocs "github.com/yhonda-ohishi/dtako_mod/docs"
   )
   
   // Swagger documentation endpoints
   // Main router swagger
   r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
   
   // DTako module swagger (optional: 別エンドポイントで提供する場合)
   r.PathPrefix("/swagger-dtako/").Handler(httpSwagger.Handler(
       httpSwagger.InstanceName("dtako"),
   ))
   ```

## 代替案：単一のSwaggerドキュメントに統合

もし単一のSwagger UIで全てのAPIを表示したい場合：

1. DTako moduleではSwaggerドキュメントの生成のみ行い、init()関数でのRegisterを削除
2. Ryohi Router側でDTako moduleのSwagger定義をインポートして統合

### DTako module側
```go
// docs/docs.go のinit()関数をコメントアウトまたは削除
// func init() {
//     swag.Register(swag.Name, &s{})
// }
```

### Ryohi Router側
APIアノテーションでDTako moduleのパスも含める:
```go
// @title           Ryohi Router API with DTako
// @version         1.0.0
// @description     高性能なリクエストルーティングシステム（DTako統合版）
```

そして、DTako moduleのエンドポイント定義を手動でコピーまたは自動マージスクリプトを作成。

## 推奨される解決策

**instanceNameを使用する方法（最初の解決方法）が推奨されます。**

理由：
- 各モジュールが独立してSwaggerドキュメントを管理できる
- 名前空間の分離により衝突を防げる
- モジュールの更新が容易
- 将来的に他のモジュールを追加する際も同じパターンで対応可能

## テスト方法

1. DTako moduleで変更を実施
2. Ryohi Routerで`go get -u`を実行
3. サーバーを起動
4. ブラウザで以下を確認:
   - http://localhost:8080/swagger/ - Router API
   - http://localhost:8080/swagger-dtako/ - DTako API (オプション)