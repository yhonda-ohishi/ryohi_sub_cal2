package swagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ModuleConfig 統合するモジュールの設定
type ModuleConfig struct {
	Name       string // モジュール名（例: "dtako"）
	SwaggerURL string // SwaggerファイルのGitHub URL
	PathPrefix string // URLパスのプレフィックス（例: "/dtako_events"）
}

var integratedModules = []ModuleConfig{
	// DTakoモジュールの統合設定
	// 注意: 現在はGitHubから直接取得するように変更
	{
		Name:       "dtako",
		SwaggerURL: "https://raw.githubusercontent.com/yhonda-ohishi/dtako_mod/master/docs/swagger.json",
		PathPrefix: "/dtako",
	},
	// ETC Meisaiモジュールの統合設定
	{
		Name:       "etc_meisai",
		SwaggerURL: "https://raw.githubusercontent.com/yhonda-ohishi/etc_meisai/master/docs/swagger.yaml",
		PathPrefix: "/etc_meisai",
	},
}

// SwaggerMerger モジュールのSwaggerを統合するツール
type SwaggerMerger struct {
	docsPath    string
	logger      *slog.Logger
	moduleURLs  map[string]string
}

// NewSwaggerMerger 新しいSwaggerMergerを作成
func NewSwaggerMerger(docsPath string, logger *slog.Logger) *SwaggerMerger {
	return &SwaggerMerger{
		docsPath:   docsPath,
		logger:     logger,
		moduleURLs: make(map[string]string),
	}
}

// SetModuleURLs sets the module URLs from registry
func (m *SwaggerMerger) SetModuleURLs(urls map[string]string) {
	m.moduleURLs = urls
}

// MergeOnStartup 起動時にSwaggerを統合
func (m *SwaggerMerger) MergeOnStartup() error {
	m.logger.Info("Starting module Swagger integration...")

	// メインのSwaggerファイルを読み込み
	mainSwaggerPath := filepath.Join(m.docsPath, "swagger.json")
	mainBytes, err := ioutil.ReadFile(mainSwaggerPath)
	if err != nil {
		return fmt.Errorf("failed to read main swagger: %w", err)
	}

	var mainDoc map[string]interface{}
	if err := json.Unmarshal(mainBytes, &mainDoc); err != nil {
		return fmt.Errorf("failed to parse main swagger: %w", err)
	}

	// レジストリから登録されたモジュールのSwaggerを統合
	for moduleName, swaggerURL := range m.moduleURLs {
		m.logger.Debug("Integrating module", "name", moduleName, "url", swaggerURL)

		moduleSwagger, err := m.fetchModuleSwagger(swaggerURL)
		if err != nil {
			m.logger.Warn("Failed to fetch module swagger, skipping", "module", moduleName, "error", err)
			continue
		}

		module := ModuleConfig{
			Name:       moduleName,
			SwaggerURL: swaggerURL,
			PathPrefix: "/" + moduleName,
		}

		if err := m.mergeModuleSwagger(mainDoc, moduleSwagger, module); err != nil {
			m.logger.Warn("Failed to merge module swagger", "module", moduleName, "error", err)
			continue
		}

		m.logger.Info("Module swagger integrated successfully", "module", moduleName)
	}

	// ハードコードされたDTakoモジュール統合（後方互換性のため）
	for _, module := range integratedModules {
		m.logger.Debug("Integrating hardcoded module", "name", module.Name, "url", module.SwaggerURL)

		moduleSwagger, err := m.fetchModuleSwagger(module.SwaggerURL)
		if err != nil {
			m.logger.Warn("Failed to fetch module swagger, skipping", "module", module.Name, "error", err)
			continue
		}

		if err := m.mergeModuleSwagger(mainDoc, moduleSwagger, module); err != nil {
			m.logger.Warn("Failed to merge module swagger", "module", module.Name, "error", err)
			continue
		}

		m.logger.Info("Module swagger integrated successfully", "module", module.Name)
	}

	// 統合されたSwaggerを保存
	mergedBytes, err := json.MarshalIndent(mainDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal merged swagger: %w", err)
	}

	if err := ioutil.WriteFile(mainSwaggerPath, mergedBytes, 0644); err != nil {
		return fmt.Errorf("failed to write merged swagger: %w", err)
	}

	m.logger.Info("Module Swagger integration completed successfully")
	return nil
}

// convertOpenAPIRefs OpenAPI 3.0の参照をSwagger 2.0形式に変換
func (m *SwaggerMerger) convertOpenAPIRefs(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if key == "$ref" {
				if strVal, ok := value.(string); ok {
					// #/components/schemas/ を #/definitions/ に変換
					result[key] = strings.ReplaceAll(strVal, "#/components/schemas/", "#/definitions/")
				} else {
					result[key] = value
				}
			} else {
				result[key] = m.convertOpenAPIRefs(value)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = m.convertOpenAPIRefs(item)
		}
		return result
	default:
		return data
	}
}

// fetchModuleSwagger マイクロサービスからSwaggerを取得
func (m *SwaggerMerger) fetchModuleSwagger(swaggerURL string) (map[string]interface{}, error) {
	m.logger.Debug("Fetching module swagger", "url", swaggerURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(swaggerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch module swagger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("module returned status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var moduleDoc map[string]interface{}

	// YAMLファイルの場合はYAMLとしてパース
	if strings.HasSuffix(swaggerURL, ".yaml") || strings.HasSuffix(swaggerURL, ".yml") {
		if err := yaml.Unmarshal(body, &moduleDoc); err != nil {
			return nil, fmt.Errorf("failed to parse module swagger as YAML: %w", err)
		}
	} else {
		// JSONファイルの場合はJSONとしてパース
		if err := json.Unmarshal(body, &moduleDoc); err != nil {
			return nil, fmt.Errorf("failed to parse module swagger as JSON: %w", err)
		}
	}

	// OpenAPI 3.0の場合、componentsをdefinitionsに変換
	if components, ok := moduleDoc["components"].(map[string]interface{}); ok {
		if schemas, ok := components["schemas"].(map[string]interface{}); ok {
			moduleDoc["definitions"] = schemas
		}
	}

	// OpenAPI 3.0の参照を変換
	moduleDoc = m.convertOpenAPIRefs(moduleDoc).(map[string]interface{})

	return moduleDoc, nil
}

// mergeModuleSwagger モジュールのSwaggerをメインに統合
func (m *SwaggerMerger) mergeModuleSwagger(mainDoc, moduleDoc map[string]interface{}, module ModuleConfig) error {
	// パスを統合
	mainPaths, ok := mainDoc["paths"].(map[string]interface{})
	if !ok {
		mainPaths = make(map[string]interface{})
		mainDoc["paths"] = mainPaths
	}

	if modulePaths, ok := moduleDoc["paths"].(map[string]interface{}); ok {
		for path, pathDef := range modulePaths {
			// DTakoモジュールの場合、パスにプレフィックスを追加
			fullPath := path
			if module.Name == "dtako" && !strings.HasPrefix(path, "/dtako") {
				fullPath = "/dtako" + path
			}

			// DTako関連のパスのみを統合
			if module.Name == "dtako" && !strings.HasPrefix(fullPath, "/dtako") {
				continue // DTako以外のルートはスキップ
			}

			// パスをメインに追加（既存のタグをそのまま保持）
			mainPaths[fullPath] = pathDef
			m.logger.Info("Added path from module", "module", module.Name, "path", fullPath)
		}
	}

	// 定義を統合
	if moduleDefinitions, ok := moduleDoc["definitions"].(map[string]interface{}); ok {
		mainDefinitions, ok := mainDoc["definitions"].(map[string]interface{})
		if !ok {
			mainDefinitions = make(map[string]interface{})
			mainDoc["definitions"] = mainDefinitions
		}

		// DTakoモジュールの定義をそのまま追加（models.DtakoEvent等）
		for defName, defValue := range moduleDefinitions {
			mainDefinitions[defName] = defValue
			m.logger.Debug("Added definition from module", "module", module.Name, "definition", defName)
		}
	}

	pathsAdded := 0
	if modulePaths, ok := moduleDoc["paths"].(map[string]interface{}); ok {
		pathsAdded = len(modulePaths)
	}
	m.logger.Debug("Module swagger merge completed", "module", module.Name, "paths_added", pathsAdded)
	return nil
}