package bot

import (
	"fmt"
	"time"

	"prediction-bot/internal/persistence"
	"prediction-bot/internal/platform"
	"prediction-bot/internal/position"
	"prediction-bot/internal/scanner"

	"github.com/rs/zerolog/log"
)

// BotConfig contains configuration for the trading bot.
type BotConfig struct {
	// DryRun determines if orders are simulated (true) or real (false).
	DryRun bool
	// ScanInterval is the duration between market scan cycles.
	ScanInterval time.Duration
	// MonitorInterval is the duration between position monitoring cycles.
	MonitorInterval time.Duration
}

// PriceProvider defines the interface for getting current market prices.
type PriceProvider interface {
	GetCurrentPrice(marketID string) (float64, error)
}

// Bot is the main trading bot that orchestrates scanning and position management.
type Bot struct {
	config       BotConfig
	platforms    []platform.Platform
	scanner      *scanner.Scanner
	manager      *position.Manager
	monitor      *position.Monitor
	volatility   position.VolatilityAnalyzer
	positionRepo *persistence.PositionRepository
}

// NewBot creates a new trading bot with the given configuration and dependencies.
func NewBot(
	config BotConfig,
	platforms []platform.Platform,
	scanner *scanner.Scanner,
	manager *position.Manager,
) *Bot {
	return &Bot{
		config:    config,
		platforms: platforms,
		scanner:   scanner,
		manager:   manager,
	}
}

// RunScanCycle executes a single scan cycle across all platforms.
// It scans each platform for eligible markets and processes them through
// the position manager for potential entry.
//
// Flow:
// 1. For each platform, scan for eligible markets
// 2. For each eligible market, process entry through position manager
// 3. Log results
func (b *Bot) RunScanCycle() error {
	log.Info().Msg("starting scan cycle")

	var totalEligible int
	var totalProcessed int
	var totalSkipped int

	for _, p := range b.platforms {
		platformName := p.Name()
		log.Info().
			Str("platform", platformName).
			Msg("scanning platform")

		// Scan platform for eligible markets
		eligibleMarkets, err := b.scanner.Scan(p)
		if err != nil {
			log.Error().
				Err(err).
				Str("platform", platformName).
				Msg("failed to scan platform")
			return fmt.Errorf("scan platform %s: %w", platformName, err)
		}

		log.Info().
			Str("platform", platformName).
			Int("eligible_markets", len(eligibleMarkets)).
			Msg("scan complete")

		totalEligible += len(eligibleMarkets)

		// Process each eligible market
		for _, market := range eligibleMarkets {
			log.Debug().
				Str("platform", platformName).
				Str("market_id", market.Market.ID).
				Str("title", market.Market.Title).
				Float64("probability", market.Probability).
				Str("bet_side", market.BetSide).
				Msg("processing eligible market")

			result, err := b.manager.ProcessEntry(market, b.config.DryRun)
			if err != nil {
				log.Error().
					Err(err).
					Str("platform", platformName).
					Str("market_id", market.Market.ID).
					Msg("failed to process entry")
				// Continue processing other markets
				continue
			}

			if result.Skipped {
				log.Info().
					Str("platform", platformName).
					Str("market_id", market.Market.ID).
					Str("skip_reason", result.SkipReason).
					Msg("market skipped")
				totalSkipped++
			} else {
				log.Info().
					Str("platform", platformName).
					Str("market_id", market.Market.ID).
					Int64("position_id", result.PositionID).
					Float64("position_size", result.PositionSize).
					Float64("entry_price", result.EntryPrice).
					Float64("quantity", result.Quantity).
					Float64("safety_margin", result.SafetyMargin).
					Bool("dry_run", b.config.DryRun).
					Msg("position opened")
				totalProcessed++
			}
		}
	}

	log.Info().
		Int("total_eligible", totalEligible).
		Int("total_processed", totalProcessed).
		Int("total_skipped", totalSkipped).
		Msg("scan cycle complete")

	return nil
}

// SetMonitor sets the position monitor for exit checks.
func (b *Bot) SetMonitor(monitor *position.Monitor) {
	b.monitor = monitor
}

// SetVolatilityAnalyzer sets the volatility analyzer for volatility exit checks.
func (b *Bot) SetVolatilityAnalyzer(analyzer position.VolatilityAnalyzer) {
	b.volatility = analyzer
}

// SetPositionRepo sets the position repository for fetching open positions.
func (b *Bot) SetPositionRepo(repo *persistence.PositionRepository) {
	b.positionRepo = repo
}

// RunMonitorCycle executes a single monitoring cycle for all open positions.
// It checks each position for stop loss and volatility exit conditions.
//
// Flow:
// 1. Fetch all open positions from database
// 2. For each position:
//    a. Get current market price
//    b. Check stop loss condition
//    c. Check volatility exit condition
//    d. Execute exit if any condition is triggered
func (b *Bot) RunMonitorCycle() error {
	log.Info().Msg("starting monitor cycle")

	// Fetch all open positions
	if b.positionRepo == nil {
		log.Warn().Msg("position repository not set, skipping monitor cycle")
		return nil
	}

	positions, err := b.positionRepo.GetOpen()
	if err != nil {
		return fmt.Errorf("get open positions: %w", err)
	}

	if len(positions) == 0 {
		log.Debug().Msg("no open positions to monitor")
		return nil
	}

	log.Info().
		Int("open_positions", len(positions)).
		Msg("monitoring positions")

	var totalExited int
	var stopLossExits int
	var volatilityExits int

	for _, pos := range positions {
		log.Debug().
			Int64("position_id", pos.ID).
			Str("platform", pos.Platform).
			Str("market_id", pos.MarketID).
			Float64("entry_price", pos.EntryPrice).
			Msg("checking position")

		// Find the platform for this position
		var platformClient PriceProvider
		for _, p := range b.platforms {
			if provider, ok := p.(PriceProvider); ok && p.Name() == pos.Platform {
				platformClient = provider
				break
			}
		}

		if platformClient == nil {
			log.Warn().
				Str("platform", pos.Platform).
				Int64("position_id", pos.ID).
				Msg("platform not found or does not support price lookup, skipping")
			continue
		}

		// Get current price for the market
		currentPrice, err := platformClient.GetCurrentPrice(pos.MarketID)
		if err != nil {
			log.Error().
				Err(err).
				Int64("position_id", pos.ID).
				Str("market_id", pos.MarketID).
				Msg("failed to get current price")
			continue
		}

		// Check stop loss
		if b.monitor != nil && b.monitor.CheckStopLoss(pos, currentPrice) {
			log.Info().
				Int64("position_id", pos.ID).
				Float64("entry_price", pos.EntryPrice).
				Float64("current_price", currentPrice).
				Msg("stop loss triggered")

			_, err := b.manager.ExecuteExit(pos.ID, currentPrice, position.ExitReasonStopLoss, b.config.DryRun)
			if err != nil {
				log.Error().
					Err(err).
					Int64("position_id", pos.ID).
					Msg("failed to execute stop loss exit")
				continue
			}

			stopLossExits++
			totalExited++
			continue
		}

		// Check volatility exit
		if b.monitor != nil && b.volatility != nil {
			// Calculate time to close (use 24h as default if not available)
			timeToClose := 24 * time.Hour

			shouldExit, err := b.monitor.CheckVolatilityExit(pos, b.volatility, timeToClose)
			if err != nil {
				log.Error().
					Err(err).
					Int64("position_id", pos.ID).
					Msg("failed to check volatility exit")
				continue
			}

			if shouldExit {
				log.Info().
					Int64("position_id", pos.ID).
					Float64("entry_price", pos.EntryPrice).
					Float64("current_price", currentPrice).
					Msg("volatility exit triggered")

				_, err := b.manager.ExecuteExit(pos.ID, currentPrice, position.ExitReasonVolatility, b.config.DryRun)
				if err != nil {
					log.Error().
						Err(err).
						Int64("position_id", pos.ID).
						Msg("failed to execute volatility exit")
					continue
				}

				volatilityExits++
				totalExited++
				continue
			}
		}

		log.Debug().
			Int64("position_id", pos.ID).
			Float64("current_price", currentPrice).
			Msg("position OK, no exit triggered")
	}

	log.Info().
		Int("total_monitored", len(positions)).
		Int("total_exited", totalExited).
		Int("stop_loss_exits", stopLossExits).
		Int("volatility_exits", volatilityExits).
		Msg("monitor cycle complete")

	return nil
}
