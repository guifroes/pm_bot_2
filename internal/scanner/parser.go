package scanner

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// ParsedMarket represents the extracted information from a market title
type ParsedMarket struct {
	Asset     string  // Normalized symbol (BTC, ETH, SPY, etc.)
	Strike    float64 // Strike price
	Direction string  // "above" or "below"
}

// Asset name to symbol mapping
var assetMap = map[string]string{
	"bitcoin":  "BTC",
	"btc":      "BTC",
	"ethereum": "ETH",
	"eth":      "ETH",
	"solana":   "SOL",
	"sol":      "SOL",
	"s&p 500":  "SPY",
	"s&p500":   "SPY",
	"spy":      "SPY",
	"sp500":    "SPY",
}

// Direction keywords mapping
var aboveKeywords = []string{"above", "over", "at or above"}
var belowKeywords = []string{"below", "under", "at or below"}

// Regex patterns
var (
	// Match prices like $100,000 or $100000 or $100k or 100000 or 5000
	pricePattern = regexp.MustCompile(`\$?([\d,]+(?:\.\d+)?)(k)?`)

	// Match asset names (case insensitive)
	assetPattern = regexp.MustCompile(`(?i)\b(bitcoin|btc|ethereum|eth|solana|sol|s&p\s*500|spy|sp500)\b`)
)

// ParseMarketTitle parses a market title and extracts asset, strike, and direction
func ParseMarketTitle(title string) (*ParsedMarket, error) {
	titleLower := strings.ToLower(title)

	// Extract asset
	asset, err := extractAsset(titleLower)
	if err != nil {
		return nil, err
	}

	// Extract strike price
	strike, err := extractStrike(title)
	if err != nil {
		return nil, err
	}

	// Extract direction
	direction, err := extractDirection(titleLower)
	if err != nil {
		return nil, err
	}

	return &ParsedMarket{
		Asset:     asset,
		Strike:    strike,
		Direction: direction,
	}, nil
}

// extractAsset finds and normalizes the asset name from the title
func extractAsset(titleLower string) (string, error) {
	matches := assetPattern.FindStringSubmatch(titleLower)
	if len(matches) < 2 {
		return "", errors.New("no recognized asset found in title")
	}

	assetName := strings.ToLower(matches[1])
	// Normalize S&P variations
	assetName = strings.ReplaceAll(assetName, " ", "")
	assetName = strings.ReplaceAll(assetName, "&", "&")

	if symbol, ok := assetMap[assetName]; ok {
		return symbol, nil
	}

	// Try with spaces removed for S&P
	assetName = strings.ReplaceAll(assetName, " ", "")
	if symbol, ok := assetMap[assetName]; ok {
		return symbol, nil
	}

	return "", errors.New("no recognized asset found in title")
}

// Patterns for asset names that contain numbers (to be excluded from price extraction)
var assetWithNumberPattern = regexp.MustCompile(`(?i)s&p\s*500|sp500`)

// extractStrike finds the strike price from the title
func extractStrike(title string) (float64, error) {
	// Remove asset names that contain numbers to avoid confusion
	cleanedTitle := assetWithNumberPattern.ReplaceAllString(title, "")

	matches := pricePattern.FindAllStringSubmatch(cleanedTitle, -1)
	if len(matches) == 0 {
		return 0, errors.New("no strike price found in title")
	}

	// Use the first price found
	for _, match := range matches {
		if len(match) >= 2 {
			priceStr := strings.ReplaceAll(match[1], ",", "")
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				continue
			}

			// Handle "k" suffix (e.g., $100k = $100,000)
			if len(match) >= 3 && strings.ToLower(match[2]) == "k" {
				price *= 1000
			}

			if price > 0 {
				return price, nil
			}
		}
	}

	return 0, errors.New("no valid strike price found in title")
}

// extractDirection determines if the market is betting above or below
func extractDirection(titleLower string) (string, error) {
	// Check for "at or above" / "at or below" first (more specific)
	if strings.Contains(titleLower, "at or above") {
		return "above", nil
	}
	if strings.Contains(titleLower, "at or below") {
		return "below", nil
	}

	// Check for above keywords
	for _, keyword := range aboveKeywords {
		if strings.Contains(titleLower, keyword) {
			return "above", nil
		}
	}

	// Check for below keywords
	for _, keyword := range belowKeywords {
		if strings.Contains(titleLower, keyword) {
			return "below", nil
		}
	}

	return "", errors.New("no direction (above/below) found in title")
}
