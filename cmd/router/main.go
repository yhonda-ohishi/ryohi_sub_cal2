package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/your-org/ryohi-router/src/lib/config"
	"github.com/your-org/ryohi-router/src/server"

	_ "github.com/joho/godotenv/autoload" // Auto-load .env file
	_ "github.com/your-org/ryohi-router/docs"
	_ "github.com/yhonda-ohishi/dtako_mod/docs" // DTako module with instance name "dtako"
)

// @title           Ryohi Router API
// @version         1.0.0
// @description     高性能なリクエストルーティングシステム
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	// Parse command line flags
	var (
		configFile     = flag.String("config", "configs/config.yaml", "Path to configuration file")
		validateConfig = flag.Bool("validate-config", false, "Validate configuration and exit")
		version        = flag.Bool("version", false, "Show version and exit")
	)
	flag.Parse()

	// Show version
	if *version {
		fmt.Println("Ryohi Router v1.0.0")
		os.Exit(0)
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	if *validateConfig {
		logger.Info("Configuration is valid")
		os.Exit(0)
	}

	// Update logger based on configuration
	logLevel := slog.LevelInfo
	switch cfg.Logging.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	if cfg.Logging.Format == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	}

	// Create server
	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Start server
	logger.Info("Starting Ryohi Router",
		"port", cfg.Router.Port,
		"admin_port", cfg.Admin.Port,
		"metrics_port", cfg.Metrics.Port,
	)

	if err := srv.Start(ctx); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown gracefully", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped")
}