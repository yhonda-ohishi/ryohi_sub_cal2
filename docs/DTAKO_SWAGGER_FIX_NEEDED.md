# DTako Module Swagger修正が必要です

## 現在の状況
DTako moduleの最新版 (v0.0.0-20250913064628-3095ded3a07c) でも、まだSwaggerのインスタンス名が "swagger" のままです。

## 必要な修正

DTako moduleのリポジトリで以下のコマンドを実行してください：

```bash
# DTako moduleのディレクトリで実行
swag init -g doc.go -o docs/ --instanceName dtako
```

このコマンドにより、docs/docs.go内の以下の部分が変更されます：

**現在（問題のあるコード）:**
```go
InfoInstanceName: "swagger",
```

**修正後（期待されるコード）:**
```go
InfoInstanceName: "dtako",
```

## 確認方法

生成後、`docs/docs.go`を開いて以下を確認：
1. `InfoInstanceName: "dtako",` になっていること
2. `swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)` の行があること

## 修正後の手順

1. 変更をコミット＆プッシュ
```bash
git add docs/
git commit -m "fix: Change swagger instance name to 'dtako' to avoid conflict"
git push
```

2. Ryohi Router側で最新版を取得
```bash
go get -u github.com/yhonda-ohishi/dtako_mod@latest
```

## 重要な注意点

`--instanceName dtako` オプションを忘れずに指定してください。このオプションなしでは、デフォルトの "swagger" が使用されてしまいます。