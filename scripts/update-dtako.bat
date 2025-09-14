@echo off
echo Updating DTako module...

REM DTakoモジュールの最新バージョンを取得
go get github.com/yhonda-ohishi/dtako_mod@latest

REM go.modからDTakoのバージョンを取得
for /f "tokens=2" %%i in ('findstr "github.com/yhonda-ohishi/dtako_mod" go.mod') do set DTAKO_VERSION=%%i

echo DTako module updated to %DTAKO_VERSION%

REM Swagger定義を更新
echo Updating Swagger documentation with DTako %DTAKO_VERSION%...

REM main.goのSwagger descriptionを更新
powershell -Command "(Get-Content cmd\router\main.go) -replace '// @description     高性能なリクエストルーティングシステム \(DTako Module v[0-9]+\.[0-9]+\.[0-9]+\)', '// @description     高性能なリクエストルーティングシステム (DTako Module %DTAKO_VERSION%)' | Set-Content cmd\router\main.go"

REM Swaggerドキュメントを再生成
swag init -g cmd/router/main.go -o docs

echo Swagger documentation updated with DTako %DTAKO_VERSION%

REM ビルドとテスト（オプション）
echo Building router with new DTako module...
go build -o router.exe cmd/router/main.go

echo Update complete!
echo DTako module: %DTAKO_VERSION%