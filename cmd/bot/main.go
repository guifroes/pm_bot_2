package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prediction-bot/internal/bot"
	"prediction-bot/internal/config"
	"prediction-bot/internal/persistence"
	"prediction-bot/internal/platform"
	"prediction-bot/internal/platform/kalshi"
	"prediction-bot/internal/platform/polymarket"
	"prediction-bot/internal/position"
	"prediction-bot/internal/scanner"
	"prediction-bot/internal/sizing"
	"prediction-bot/internal/volatility"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Parse CLI flags
	configPath := flag.String("config", "config/config.yaml", "Path to config file")
	dryRun := flag.Bool("dry-run", true, "Run in dry-run mode (no real orders)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().
		Str("config", *configPath).
		Bool("dry_run", *dryRun).
		Bool("verbose", *verbose).
		Msg("Bot starting...")

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	log.Info().
		Float64("bankroll_polymarket", cfg.Bankroll.Polymarket).
		Float64("bankroll_kalshi", cfg.Bankroll.Kalshi).
		Msg("Configuration loaded")

	// Initialize database
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "bot.db"
	}
	db, err := persistence.OpenDB(dbPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", dbPath).Msg("Failed to open database")
	}
	defer db.Close()

	// Run migrations
	if err := persistence.RunMigrations(db, "migrations"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Initialize repositories
	posRepo := persistence.NewPositionRepository(db)
	bankRepo := persistence.NewBankrollRepository(db)

	// Initialize bankroll for platforms
	if err := bankRepo.Initialize("polymarket", cfg.Bankroll.Polymarket); err != nil {
		log.Warn().Err(err).Msg("Failed to initialize polymarket bankroll (may already exist)")
	}
	if err := bankRepo.Initialize("kalshi", cfg.Bankroll.Kalshi); err != nil {
		log.Warn().Err(err).Msg("Failed to initialize kalshi bankroll (may already exist)")
	}

	// Get Alpha Vantage API key from environment
	alphaVantageKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	if alphaVantageKey == "" {
		log.Warn().Msg("ALPHAVANTAGE_API_KEY not set, stock data will not be available")
	}

	// Initialize volatility service
	volService := volatility.NewService(alphaVantageKey)

	// Initialize sizer
	sizerConfig := sizing.SizerConfig{
		KellyFraction:  cfg.Parameters.KellyFraction,
		MinPosition:    1.0,
		MaxBankrollPct: 0.20,
	}
	sizer := sizing.NewSizer(sizerConfig)

	// Initialize position manager
	manager := position.NewManager(posRepo, bankRepo, volService, sizer)

	// Initialize position monitor
	monitor := position.NewMonitor(cfg.Parameters.StopLossPercent)

	// Initialize scanner
	sc := scanner.NewScanner(cfg.Parameters)

	// Initialize platforms
	var platforms []platform.Platform

	// Try to initialize Polymarket client
	polyClient, err := polymarket.NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize Polymarket client (check POLYMARKET_PRIVATE_KEY)")
	} else {
		platforms = append(platforms, polyClient)
		log.Info().Msg("Polymarket client initialized")
	}

	// Try to initialize Kalshi client
	kalshiClient, err := kalshi.NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize Kalshi client (check KALSHI_* env vars)")
	} else {
		platforms = append(platforms, kalshiClient)
		log.Info().Msg("Kalshi client initialized")
	}

	if len(platforms) == 0 {
		log.Fatal().Msg("No platforms initialized. Check your API keys.")
	}

	// Create bot config
	botConfig := bot.BotConfig{
		DryRun:          *dryRun,
		ScanInterval:    time.Duration(cfg.Scan.IntervalSeconds) * time.Second,
		MonitorInterval: 5 * time.Second,
	}

	// Create bot
	tradingBot := bot.NewBot(botConfig, platforms, sc, manager)
	tradingBot.SetMonitor(monitor)
	tradingBot.SetVolatilityAnalyzer(volService)
	tradingBot.SetPositionRepo(posRepo)

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
		cancel()
	}()

	log.Info().
		Bool("dry_run", *dryRun).
		Int("platforms", len(platforms)).
		Msg("Starting bot main loop")

	// Run bot
	if err := tradingBot.Run(ctx); err != nil {
		log.Error().Err(err).Msg("Bot stopped with error")
		os.Exit(1)
	}

	log.Info().Msg("Bot stopped gracefully")
}
