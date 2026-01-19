package bot

import (
	"fmt"
	"time"

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

// Bot is the main trading bot that orchestrates scanning and position management.
type Bot struct {
	config    BotConfig
	platforms []platform.Platform
	scanner   *scanner.Scanner
	manager   *position.Manager
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
