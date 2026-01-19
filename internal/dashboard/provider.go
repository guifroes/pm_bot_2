package dashboard

import (
	"prediction-bot/internal/dashboard/views"
	"prediction-bot/internal/persistence"
)

// DBDataProvider implements DataProvider using database repositories.
type DBDataProvider struct {
	bankrollRepo *persistence.BankrollRepository
	positionRepo *persistence.PositionRepository
	priceGetter  PriceGetter
}

// PriceGetter interface for getting current market prices.
type PriceGetter interface {
	GetCurrentPrice(platform, marketID string) (float64, error)
}

// NewDBDataProvider creates a new DBDataProvider.
func NewDBDataProvider(
	bankrollRepo *persistence.BankrollRepository,
	positionRepo *persistence.PositionRepository,
	priceGetter PriceGetter,
) *DBDataProvider {
	return &DBDataProvider{
		bankrollRepo: bankrollRepo,
		positionRepo: positionRepo,
		priceGetter:  priceGetter,
	}
}

// GetBankrolls implements DataProvider.
func (p *DBDataProvider) GetBankrolls() ([]views.BankrollData, error) {
	if p.bankrollRepo == nil {
		return nil, nil
	}

	bankrolls, err := p.bankrollRepo.GetAll()
	if err != nil {
		return nil, err
	}

	var result []views.BankrollData
	for _, b := range bankrolls {
		result = append(result, views.BankrollData{
			Platform:      b.Platform,
			InitialAmount: b.InitialAmount,
			CurrentAmount: b.CurrentAmount,
		})
	}

	return result, nil
}

// GetPositions implements DataProvider.
func (p *DBDataProvider) GetPositions() ([]views.PositionData, error) {
	if p.positionRepo == nil {
		return nil, nil
	}

	positions, err := p.positionRepo.GetOpen()
	if err != nil {
		return nil, err
	}

	var result []views.PositionData
	for _, pos := range positions {
		// Get current price if price getter is available
		currentPrice := pos.EntryPrice // Default to entry price
		if p.priceGetter != nil {
			if price, err := p.priceGetter.GetCurrentPrice(pos.Platform, pos.MarketID); err == nil {
				currentPrice = price
			}
		}

		result = append(result, views.PositionData{
			ID:           pos.ID,
			Platform:     pos.Platform,
			MarketTitle:  pos.MarketTitle,
			Asset:        pos.Asset,
			EntryPrice:   pos.EntryPrice,
			CurrentPrice: currentPrice,
			Quantity:     pos.Quantity,
			Side:         pos.Side,
			EntryTime:    pos.EntryTime,
		})
	}

	return result, nil
}

// GetStats implements DataProvider.
func (p *DBDataProvider) GetStats() (views.StatsData, error) {
	if p.positionRepo == nil {
		return views.StatsData{}, nil
	}

	// Get all closed positions for stats
	positions, err := p.positionRepo.GetClosed()
	if err != nil {
		return views.StatsData{}, err
	}

	// Get open positions for unrealized PnL
	openPositions, err := p.positionRepo.GetOpen()
	if err != nil {
		return views.StatsData{}, err
	}

	var stats views.StatsData
	stats.TotalTrades = len(positions)

	var totalRealizedPnL float64
	var maxBalance, minBalance, currentBalance float64

	for _, pos := range positions {
		pnl := 0.0
		if pos.RealizedPnL != nil {
			pnl = *pos.RealizedPnL
		}

		if pnl > 0 {
			stats.WinningTrades++
		} else if pnl < 0 {
			stats.LosingTrades++
		}
		totalRealizedPnL += pnl

		// Track balance for drawdown calculation
		currentBalance += pnl
		if currentBalance > maxBalance {
			maxBalance = currentBalance
		}
		if currentBalance < minBalance {
			minBalance = currentBalance
		}
	}

	stats.RealizedPnL = totalRealizedPnL

	// Calculate unrealized PnL from open positions
	var unrealizedPnL float64
	for _, pos := range openPositions {
		currentPrice := pos.EntryPrice // Default
		if p.priceGetter != nil {
			if price, err := p.priceGetter.GetCurrentPrice(pos.Platform, pos.MarketID); err == nil && price > 0 {
				currentPrice = price
			}
		}
		unrealizedPnL += (currentPrice - pos.EntryPrice) * pos.Quantity
	}
	stats.UnrealizedPnL = unrealizedPnL

	stats.TotalPnL = stats.RealizedPnL + stats.UnrealizedPnL

	// Calculate max drawdown
	if maxBalance > 0 {
		stats.MaxDrawdown = (maxBalance - minBalance) / maxBalance
	}

	return stats, nil
}

// NullPriceGetter is a no-op price getter that returns the entry price.
type NullPriceGetter struct{}

// GetCurrentPrice always returns an error (forcing fallback to entry price).
func (n *NullPriceGetter) GetCurrentPrice(platform, marketID string) (float64, error) {
	return 0, nil
}
