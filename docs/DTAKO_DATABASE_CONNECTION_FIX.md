# DTako Module データベース接続修正指示書

## 緊急度: 高 🔴
DTako moduleがRyohi Routerから呼び出される際、環境変数からデータベース接続情報を読み取れていません。

## 問題の詳細
- **エラー**: `Error 1049 (42000): Unknown database 'dtako_local'`
- **原因**: DTako moduleが環境変数を読み取れていない
- **影響**: すべてのDTako APIエンドポイントが動作しない

## 確認済みの環境
```
MySQL Host: localhost
MySQL Port: 3307
Database: dtako_local
User: root
Password: kikuraku
```

データベースとテーブルは存在し、MySQLクライアントからアクセス可能です。

## 必要な修正

### 1. godotenvパッケージを追加

```bash
go get github.com/joho/godotenv
```

### 2. 環境変数の読み取り実装

**config/database.go を作成または修正:**

```go
package config

import (
    "fmt"
    "os"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    _ "github.com/joho/godotenv/autoload" // .envファイルを自動読み込み
)

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() *DatabaseConfig {
    config := &DatabaseConfig{
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        Database: os.Getenv("DB_NAME"),
    }

    // デフォルト値の設定
    if config.Host == "" {
        config.Host = "localhost"
    }
    if config.Port == "" {
        config.Port = "3306"
    }
    if config.User == "" {
        config.User = "root"
    }
    if config.Database == "" {
        config.Database = "dtako_local"
    }

    return config
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        c.User, c.Password, c.Host, c.Port, c.Database)
}

// Connect establishes database connection
func (c *DatabaseConfig) Connect() (*sql.DB, error) {
    dsn := c.GetDSN()

    // デバッグログ（環境変数DEBUGがtrueの場合のみ）
    if os.Getenv("DEBUG") == "true" {
        fmt.Printf("[DEBUG] Connecting to database at %s:%s/%s\n",
            c.Host, c.Port, c.Database)
    }

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // 接続テスト
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

### 3. リポジトリの初期化を修正

**repositories/base_repository.go を修正:**

```go
package repositories

import (
    "database/sql"
    "log"
    "sync"
    "github.com/yhonda-ohishi/dtako_mod/config"
)

var (
    db   *sql.DB
    once sync.Once
    dbErr error
)

// GetDB returns a singleton database connection
func GetDB() (*sql.DB, error) {
    once.Do(func() {
        cfg := config.GetDatabaseConfig()
        db, dbErr = cfg.Connect()
        if dbErr != nil {
            log.Printf("Failed to connect to database: %v", dbErr)
        }
    })
    return db, dbErr
}

// SetDatabaseConfig allows external configuration (optional)
func SetDatabaseConfig(host, port, user, password, database string) error {
    // 環境変数を設定
    os.Setenv("DB_HOST", host)
    os.Setenv("DB_PORT", port)
    os.Setenv("DB_USER", user)
    os.Setenv("DB_PASSWORD", password)
    os.Setenv("DB_NAME", database)

    // 再接続
    cfg := config.GetDatabaseConfig()
    newDB, err := cfg.Connect()
    if err != nil {
        return err
    }

    // 古い接続をクローズ
    if db != nil {
        db.Close()
    }

    db = newDB
    return nil
}
```

### 4. ハンドラーのエラーハンドリング改善

**handlers/dtako_rows_handler.go を修正:**

```go
func NewDtakoRowsHandler() *DtakoRowsHandler {
    return &DtakoRowsHandler{}
}

func (h *DtakoRowsHandler) List(w http.ResponseWriter, r *http.Request) {
    // データベース接続を取得
    db, err := repositories.GetDB()
    if err != nil {
        log.Printf("Database connection error: %v", err)
        http.Error(w, fmt.Sprintf("Database connection error: %v", err),
            http.StatusInternalServerError)
        return
    }

    // リポジトリを作成
    repo := repositories.NewDtakoRowsRepository(db)

    // 以下、既存の処理...
}
```

### 5. 環境変数のドキュメント化

**README.md に追加:**

```markdown
## 環境変数

DTako moduleは以下の環境変数を使用します：

| 変数名 | 説明 | デフォルト値 | 必須 |
|--------|------|-------------|------|
| DB_HOST | MySQLホスト | localhost | No |
| DB_PORT | MySQLポート | 3306 | No |
| DB_USER | MySQLユーザー | root | No |
| DB_PASSWORD | MySQLパスワード | (空) | No |
| DB_NAME | データベース名 | dtako_local | No |
| DEBUG | デバッグモード | false | No |

### 設定方法

1. `.env`ファイルを作成
```bash
cp .env.example .env
```

2. `.env`ファイルを編集
```env
DB_HOST=localhost
DB_PORT=3307
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=dtako_local
```

3. アプリケーションを起動
```bash
go run cmd/main.go
```
```

### 6. .env.exampleファイルを作成

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=dtako_local

# Debug Mode (true/false)
DEBUG=false
```

### 7. 接続診断機能の追加

**cmd/diagnose/main.go を作成（オプション）:**

```go
package main

import (
    "fmt"
    "log"
    "os"
    "github.com/yhonda-ohishi/dtako_mod/config"
    _ "github.com/joho/godotenv/autoload"
)

func main() {
    fmt.Println("=== DTako Database Connection Diagnostic ===")

    // 環境変数の確認
    fmt.Println("\n[Environment Variables]")
    fmt.Printf("DB_HOST: %s\n", os.Getenv("DB_HOST"))
    fmt.Printf("DB_PORT: %s\n", os.Getenv("DB_PORT"))
    fmt.Printf("DB_USER: %s\n", os.Getenv("DB_USER"))
    fmt.Printf("DB_PASSWORD: %s\n", maskPassword(os.Getenv("DB_PASSWORD")))
    fmt.Printf("DB_NAME: %s\n", os.Getenv("DB_NAME"))

    // 接続テスト
    fmt.Println("\n[Connection Test]")
    cfg := config.GetDatabaseConfig()
    db, err := cfg.Connect()
    if err != nil {
        log.Fatalf("❌ Connection failed: %v", err)
    }
    defer db.Close()

    fmt.Println("✅ Database connection successful!")

    // テーブル確認
    fmt.Println("\n[Tables Check]")
    tables := []string{"dtako_rows", "dtako_events", "dtako_ferry_rows"}
    for _, table := range tables {
        var exists string
        err := db.QueryRow("SHOW TABLES LIKE ?", table).Scan(&exists)
        if err != nil {
            fmt.Printf("❌ Table %s not found\n", table)
        } else {
            fmt.Printf("✅ Table %s exists\n", table)
        }
    }
}

func maskPassword(password string) string {
    if len(password) == 0 {
        return "(empty)"
    }
    return "***"
}
```

## テスト手順

1. **環境変数の確認:**
```bash
# Windows
echo %DB_HOST%
echo %DB_PORT%
echo %DB_NAME%

# Linux/Mac
echo $DB_HOST
echo $DB_PORT
echo $DB_NAME
```

2. **診断ツールの実行:**
```bash
go run cmd/diagnose/main.go
```

3. **APIエンドポイントのテスト:**
```bash
curl http://localhost:8080/dtako/rows
```

## トラブルシューティング

### 環境変数が読み込まれない場合

1. `.env`ファイルが正しい場所にあるか確認
2. godotenv/autoloadが正しくインポートされているか確認
3. 環境変数名が正しいか確認（大文字小文字を含む）

### データベースに接続できない場合

1. MySQLサービスが起動しているか確認
2. ポート番号が正しいか確認（3306 vs 3307）
3. ファイアウォール設定を確認
4. MySQLユーザーの権限を確認

## 実装チェックリスト

- [ ] godotenvパッケージをインストール
- [ ] config/database.goを作成
- [ ] 各ハンドラーでエラーハンドリングを実装
- [ ] .env.exampleファイルを作成
- [ ] READMEに環境変数の説明を追加
- [ ] 診断ツールで接続テスト実施
- [ ] すべてのAPIエンドポイントで動作確認

## Ryohi Router側の状況

Ryohi Router側では以下が完了しています：
- ✅ godotenv/autoloadがmain.goに追加済み
- ✅ .envファイルに正しい接続情報を設定済み
- ✅ DTako moduleが`/dtako/*`パスで正しくマウント済み
- ✅ MySQL localhost:3307でdtako_localデータベースが稼働中

## 完了確認

修正完了後、以下のレスポンスが返ることを確認してください：

```bash
# 成功例（データがない場合）
curl http://localhost:8080/dtako/rows
# Response: []

# 成功例（データがある場合）
curl http://localhost:8080/dtako/rows
# Response: [{"id":"1","vehicle_id":"V001",...}]

# 失敗例（現在のエラー）
curl http://localhost:8080/dtako/rows
# Response: Error 1049 (42000): Unknown database 'dtako_local'
```