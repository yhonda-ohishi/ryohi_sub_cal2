package dtako

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GetDTakoVersion go.modからDTakoモジュールのバージョンを取得
func GetDTakoVersion() (string, error) {
	// プロジェクトルートのgo.modファイルを読み込み
	goModPath := "go.mod"
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	dtakoRegex := regexp.MustCompile(`github\.com/yhonda-ohishi/dtako_mod\s+v?(.+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := dtakoRegex.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod: %w", err)
	}

	return "unknown", nil
}

// UpdateSwaggerDescription Swagger定義のdescriptionを動的に更新
func UpdateSwaggerDescription(docsPath string) error {
	version, err := GetDTakoVersion()
	if err != nil {
		return err
	}

	swaggerPath := filepath.Join(docsPath, "swagger.json")
	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		return fmt.Errorf("failed to read swagger.json: %w", err)
	}

	content := string(data)

	// descriptionフィールドを更新
	oldPattern := `"description": "高性能なリクエストルーティングシステム \(DTako Module v[0-9]+\.[0-9]+\.[0-9]+\)"`
	newDescription := fmt.Sprintf(`"description": "高性能なリクエストルーティングシステム (DTako Module %s)"`, version)

	// 既存のバージョン記載がない場合の処理
	if !strings.Contains(content, "DTako Module") {
		oldPattern = `"description": "高性能なリクエストルーティングシステム"`
		newDescription = fmt.Sprintf(`"description": "高性能なリクエストルーティングシステム (DTako Module %s)"`, version)
	}

	re := regexp.MustCompile(oldPattern)
	content = re.ReplaceAllString(content, newDescription)

	// ファイルに書き戻し
	err = os.WriteFile(swaggerPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write swagger.json: %w", err)
	}

	// swagger.yamlも同様に更新
	swaggerYamlPath := filepath.Join(docsPath, "swagger.yaml")
	if _, err := os.Stat(swaggerYamlPath); err == nil {
		data, err := os.ReadFile(swaggerYamlPath)
		if err == nil {
			content := string(data)
			oldPattern := `description: 高性能なリクエストルーティングシステム \(DTako Module v[0-9]+\.[0-9]+\.[0-9]+\)`
			newDescription := fmt.Sprintf(`description: 高性能なリクエストルーティングシステム (DTako Module %s)`, version)

			if !strings.Contains(content, "DTako Module") {
				oldPattern = `description: 高性能なリクエストルーティングシステム`
				newDescription = fmt.Sprintf(`description: 高性能なリクエストルーティングシステム (DTako Module %s)`, version)
			}

			re := regexp.MustCompile(oldPattern)
			content = re.ReplaceAllString(content, newDescription)
			os.WriteFile(swaggerYamlPath, []byte(content), 0644)
		}
	}

	return nil
}