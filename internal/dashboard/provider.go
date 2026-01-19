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

// NullPriceGetter is a no-op price getter that returns the entry price.
type NullPriceGetter struct{}

// GetCurrentPrice always returns an error (forcing fallback to entry price).
func (n *NullPriceGetter) GetCurrentPrice(platform, marketID string) (float64, error) {
	return 0, nil
}
