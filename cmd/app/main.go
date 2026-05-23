package main

import (
	"github.com/joho/godotenv"
	"github.com/yourusername/wemake/api"
	"github.com/yourusername/wemake/internal/config"
	"github.com/yourusername/wemake/internal/jobs"
	"github.com/yourusername/wemake/internal/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Warn("env file not loaded", "err", err)
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("failed to load config", "err", err)
	}

	// Initialize database
	db, err := config.InitDatabase(cfg)
	if err != nil {
		logger.Fatal("failed to initialize database", "err", err)
	}
	defer db.Close()

	// Start background jobs (expiration + auto-matching notifications)
	jobs.Start(db)

	// Initialize router and start server
	app := api.SetupRoutes(db, cfg)

	logger.Info("starting server", "port", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		logger.Fatal("server failed to start", "err", err, "port", cfg.Port)
	}
}
