package volatility

import (
	"fmt"
	"time"

	"prediction-bot/internal/datasource"
)

// ServiceResult contains the complete volatility analysis result with context
type ServiceResult struct {
	// Asset is the analyzed asset name (e.g., "BTC", "ETH")
	Asset string
	// CurrentPrice is the current price fetched from the data source
	CurrentPrice float64
	// StrikePrice is the target strike price
	StrikePrice float64
	// Direction is the bet direction (above/below)
	Direction Direction
	// TimeToClose is the duration until market closes
	TimeToClose time.Duration
	// IsCrypto indicates if this is a cryptocurrency
	IsCrypto bool
	// Volatility is the calculated annualized volatility
	Volatility float64
	// DistanceToStrike is the relative distance from current to strike
	DistanceToStrike float64
	// ExpectedMove is the expected price movement based on volatility
	ExpectedMove float64
	// SafetyMargin is the ratio of distance to expected move
	SafetyMargin float64
	// Recommendation is the trade recommendation
	Recommendation Recommendation
	// Timestamp when the analysis was performed
	Timestamp time.Time
}

// Service combines data source and volatility analysis capabilities
type Service struct {
	aggregator *datasource.Aggregator
}

// NewService creates a new volatility service.
// alphaVantageKey can be empty if only crypto analysis is needed.
func NewService(alphaVantageKey string) *Service {
	return &Service{
		aggregator: datasource.NewAggregator(alphaVantageKey),
	}
}

// AnalyzeAsset fetches real price data and performs volatility analysis.
// It returns a complete ServiceResult with all analysis data.
//
// Parameters:
//   - asset: Asset name (e.g., "BTC", "Bitcoin", "ETH", "Ethereum")
//   - strikePrice: The strike price for the market condition
//   - direction: Whether betting above or below strike
//   - timeToClose: Duration until market closes
func (s *Service) AnalyzeAsset(asset string, strikePrice float64, direction Direction, timeToClose time.Duration) (ServiceResult, error) {
	result := ServiceResult{
		Asset:       asset,
		StrikePrice: strikePrice,
		Direction:   direction,
		TimeToClose: timeToClose,
		Timestamp:   time.Now(),
	}

	// Get current price
	price, err := s.aggregator.GetPrice(asset)
	if err != nil {
		return result, fmt.Errorf("failed to get current price for %s: %w", asset, err)
	}
	result.CurrentPrice = price.Price
	result.IsCrypto = s.aggregator.IsCrypto(asset)

	// Get historical data for volatility calculation (14 days = 336 hours)
	const historyHours = 336
	history, err := s.aggregator.GetHistory(asset, historyHours)
	if err != nil {
		return result, fmt.Errorf("failed to get history for %s: %w", asset, err)
	}

	// Calculate volatility
	result.Volatility = CalculateVolatility(history, result.IsCrypto)
	if result.Volatility <= 0 {
		return result, fmt.Errorf("could not calculate volatility for %s: insufficient data", asset)
	}

	// Perform analysis
	analysisInput := AnalysisInput{
		CurrentPrice:     result.CurrentPrice,
		StrikePrice:      strikePrice,
		Direction:        direction,
		Volatility:       result.Volatility,
		TimeToCloseHours: timeToClose.Hours(),
		IsCrypto:         result.IsCrypto,
	}

	analysisResult := Analyze(analysisInput)

	// Copy analysis results
	result.DistanceToStrike = analysisResult.DistanceToStrike
	result.ExpectedMove = analysisResult.ExpectedMove
	result.SafetyMargin = analysisResult.SafetyMargin
	result.Recommendation = analysisResult.Recommendation

	return result, nil
}
