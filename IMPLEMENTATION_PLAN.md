# IMPLEMENTATION PLAN - Prediction Market Bot
## Micro-Fatias Verticais (v2 - Revisado)

> **Regra de Atomicidade:** Cada fatia toca no máximo 3 arquivos e resolve UM caminho de execução.
> **Estratégia de Teste:** Testes comportamentais realistas, sem mocks quando possível, acessando endpoints reais (exceto operações com dinheiro real).

---

# FASE 1: FUNDAÇÕES (Setup + Conectividade Básica)

## Fatia 1.1: Projeto Go - Estrutura Base ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `go.mod`, `cmd/bot/main.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `go run cmd/bot/main.go` imprime "Bot starting..." e encerra sem erros.
- **Comando:** `go run cmd/bot/main.go`

**2. The Implementation Steps**
- Passo A: Criar `go.mod` com module `prediction-bot`
- Passo B: Criar `cmd/bot/main.go` com func main que imprime e sai
- Passo C: Criar diretórios vazios: `internal/`, `pkg/`, `config/`, `migrations/`

---

## Fatia 1.2: Configuração - Carregar YAML ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/config/config.go`, `config/config.yaml`, `cmd/bot/main.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** O bot carrega `config.yaml` e imprime os valores de bankroll.
- **Comando:** `go run cmd/bot/main.go`

**2. The Implementation Steps**
- Passo A: Criar struct `Config` com campos bankroll, scan, parameters, database
- Passo B: Implementar `LoadConfig(path string) (*Config, error)` usando yaml.v3
- Passo C: Criar `config/config.yaml` com valores default do CLAUDE.md
- Passo D: Atualizar main.go para carregar e imprimir config

---

## Fatia 1.3: SQLite - Conexão e Migration Runner ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/persistence/db.go`, `migrations/001_initial.sql`

**1. The Outer Gate (Behavior Test)**
- **Teste:** O bot cria `bot.db`, tabela `schema_version` existe.
- **Comando:** `go test ./internal/persistence/... -v`

**2. The Implementation Steps**
- Passo A: Criar `OpenDB(path string) (*sql.DB, error)` com WAL mode
- Passo B: Criar `RunMigrations(db, migrationsDir)` que executa .sql em ordem
- Passo C: Criar `001_initial.sql` com `schema_version` e `bankroll`

---

## Fatia 1.4: SQLite - Schema Completo ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `migrations/002_core_tables.sql`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Todas as tabelas (positions, parameters, events, price_cache, api_log) existem.
- **Comando:** `go test ./internal/persistence/... -v`

**2. The Implementation Steps**
- Passo A: Criar migration com tabelas: positions, parameters, events
- Passo B: Criar tabelas: price_cache, price_history, api_log
- Passo C: Inserir parâmetros default (prob=0.80, margin=1.5, stop=0.15, kelly=0.25)

---

# FASE 2: BUSCAR PREÇO DO BITCOIN (End-to-End Vertical)

## Fatia 2.1: Binance - Fetch Preço Atual ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/datasource/binance/client.go`, `pkg/types/price.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetPrice("BTCUSDT")` retorna preço > 0 da API real da Binance.
- **Comando:** `go test ./internal/datasource/binance/... -v`

**2. The Implementation Steps**
- Passo A: Criar struct `Price` em `pkg/types/price.go`
- Passo B: Criar `BinanceClient.GetPrice(symbol) (Price, error)` usando REST

---

## Fatia 2.2: Binance - Fetch Histórico ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/datasource/binance/client.go` (update)

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetHistory("BTCUSDT", 336)` retorna 336 pontos de preço horário.
- **Comando:** `go test ./internal/datasource/binance/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `GetHistory(symbol, hours int) ([]Price, error)` usando klines

---

## Fatia 2.3: Alpha Vantage - Fetch Preço Atual ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/datasource/alphavantage/client.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetPrice("SPY")` retorna preço do S&P 500 ETF (requer API key).
- **Comando:** `go test ./internal/datasource/alphavantage/... -v`

**2. The Implementation Steps**
- Passo A: Criar `AlphaVantageClient` com API key do env
- Passo B: Implementar `GetPrice(symbol) (Price, error)` usando GLOBAL_QUOTE

---

## Fatia 2.4: Data Source - Aggregator + Symbol Mapper ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/datasource/aggregator.go`, `internal/datasource/mapper.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetPrice("Bitcoin")` roteia para Binance e retorna preço BTC.
- **Comando:** `go test ./internal/datasource/... -v`

**2. The Implementation Steps**
- Passo A: Criar `SymbolMapper` com mapeamentos (Bitcoin→BTCUSDT, S&P 500→SPY)
- Passo B: Criar `DataSourceAggregator` que roteia para fonte correta

---

# FASE 3: LISTAR MERCADOS DA POLYMARKET (End-to-End Vertical)

## Fatia 3.1: Polymarket - Client Base + Auth ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/platform/polymarket/client.go`, `internal/platform/polymarket/auth.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Request autenticado para Polymarket não retorna erro de auth.
- **Comando:** `go test ./internal/platform/polymarket/... -v`

**2. The Implementation Steps**
- Passo A: Criar `PolymarketClient` com private key do env
- Passo B: Implementar assinatura de requests (wallet signature EIP-712)

---

## Fatia 3.2: Polymarket - Listar Mercados Ativos ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/polymarket/markets.go`, `pkg/types/market.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `ListMarkets(filter{IsActive: true})` retorna lista não-vazia de mercados reais.
- **Comando:** `go test ./internal/platform/polymarket/... -v`

**2. The Implementation Steps**
- Passo A: Criar structs `Market`, `MarketFilter` em pkg/types
- Passo B: Implementar `ListMarkets(filter) ([]Market, error)`

---

## Fatia 3.3: Polymarket - Get Market + OrderBook ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/polymarket/orderbook.go`, `pkg/types/orderbook.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetOrderBook(marketID)` para mercado ativo retorna bids/asks.
- **Comando:** `go test ./internal/platform/polymarket/... -v`

**2. The Implementation Steps**
- Passo A: Criar structs `OrderBook`, `Level` em pkg/types
- Passo B: Implementar `GetOrderBook(marketID) (OrderBook, error)`

---

## Fatia 3.4: Polymarket - Get Balance ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/polymarket/account.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetBalance()` retorna saldo USDC da wallet (pode ser 0).
- **Comando:** `go test ./internal/platform/polymarket/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `GetBalance() (Balance, error)` consultando Polygon

---

# FASE 4: LISTAR MERCADOS DA KALSHI (End-to-End Vertical)

## Fatia 4.1: Kalshi - Client Base + Auth HMAC ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/platform/kalshi/client.go`, `internal/platform/kalshi/auth.go`

**Notas de Implementação:**
- Kalshi usa RSA-PSS (não HMAC) para assinatura com SHA256
- Formato da mensagem: `timestamp + method + path` (path inclui `/trade-api/v2`)
- Suporte para `KALSHI_PRIVATE_KEY` (conteúdo) e `KALSHI_PRIVATE_KEY_PATH` (arquivo)

**1. The Outer Gate (Behavior Test)**
- **Teste:** Request autenticado para Kalshi não retorna erro de auth.
- **Comando:** `go test ./internal/platform/kalshi/... -v`

**2. The Implementation Steps**
- Passo A: Criar `KalshiClient` com API key/secret do env
- Passo B: Implementar HMAC signature para requests

---

## Fatia 4.2: Kalshi - Listar Mercados Ativos ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/kalshi/markets.go`

**Notas de Implementação:**
- API Kalshi aceita `status=open` como filtro mas retorna `status: "active"` na resposta
- Preços são retornados em centavos (0-100), convertidos para decimal (0.0-1.0)
- Volume e liquidez são em centavos, convertidos para dólares

**1. The Outer Gate (Behavior Test)**
- **Teste:** `ListMarkets(filter{IsActive: true})` retorna mercados ativos.
- **Comando:** `go test ./internal/platform/kalshi/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `ListMarkets(filter) ([]Market, error)`
- Passo B: Mapear response Kalshi para struct Market comum

---

## Fatia 4.3: Kalshi - Get Balance + Positions ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/kalshi/account.go`

**Notas de Implementação:**
- GetBalance já existia no client.go, movido conceitualmente para account.go
- GetPositions implementado retornando []types.Position
- Valores monetários da Kalshi (centavos) convertidos para dólares
- Tipo Position criado em pkg/types/position.go

**1. The Outer Gate (Behavior Test)**
- **Teste:** `GetBalance()` e `GetPositions()` retornam sem erro.
- **Comando:** `go test ./internal/platform/kalshi/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `GetBalance() (Balance, error)`
- Passo B: Implementar `GetPositions() ([]Position, error)`

---

## Fatia 4.4: Platform - Interface Comum + Rate Limiter ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/platform/platform.go`, `internal/platform/ratelimit.go`

**Notas de Implementação:**
- Interface Platform define métodos comuns: Name, ListMarkets, GetOrderBook, GetBalance, GetPositions
- Token bucket rate limiter com Allow() (non-blocking) e Wait() (blocking)
- Factory functions NewPolymarketRateLimiter() e NewKalshiRateLimiter() para rates predefinidos
- Testes comportamentais incluem: burst, blocking, refill, concurrency

**1. The Outer Gate (Behavior Test)**
- **Teste:** Rate limiter bloqueia chamada excedente (100/min Poly, 30/min Kalshi).
- **Comando:** `go test ./internal/platform/... -v`

**2. The Implementation Steps**
- Passo A: Definir interface `Platform` em platform.go
- Passo B: Implementar token bucket rate limiter

---

# FASE 5: ANALISAR VOLATILIDADE DE UM ATIVO

## Fatia 5.1: Volatility - Cálculo de Volatilidade Histórica ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/volatility/calculator.go`

**Notas de Implementação:**
- Cálculo usa log returns para precisão estatística
- Variância usa (n-1) para sample standard deviation
- Crypto anualiza com sqrt(365), stocks com sqrt(252)
- Retorna 0 para dados insuficientes (<2 preços)

**1. The Outer Gate (Behavior Test)**
- **Teste:** Dado array de preços, retorna volatilidade anualizada entre 0 e 2.
- **Comando:** `go test ./internal/volatility/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `CalculateVolatility(prices []Price, isCrypto bool) float64`
- Passo B: Calcular std dev de daily returns, anualizar (365 crypto, 252 stocks)

---

## Fatia 5.2: Volatility - Safety Margin Calculation ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/volatility/analyzer.go`

**Notas de Implementação:**
- Fórmula: safety_margin = distance_to_strike / (2 * expected_move)
- distance_to_strike considera direção (above/below)
- Thresholds: Valid >= 1.5, Risky >= 0.8, Reject < 0.8
- Correção: BTC @ $100k, strike $95k, vol 0.5, 24h → margin ~0.95 (risky)
- Para margin > 1.0, precisa de mais distância (ex: strike $90k → margin ~1.91)

**1. The Outer Gate (Behavior Test)**
- **Teste:** BTC @ $100k, strike $90k, vol 0.5, 24h → safety_margin > 1.0.
- **Comando:** `go test ./internal/volatility/... -v`

**2. The Implementation Steps**
- Passo A: Criar structs `AnalysisInput`, `AnalysisResult`
- Passo B: Implementar `Analyze(input) Result` com distance_to_strike, expected_move, safety_margin
- Passo C: Determinar recommendation (valid/risky/reject)

---

## Fatia 5.3: Volatility - Service Integrado com DataSource ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/volatility/service.go`

**Notas de Implementação:**
- Service combina DataSource Aggregator + Calculator + Analyzer
- Busca preço atual e histórico (336h = 14 dias) via aggregator
- Calcula volatilidade e retorna ServiceResult completo
- Testes com BTC e ETH reais confirmam funcionamento

**1. The Outer Gate (Behavior Test)**
- **Teste:** `AnalyzeAsset("BTC", $100000, "above", 24h)` busca dados reais e retorna análise.
- **Comando:** `go test ./internal/volatility/... -v`

**2. The Implementation Steps**
- Passo A: Criar `VolatilityService` que combina DataSource + Analyzer
- Passo B: Buscar histórico via aggregator, calcular, analisar

---

# FASE 6: ESCANEAR MERCADOS ELEGÍVEIS

## Fatia 6.1: Scanner - Parser de Títulos ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/scanner/parser.go`

**Notas de Implementação:**
- ParsedMarket struct com Asset, Strike e Direction
- Suporte para assets: Bitcoin/BTC, Ethereum/ETH, Solana/SOL, S&P 500/SPY
- Suporte para formatos de preço: $100,000, $100k, 100000
- Direções: above/over/at or above → "above", below/under/at or below → "below"
- Regex especial para remover "500" do "S&P 500" antes de extrair strike

**1. The Outer Gate (Behavior Test)**
- **Teste:** Parse "Will Bitcoin be above $100,000 on Jan 18?" → `{Asset: "BTC", Strike: 100000, Direction: "above"}`
- **Comando:** `go test ./internal/scanner/... -v`

**2. The Implementation Steps**
- Passo A: Regex para extrair asset (Bitcoin→BTC, S&P 500→SPY)
- Passo B: Regex para extrair strike price ($100,000, $100k, 100000)
- Passo C: Regex para extrair direction (above/below/over/under)

---

## Fatia 6.2: Scanner - Filtro de Elegibilidade ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/scanner/filter.go`

**Notas de Implementação:**
- EligibilityFilter struct com parâmetros de config injetados
- EligibilityResult retorna Eligible, Reasons, Probability, BetSide
- Constantes: MaxTimeToResolution=48h, MinLiquidity=$100
- Suporte para YES e NO outcomes (escolhe maior probabilidade)
- Edge cases: market closed, not active, already ended

**1. The Outer Gate (Behavior Test)**
- **Teste:** Mercado prob=85%, closes=24h, liquidity=$500 → eligible=true.
- **Comando:** `go test ./internal/scanner/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `IsEligible(market, params) bool`
- Passo B: Checks: probability >= threshold, time < 48h, liquidity >= $100

---

## Fatia 6.3: Scanner - Scan Single Platform ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/scanner/scanner.go`

**Notas de Implementação:**
- Scanner struct com EligibilityFilter injetado
- Scan(platform) lista markets, filtra por elegibilidade, parseia títulos
- EligibleMarket contém Market, ParsedMarket, Probability e BetSide
- Mercados elegíveis mas não parseáveis (políticos, esportes) são ignorados silenciosamente
- Testes com MockPlatform validam comportamento completo

**1. The Outer Gate (Behavior Test)**
- **Teste:** Scan Polymarket retorna mercados elegíveis (pode ser 0 ou mais).
- **Comando:** `go test ./internal/scanner/... -v -timeout 60s`

**2. The Implementation Steps**
- Passo A: Criar `Scanner.Scan(platform Platform) ([]EligibleMarket, error)`
- Passo B: ListMarkets → Filter → Parse titles → Return eligible

---

# FASE 7: CALCULAR TAMANHO DA POSIÇÃO

## Fatia 7.1: Sizing - Kelly Criterion ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/sizing/kelly.go`

**Notas de Implementação:**
- Fórmula Kelly: f = (p*b - q) / b onde b = (1-price)/price
- Suporta fractional Kelly (quarter Kelly = 0.25 para reduzir risco)
- Validação completa de inputs (edge cases, negative edge retorna 0)
- Testes cobrem: standard case, negative edge, zero inputs, formula verification

**1. The Outer Gate (Behavior Test)**
- **Teste:** Entry 0.90, prob 0.92, bankroll $50, fraction 0.25 → position ~$2.50.
- **Comando:** `go test ./internal/sizing/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `CalculateKelly(entryPrice, winProb, bankroll, fraction) float64`
- Passo B: Fórmula: f = (p*b - q) / b, b = (1-price)/price

---

## Fatia 7.2: Sizing - Constraints e Win Probability ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/sizing/sizer.go`

**Notas de Implementação:**
- Sizer struct com SizerConfig injetado (KellyFraction, MinPosition, MaxBankrollPct)
- SizingInput: EntryPrice, WinProb, Bankroll, SafetyMargin
- SizingOutput: PositionSize, RawKelly, BankrollPct, Reason
- Constraints aplicados: min $1, max 20% bankroll, round down to cents
- EstimateWinProbability usa safety margin para boost/penalty
- Posição abaixo do mínimo após constraints retorna 0 (below_minimum)

**1. The Outer Gate (Behavior Test)**
- **Teste:** Position nunca excede 20% do bankroll, mínimo $1.
- **Comando:** `go test ./internal/sizing/... -v`

**2. The Implementation Steps**
- Passo A: Criar `Sizer.Calculate(input) SizingOutput`
- Passo B: Aplicar constraints: min $1, max 20%, round down
- Passo C: Implementar `EstimateWinProbability(marketPrice, safetyMargin)`

---

# FASE 8: ABRIR POSIÇÃO (DRY-RUN)

## Fatia 8.1: Position Manager - Repositories ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/persistence/positions.go`, `internal/persistence/bankroll.go`

**Notas de Implementação:**
- PositionRepository: Create, GetByID, GetOpen, GetOpenByPlatform, GetByMarket, Update, Close
- BankrollRepository: Get, GetAll, Update, Initialize, AddToBalance
- Testes cobrem todos os métodos CRUD e cenários de uso
- Struct Position estendida com campos para tracking (safety_margin, volatility, exit details)

**1. The Outer Gate (Behavior Test)**
- **Teste:** CRUD de positions e bankroll no SQLite.
- **Comando:** `go test ./internal/persistence/... -v`

**2. The Implementation Steps**
- Passo A: Criar `PositionRepository` com Create, GetOpen, GetByMarket, Update, Close
- Passo B: Criar `BankrollRepository` com Get, Update, Initialize

---

## Fatia 8.2: Position Manager - Entry Flow (DRY-RUN) ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/position/manager.go`

**Notas de Implementação:**
- Manager struct com dependencies injetadas (PositionRepo, BankrollRepo, VolatilityService, Sizer)
- ProcessEntry flow completo: duplicate check → volatility analysis → sizing → persist → bankroll deduct
- EntryResult retorna detalhes da operação (PositionID, PositionSize, Quantity, SkipReason)
- Skip reasons definidos: duplicate, volatility_reject, volatility_risky, sizing_no_edge, sizing_below_minimum
- Suporte para allowRisky flag (permite trades com safety margin entre 0.8 e 1.5)
- Testes cobrem: dry-run entry, duplicate detection, volatility reject, sizing constraints, bankroll deduction

**1. The Outer Gate (Behavior Test)**
- **Teste:** Processar EligibleMarket → volatility OK → sizing → persist Position (sem ordem real).
- **Comando:** `go test ./internal/position/... -v`

**2. The Implementation Steps**
- Passo A: Criar `PositionManager` com dependencies injetadas
- Passo B: Implementar `ProcessEntry(market, dryRun bool) error`
- Passo C: Check duplicate → Volatility → Sizing → Persist → Log

---

## Fatia 8.3: Polymarket - Place Order (DRY-RUN) ✅ CONCLUÍDA
**Risco de Contexto:** Médio ⚠️
**Arquivos Esperados:** `internal/platform/polymarket/orders.go`, `pkg/types/order.go`

**Notas de Implementação:**
- Structs Order e OrderResult em pkg/types/order.go com tipos auxiliares (OrderSide, OrderType, TimeInForce, OrderStatus)
- PlaceOrder valida campos obrigatórios (MarketID, TokenID, Size > 0, Price 0-1)
- DRY-RUN retorna OrderResult simulado com UUID único e IsDryRun=true
- Live trading placeholder para Fatia 13.1

**1. The Outer Gate (Behavior Test)**
- **Teste:** `PlaceOrder()` em DRY_RUN retorna OrderResult simulado sem executar.
- **Comando:** `go test ./internal/platform/polymarket/... -v`

**2. The Implementation Steps**
- Passo A: Criar structs `Order`, `OrderResult` em pkg/types
- Passo B: Implementar `PlaceOrder(order, dryRun bool) (OrderResult, error)`
- Passo C: Em dryRun=true, retornar resultado simulado

---

# FASE 9: MONITORAR E SAIR DA POSIÇÃO

## Fatia 9.1: Position Monitor - Stop Loss ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/position/monitor.go`

**Notas de Implementação:**
- Monitor struct com stopLossPercent injetado
- CheckStopLoss retorna true se currentPrice < threshold
- Threshold = entry_price * (1 - stop_loss_percent)
- Usa comparação estrita (< não <=) para evitar trigger no exato threshold

**1. The Outer Gate (Behavior Test)**
- **Teste:** Position entry=0.90, current=0.76 (>15% loss) → trigger exit.
- **Comando:** `go test ./internal/position/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `CheckStopLoss(position, currentPrice) bool`
- Passo B: Threshold: entry_price * (1 - stop_loss_percent)

---

## Fatia 9.2: Position Monitor - Volatility Exit ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/position/monitor.go` (update)

**Notas de Implementação:**
- Threshold de saída: safety_margin < 0.8 (VolatilityExitThreshold)
- Usa VolatilityAnalyzer interface existente (AnalyzeAsset)
- Converte direction string para volatility.Direction
- Comparação estrita (< não <=) para evitar trigger no exato threshold
- Testes cobrem: low margin, good margin, valid margin, exact threshold, just below, negative, error, direction below

**1. The Outer Gate (Behavior Test)**
- **Teste:** Safety margin atual < 0.8 → trigger volatility exit.
- **Comando:** `go test ./internal/position/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `CheckVolatilityExit(position) (bool, error)`
- Passo B: Re-calcular safety margin com dados atuais

---

## Fatia 9.3: Position Manager - Execute Exit (DRY-RUN) ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/position/manager.go` (update)

**Notas de Implementação:**
- Exit reasons definidos: stop_loss, volatility_exit, market_resolved, manual_exit
- ExitResult struct com PositionID, ExitPrice, ExitReason, RealizedPnL, EntryPrice, Quantity
- PnL calculado: (exitPrice - entryPrice) * quantity
- Exit proceeds adicionados ao bankroll: exitPrice * quantity
- Testes cobrem: stop loss exit, volatility exit, winning trade, position not found, already closed

**1. The Outer Gate (Behavior Test)**
- **Teste:** Exit em dryRun calcula PnL e atualiza position no DB.
- **Comando:** `go test ./internal/position/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `ExecuteExit(positionID, reason, dryRun) error`
- Passo B: Calcular realized PnL
- Passo C: Atualizar position status='closed', bankroll

---

# FASE 10: BOT LOOP PRINCIPAL

## Fatia 10.1: Bot - Scan Cycle ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/bot/bot.go`

**Notas de Implementação:**
- Bot struct com BotConfig, platforms, scanner e manager injetados
- RunScanCycle itera sobre todas platforms, escaneia mercados elegíveis e processa entradas
- Logging estruturado com zerolog para todas operações
- Testes cobrem: single platform, multiple platforms, no markets, ineligible markets

**1. The Outer Gate (Behavior Test)**
- **Teste:** Um ciclo de scan executa sem erro (APIs reais).
- **Comando:** `go test ./internal/bot/... -v -timeout 60s`

**2. The Implementation Steps**
- Passo A: Criar `Bot` struct com todas dependencies
- Passo B: Implementar `RunScanCycle() error`
- Passo C: Scan both platforms → Process eligible markets

---

## Fatia 10.2: Bot - Monitor Cycle ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/bot/bot.go` (update)

**Notas de Implementação:**
- RunMonitorCycle itera sobre todas posições abertas via PositionRepository
- Para cada posição: get current price, check stop loss, check volatility exit
- PriceProvider interface para obter preço atual do mercado via plataforma
- Setters adicionados: SetMonitor, SetVolatilityAnalyzer, SetPositionRepo
- Logging estruturado para todas operações e exits
- Testes cobrem: position OK, stop loss trigger, volatility exit trigger, no positions

**1. The Outer Gate (Behavior Test)**
- **Teste:** Monitor cycle checa todas positions abertas.
- **Comando:** `go test ./internal/bot/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `RunMonitorCycle() error`
- Passo B: Para cada position aberta: check stop loss, volatility exit, resolution

---

## Fatia 10.3: Bot - Main Loop Contínuo ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/bot/bot.go`, `cmd/bot/main.go`

**Notas de Implementação:**
- Run(ctx context.Context) implementado com scan e monitor cycles
- Immediate scan e monitor na inicialização
- Tickers configuráveis para ScanInterval e MonitorInterval
- Graceful shutdown via context cancellation
- CLI com flags: --config, --dry-run, --verbose
- Signal handling para SIGINT/SIGTERM
- Logging estruturado com zerolog console output
- Platform interface implementada corretamente em Polymarket e Kalshi clients

**1. The Outer Gate (Behavior Test)**
- **Teste:** `go run cmd/bot/main.go --dry-run` roda por 30s e para gracefully.
- **Comando:** `go run cmd/bot/main.go --dry-run` (Ctrl+C after 30s)

**2. The Implementation Steps**
- Passo A: Implementar `Run(ctx context.Context) error` com ticker 10s
- Passo B: Graceful shutdown via context
- Passo C: CLI com flags: --config, --dry-run, --verbose

---

# FASE 11: LEARNING SYSTEM

## Fatia 11.1: Learning - Coletar Trade Outcomes ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/learning/collector.go`

**Notas de Implementação:**
- TradeOutcome struct com todos os campos relevantes: position ID, platform, asset, strike, direction, prices, PnL, times, safety margin, volatility
- Métodos auxiliares: IsWin() para determinar trades vencedores, ReturnPercent() para calcular retorno percentual
- CollectOutcomes retorna slice vazia se há menos de minTrades fechados
- Resultados ordenados por exit_time DESC (mais recentes primeiro)
- Limite de minTrades para retornar apenas N trades mais recentes
- Parser de timestamps SQLite (formato "2006-01-02 15:04:05")

**1. The Outer Gate (Behavior Test)**
- **Teste:** Coleta últimos 20 trades fechados com parâmetros usados.
- **Comando:** `go test ./internal/learning/... -v`

**2. The Implementation Steps**
- Passo A: Criar struct `TradeOutcome`
- Passo B: Implementar `CollectOutcomes(minTrades int) ([]TradeOutcome, error)`

---

## Fatia 11.2: Learning - Análise por Segmento ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/learning/analyzer.go`

**Notas de Implementação:**
- SegmentStats struct com ParamName, RangeStart, RangeEnd, TradeCount, WinCount, LossCount, WinRate, TotalPnL, AvgPnL
- Suporte para "probability" (segmentos 80-85, 85-90, 90-95, 95-100)
- Suporte para "safety_margin" (segmentos 0.8-1.2, 1.2-1.5, 1.5-2.0, 2.0-2.5, 2.5+)
- Retorna slice vazia para parâmetros desconhecidos ou outcomes vazios
- Testes cobrem: probability segments, safety margin segments, empty outcomes, invalid param, single trade, total PnL

**1. The Outer Gate (Behavior Test)**
- **Teste:** Calcular win rate por segmento de probability.
- **Comando:** `go test ./internal/learning/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `AnalyzeBySegment(outcomes, paramName) []SegmentStats`
- Passo B: Agrupar por ranges, calcular win rate e avg PnL

---

## Fatia 11.3: Learning - Ajustar Parâmetros ✅ CONCLUÍDA
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/learning/adjuster.go`, `internal/persistence/parameters.go`

**Notas de Implementação:**
- Adjuster com SuggestAdjustment que move gradualmente (max 10%) para melhor segmento
- Usa scoring baseado em win rate * (1 + normalized avg pnl) para escolher melhor segmento
- Min 5 trades por segmento para ser considerado
- ParametersRepository com GetCurrent, GetByName, Save, SaveWithReason, GetHistory, GetLastAdjustmentTime
- Migration 003_parameter_history.sql para histórico de ajustes
- Guardrails struct com CheckCanAdjust (min 20 trades, 24h cooldown) e CheckDrawdown (20% threshold)
- DefaultParameters() para revert em caso de drawdown

**1. The Outer Gate (Behavior Test)**
- **Teste:** Ajuste gradual (max 10%) em direção ao melhor segmento.
- **Comando:** `go test ./internal/learning/... -v`

**2. The Implementation Steps**
- Passo A: Implementar `SuggestAdjustment(current, segments, bounds) float64`
- Passo B: Criar `ParametersRepository` com GetCurrent, Save
- Passo C: Guardrails: min 20 trades, cooldown, revert on 20% drawdown

---

# FASE 12: DASHBOARD (Terminal UI)

## Fatia 12.1: Dashboard - Layout Base ✅ CONCLUÍDA
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/dashboard/app.go`, `internal/dashboard/model.go`

**Notas de Implementação:**
- App struct com tea.Program para gerenciar o dashboard
- Model struct com lastUpdate, quitting, width, height
- View renderiza header com título "Prediction Market Bot" e timestamp
- Tick a cada segundo para atualizar timestamp
- Suporte para 'q' e 'ctrl+c' para sair
- Lipgloss para estilização (cores, bordas, padding)
- Flag --dashboard no main.go para ativar modo dashboard

**1. The Outer Gate (Behavior Test)**
- **Teste:** Dashboard mostra header com título e timestamp.
- **Comando:** `go run cmd/bot/main.go --dashboard`

**2. The Implementation Steps**
- Passo A: Setup bubbletea com Model, Update, View
- Passo B: Renderizar header com timestamp atualizando a cada segundo

---

## Fatia 12.2: Dashboard - Seções Bankroll e Positions
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/dashboard/views/bankroll.go`, `internal/dashboard/views/positions.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Mostra saldo por plataforma e lista de positions abertas.
- **Comando:** `go run cmd/bot/main.go --dashboard`

**2. The Implementation Steps**
- Passo A: View bankroll com delta desde initial
- Passo B: View positions com PnL colorido (verde/vermelho)

---

## Fatia 12.3: Dashboard - Seções Stats e Keyboard
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/dashboard/views/stats.go`, `internal/dashboard/keys.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Mostra stats (win rate, total PnL), Q para sair funciona.
- **Comando:** `go run cmd/bot/main.go --dashboard`

**2. The Implementation Steps**
- Passo A: View stats com trades, win rate, PnL, drawdown
- Passo B: Key handlers: Q (quit), P (pause), R (refresh)

---

# FASE 13: LIVE TRADING (CUIDADO!)

## Fatia 13.1: Polymarket - Place Order REAL
**Risco de Contexto:** ALTO ⚠️ - REQUER REVISÃO MANUAL
**Arquivos Esperados:** `internal/platform/polymarket/orders.go` (update)

**1. The Outer Gate (Behavior Test)**
- **Teste:** Com `--live` flag, ordem é submetida à CLOB API real.
- **Comando:** Manual testing with minimal amount ($1)

**2. The Implementation Steps**
- Passo A: Remover DRY_RUN quando flag `--live` presente
- Passo B: Adicionar confirmação interativa antes de ativar
- Passo C: Log MUITO verbose para auditoria completa

---

## Fatia 13.2: Kalshi - Place Order REAL
**Risco de Contexto:** ALTO ⚠️ - REQUER REVISÃO MANUAL
**Arquivos Esperados:** `internal/platform/kalshi/orders.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Com `--live` flag, ordem é submetida à API Kalshi real.
- **Comando:** Manual testing with minimal amount ($1)

**2. The Implementation Steps**
- Passo A: Implementar `PlaceOrder(order, dryRun) (OrderResult, error)`
- Passo B: Mesmas proteções da Polymarket

---

# FASE 14: BACKTESTING (OPCIONAL)

## Fatia 14.1: Backtest - Data Loader
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/backtest/loader.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Carregar CSV de histórico e retornar timeline.
- **Comando:** `go test ./internal/backtest/... -v`

**2. The Implementation Steps**
- Passo A: Definir formato CSV para dados históricos
- Passo B: Implementar parser para HistoricalMarket

---

## Fatia 14.2: Backtest - Simulation Engine
**Risco de Contexto:** Médio
**Arquivos Esperados:** `internal/backtest/engine.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Simular período com dados mock retorna métricas.
- **Comando:** `go test ./internal/backtest/... -v`

**2. The Implementation Steps**
- Passo A: Implementar loop de simulação timestamp por timestamp
- Passo B: Usar mesma lógica de Scanner/PositionManager

---

## Fatia 14.3: Backtest - Report
**Risco de Contexto:** Baixo
**Arquivos Esperados:** `internal/backtest/report.go`

**1. The Outer Gate (Behavior Test)**
- **Teste:** Gerar sumário no terminal e CSV de trades.
- **Comando:** `go run cmd/bot/main.go backtest --start 2024-01-01 --end 2024-06-30`

**2. The Implementation Steps**
- Passo A: Formatar BacktestResult para terminal
- Passo B: Exportar TradeLog para CSV

---

# CHECKPOINTS DE VALIDAÇÃO

Após CADA fatia:

```bash
go build ./...
go test ./...
go vet ./...
```

Se falhar, NÃO avançar.

---

# RESUMO

| Fase | Descrição | Fatias | Crítico |
|------|-----------|--------|---------|
| 1 | Fundações | 4 | ✅ |
| 2 | Fetch Preço BTC | 4 | ✅ |
| 3 | Listar Mercados Poly | 4 | ✅ |
| 4 | Listar Mercados Kalshi | 4 | ✅ |
| 5 | Analisar Volatilidade | 3 | ✅ |
| 6 | Escanear Mercados | 3 | ✅ |
| 7 | Calcular Posição | 2 | ✅ |
| 8 | Abrir Posição (DRY) | 3 | ✅ |
| 9 | Monitorar e Sair | 3 | ✅ |
| 10 | Bot Loop | 3 | ✅ |
| 11 | Learning System | 3 | ✅ |
| 12 | Dashboard | 3 | 🔄 |
| 13 | Live Trading | 2 | ⚠️ |
| 14 | Backtesting | 3 | |

**Total: 44 micro-fatias**

---

# ORDEM DE EXECUÇÃO

```
1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11 → 12 → 13 → 14
```

Fases 1-10 são o MVP funcional (DRY-RUN).
Fase 13 ativa dinheiro real - REQUER REVISÃO MANUAL COMPLETA.

---

# NOTAS DE SEGURANÇA

1. **NUNCA commitar chaves privadas** - use .env ou variáveis de ambiente
2. **Fase 13 (Live Trading)**: SEMPRE testar com valor mínimo ($1) primeiro
3. **Todos os testes contra APIs reais**: respeitar rate limits
4. **Default é DRY-RUN**: flag `--live` necessária para ordens reais
