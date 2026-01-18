package main

import (
	"fmt"
	"os"

	"prediction-bot/internal/config"
)

func main() {
	fmt.Println("Bot starting...")

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Bankroll - Polymarket: $%.2f, Kalshi: $%.2f\n",
		cfg.Bankroll.Polymarket, cfg.Bankroll.Kalshi)
}
