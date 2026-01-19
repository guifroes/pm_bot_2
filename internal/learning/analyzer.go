package learning

// SegmentStats contains statistics for a parameter segment.
type SegmentStats struct {
	ParamName  string  // Name of the parameter being analyzed
	RangeStart float64 // Start of the range (inclusive)
	RangeEnd   float64 // End of the range (exclusive)
	TradeCount int     // Total number of trades in this segment
	WinCount   int     // Number of winning trades
	LossCount  int     // Number of losing trades
	WinRate    float64 // Win rate (0.0 - 1.0)
	TotalPnL   float64 // Sum of all realized PnL
	AvgPnL     float64 // Average PnL per trade
}

// Analyzer analyzes trade outcomes to identify optimal parameter segments.
type Analyzer struct{}

// NewAnalyzer creates a new Analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeBySegment groups trade outcomes by parameter ranges and calculates
// statistics for each segment. Supported parameters:
// - "probability": segments based on entry price (0.80-0.85, 0.85-0.90, etc.)
// - "safety_margin": segments based on safety margin (0.8-1.2, 1.2-1.5, 1.5-2.0, 2.0-2.5, 2.5+)
func (a *Analyzer) AnalyzeBySegment(outcomes []TradeOutcome, paramName string) []SegmentStats {
	if len(outcomes) == 0 {
		return []SegmentStats{}
	}

	var ranges []struct {
		start float64
		end   float64
	}

	switch paramName {
	case "probability":
		// Segment by entry price (probability ranges)
		ranges = []struct {
			start float64
			end   float64
		}{
			{0.80, 0.85},
			{0.85, 0.90},
			{0.90, 0.95},
			{0.95, 1.00},
		}
	case "safety_margin":
		// Segment by safety margin
		ranges = []struct {
			start float64
			end   float64
		}{
			{0.8, 1.2},
			{1.2, 1.5},
			{1.5, 2.0},
			{2.0, 2.5},
			{2.5, 5.0},
		}
	default:
		return []SegmentStats{}
	}

	segments := make([]SegmentStats, len(ranges))
	for i, r := range ranges {
		segments[i] = SegmentStats{
			ParamName:  paramName,
			RangeStart: r.start,
			RangeEnd:   r.end,
		}
	}

	// Group outcomes into segments
	for _, outcome := range outcomes {
		var value float64
		switch paramName {
		case "probability":
			value = outcome.EntryPrice
		case "safety_margin":
			value = outcome.SafetyMargin
		}

		// Find the matching segment
		for i := range segments {
			if value >= segments[i].RangeStart && value < segments[i].RangeEnd {
				segments[i].TradeCount++
				segments[i].TotalPnL += outcome.RealizedPnL
				if outcome.IsWin() {
					segments[i].WinCount++
				} else {
					segments[i].LossCount++
				}
				break
			}
		}
	}

	// Calculate averages and win rates
	for i := range segments {
		if segments[i].TradeCount > 0 {
			segments[i].WinRate = float64(segments[i].WinCount) / float64(segments[i].TradeCount)
			segments[i].AvgPnL = segments[i].TotalPnL / float64(segments[i].TradeCount)
		}
	}

	return segments
}
