package config

import (
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/your-org/ryohi-router/src/models"
)

// Config represents the complete router configuration
type Config struct {
	Version  string                   `yaml:"version" mapstructure:"version"`
	Router   RouterConfig             `yaml:"router" mapstructure:"router"`
	Admin    AdminConfig              `yaml:"admin" mapstructure:"admin"`
	Logging  LoggingConfig            `yaml:"logging" mapstructure:"logging"`
	Metrics  MetricsConfig            `yaml:"metrics" mapstructure:"metrics"`
	Backends []models.BackendService  `yaml:"backends" mapstructure:"backends"`
	Routes   []models.RouteConfig     `yaml:"routes" mapstructure:"routes"`
	Middleware MiddlewareConfig       `yaml:"middleware" mapstructure:"middleware"`
}

// RouterConfig represents router-specific configuration
type RouterConfig struct {
	Port            int           `yaml:"port" mapstructure:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	MaxHeaderBytes  int           `yaml:"max_header_bytes" mapstructure:"max_header_bytes"`
}

// AdminConfig represents admin API configuration
type AdminConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	APIKey  string `yaml:"api_key" mapstructure:"api_key"`
	Port    int    `yaml:"port" mapstructure:"port"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level    string `yaml:"level" mapstructure:"level"`
	Format   string `yaml:"format" mapstructure:"format"`
	Output   string `yaml:"output" mapstructure:"output"`
	FilePath string `yaml:"file_path" mapstructure:"file_path"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
	Port    int    `yaml:"port" mapstructure:"port"`
}

// MiddlewareConfig represents middleware configuration
type MiddlewareConfig struct {
	Logging     MiddlewareLoggingConfig     `yaml:"logging" mapstructure:"logging"`
	CORS        CORSConfig                  `yaml:"cors" mapstructure:"cors"`
	Compression CompressionConfig           `yaml:"compression" mapstructure:"compression"`
	Security    SecurityConfig              `yaml:"security" mapstructure:"security"`
}

// MiddlewareLoggingConfig represents logging middleware configuration
type MiddlewareLoggingConfig struct {
	Enabled     bool     `yaml:"enabled" mapstructure:"enabled"`
	SkipPaths   []string `yaml:"skip_paths" mapstructure:"skip_paths"`
	LogBody     bool     `yaml:"log_body" mapstructure:"log_body"`
	LogHeaders  bool     `yaml:"log_headers" mapstructure:"log_headers"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" mapstructure:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" mapstructure:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" mapstructure:"max_age"`
}

// CompressionConfig represents compression configuration
type CompressionConfig struct {
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	Level   int  `yaml:"level" mapstructure:"level"`
	MinSize int  `yaml:"min_size" mapstructure:"min_size"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Enabled                 bool   `yaml:"enabled" mapstructure:"enabled"`
	FrameDeny               bool   `yaml:"frame_deny" mapstructure:"frame_deny"`
	ContentTypeNosniff      bool   `yaml:"content_type_nosniff" mapstructure:"content_type_nosniff"`
	BrowserXSSFilter        bool   `yaml:"browser_xss_filter" mapstructure:"browser_xss_filter"`
	ContentSecurityPolicy   string `yaml:"content_security_policy" mapstructure:"content_security_policy"`
	HSTSMaxAge              int    `yaml:"hsts_max_age" mapstructure:"hsts_max_age"`
	HSTSIncludeSubdomains   bool   `yaml:"hsts_include_subdomains" mapstructure:"hsts_include_subdomains"`
}

// Load loads configuration from a file
func Load(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v)

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables
	overrideWithEnv(&config)

	return &config, nil
}

// LoadWithWatcher loads configuration and watches for changes
func LoadWithWatcher(configFile string, onChange func(*Config)) (*Config, error) {
	config, err := Load(configFile)
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		newConfig, err := Load(configFile)
		if err == nil {
			onChange(newConfig)
		}
	})

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate router config
	if c.Router.Port <= 0 || c.Router.Port > 65535 {
		return fmt.Errorf("invalid router port: %d", c.Router.Port)
	}

	// Validate admin config
	if c.Admin.Enabled {
		if c.Admin.APIKey == "" {
			return fmt.Errorf("admin API key is required when admin is enabled")
		}
		if c.Admin.Port <= 0 || c.Admin.Port > 65535 {
			return fmt.Errorf("invalid admin port: %d", c.Admin.Port)
		}
		if c.Admin.Port == c.Router.Port {
			return fmt.Errorf("admin port cannot be the same as router port")
		}
	}

	// Validate metrics config
	if c.Metrics.Enabled {
		if c.Metrics.Port <= 0 || c.Metrics.Port > 65535 {
			return fmt.Errorf("invalid metrics port: %d", c.Metrics.Port)
		}
		if c.Metrics.Port == c.Router.Port || c.Metrics.Port == c.Admin.Port {
			return fmt.Errorf("metrics port must be different from router and admin ports")
		}
	}

	// Validate backends
	backendIDs := make(map[string]bool)
	for i, backend := range c.Backends {
		if err := backend.Validate(); err != nil {
			return fmt.Errorf("invalid backend %d: %w", i, err)
		}
		if backendIDs[backend.ID] {
			return fmt.Errorf("duplicate backend ID: %s", backend.ID)
		}
		backendIDs[backend.ID] = true
	}

	// Validate routes
	routeIDs := make(map[string]bool)
	for i, route := range c.Routes {
		if err := route.Validate(); err != nil {
			return fmt.Errorf("invalid route %d: %w", i, err)
		}
		if routeIDs[route.ID] {
			return fmt.Errorf("duplicate route ID: %s", route.ID)
		}
		routeIDs[route.ID] = true

		// Check that backend exists
		if !backendIDs[route.Backend] {
			return fmt.Errorf("route %s references non-existent backend: %s", route.ID, route.Backend)
		}
	}

	return nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Router defaults
	v.SetDefault("router.port", 8080)
	v.SetDefault("router.read_timeout", "30s")
	v.SetDefault("router.write_timeout", "30s")
	v.SetDefault("router.idle_timeout", "120s")
	v.SetDefault("router.max_header_bytes", 1048576)

	// Admin defaults
	v.SetDefault("admin.enabled", false)
	v.SetDefault("admin.port", 8081)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.path", "/metrics")
	v.SetDefault("metrics.port", 9090)

	// Middleware defaults
	v.SetDefault("middleware.logging.enabled", true)
	v.SetDefault("middleware.cors.enabled", true)
	v.SetDefault("middleware.compression.enabled", true)
	v.SetDefault("middleware.compression.level", 5)
	v.SetDefault("middleware.security.enabled", true)
}

// overrideWithEnv overrides configuration with environment variables
func overrideWithEnv(config *Config) {
	// Router port
	if port := os.Getenv("ROUTER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Router.Port)
	}

	// Admin API key
	if apiKey := os.Getenv("ADMIN_API_KEY"); apiKey != "" {
		config.Admin.APIKey = apiKey
	}

	// Log level
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
}