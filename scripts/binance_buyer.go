package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BinanceClient represents the Binance API client
type BinanceClient struct {
	apiKey     string
	secretKey  string
	baseURL    string
	httpClient *http.Client
}

// OrderResponse represents the response from Binance order API
type OrderResponse struct {
	Symbol        string `json:"symbol"`
	OrderID       int64  `json:"orderId"`
	ClientOrderID string `json:"clientOrderId"`
	TransactTime  int64  `json:"transactTime"`
	Price         string `json:"price"`
	OrigQty       string `json:"origQty"`
	ExecutedQty   string `json:"executedQty"`
	Status        string `json:"status"`
	Type          string `json:"type"`
	Side          string `json:"side"`
}

// TickerPrice represents the current price of a symbol
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// AccountInfo represents the account information from Binance
type AccountInfo struct {
	Balances []Balance `json:"balances"`
}

// Balance represents a single balance in the account
type Balance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

// NewBinanceClient creates a new Binance API client
func NewBinanceClient(apiKey, secretKey string) *BinanceClient {
	return &BinanceClient{
		apiKey:     apiKey,
		secretKey:  secretKey,
		baseURL:    "https://api.binance.com",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// generateSignature generates HMAC SHA256 signature for Binance API
func (c *BinanceClient) generateSignature(query string) string {
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

// GetAccountInfo gets the account information including balances
func (c *BinanceClient) GetAccountInfo() (*AccountInfo, error) {
	// Create query parameters
	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))
	params.Set("recvWindow", "5000")

	// Generate signature
	signature := c.generateSignature(params.Encode())
	params.Set("signature", signature)

	// Create request
	req, err := http.NewRequest("GET", c.baseURL+"/api/v3/account", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers and query parameters
	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	req.URL.RawQuery = params.Encode()

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse response
	var accountInfo AccountInfo
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &accountInfo, nil
}

// GetUSDTBalance gets the USDT balance from the account
func (c *BinanceClient) GetUSDTBalance() (float64, error) {
	accountInfo, err := c.GetAccountInfo()
	if err != nil {
		return 0, err
	}

	for _, balance := range accountInfo.Balances {
		if balance.Asset == "USDT" {
			free, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing USDT balance: %v", err)
			}
			return free, nil
		}
	}

	return 0, fmt.Errorf("USDT balance not found")
}

// GetAssetBalance gets the free balance for a given asset symbol (e.g., BTC, ETH)
func (c *BinanceClient) GetAssetBalance(asset string) (float64, error) {
	accountInfo, err := c.GetAccountInfo()
	if err != nil {
		return 0, err
	}

	for _, balance := range accountInfo.Balances {
		if balance.Asset == asset {
			free, err := strconv.ParseFloat(balance.Free, 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing %s balance: %v", asset, err)
			}
			return free, nil
		}
	}

	return 0, fmt.Errorf("%s balance not found", asset)
}

// GetCurrentPrice gets the current price of a symbol
func (c *BinanceClient) GetCurrentPrice(symbol string) (float64, error) {
	// Create request
	req, err := http.NewRequest("GET", c.baseURL+"/api/v3/ticker/price", nil)
	if err != nil {
		return 0, fmt.Errorf("error creating request: %v", err)
	}

	// Add query parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	req.URL.RawQuery = params.Encode()

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response: %v", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error: %s", string(body))
	}

	// Parse response
	var tickerPrice TickerPrice
	if err := json.Unmarshal(body, &tickerPrice); err != nil {
		return 0, fmt.Errorf("error parsing response: %v", err)
	}

	// Convert price string to float64
	price, err := strconv.ParseFloat(tickerPrice.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing price: %v", err)
	}

	return price, nil
}

// PlaceOrder places a market order on Binance for the given side using quote quantity
func (c *BinanceClient) PlaceOrder(symbol string, side string, quoteQuantity float64) (*OrderResponse, error) {
	// Create query parameters
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", strings.ToUpper(side))
	params.Set("type", "MARKET")
	params.Set("quoteOrderQty", fmt.Sprintf("%.8f", quoteQuantity))
	params.Set("timestamp", strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10))
	params.Set("recvWindow", "5000")

	// Generate signature
	signature := c.generateSignature(params.Encode())
	params.Set("signature", signature)

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+"/api/v3/order", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add headers and query parameters
	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	req.URL.RawQuery = params.Encode()

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	// Parse response
	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &orderResp, nil
}

func parseDuration(durationStr string) (time.Duration, error) {
	// Regular expression to match the duration pattern
	re := regexp.MustCompile(`^(\d+)([smhHdDwWM])$`)
	matches := re.FindStringSubmatch(durationStr)

	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid duration format. Use format like '30s', '30m', '2H', '1D', '1W', or '1M'")
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number in duration: %v", err)
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "s": // Second
		duration = time.Duration(value) * time.Second
	case "m": // Minute
		duration = time.Duration(value) * time.Minute
	case "H": // Hour
		duration = time.Duration(value) * time.Hour
	case "D": // Day
		duration = time.Duration(value) * 24 * time.Hour
	case "W": // Week
		duration = time.Duration(value) * 7 * 24 * time.Hour
	case "M": // Month (approximated to 30 days)
		duration = time.Duration(value) * 30 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("invalid time unit. Use s, m, H, D, W, or M")
	}

	return duration, nil
}

func main() {
	// Parse command line flags
	apiKey := flag.String("api-key", "", "Binance API key")
	secretKey := flag.String("secret-key", "", "Binance secret key")
	symbol := flag.String("symbol", "BTCUSDT", "Trading pair symbol")
	totalRunTime := flag.String("total-run-time", "1H", "Total run time (e.g., 30m, 2H, 1D, 1W, 1M)")
	totalAmount := flag.Float64("total-amount", -1, "Total USDT amount to use for buying (optional, default: use full balance)")
	side := flag.String("side", "BUY", "Order side: BUY or SELL")
	flag.Parse()

	// Validate required flags
	if *apiKey == "" || *secretKey == "" {
		log.Fatal("API key and secret key are required")
	}

	// Parse total run time
	duration, err := parseDuration(*totalRunTime)
	if err != nil {
		log.Fatalf("Error parsing total run time: %v", err)
	}

	// Create Binance client
	client := NewBinanceClient(*apiKey, *secretKey)

	// Set up logging
	log.SetPrefix("[Binance Buyer] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Validate and normalize side
	sideUpper := strings.ToUpper(*side)
	if sideUpper != "BUY" && sideUpper != "SELL" {
		log.Fatalf("Invalid side: %s. Use BUY or SELL.", *side)
	}

	// Detect quote asset (supporting USDT quotes)
	quoteAsset := "USDT"
	if !strings.HasSuffix(*symbol, quoteAsset) {
		log.Fatalf("Unsupported quote asset. Only %s quote pairs supported, got: %s", quoteAsset, *symbol)
	}

	// Fetch current price once for SELL calculations and logging
	currentPrice, err := client.GetCurrentPrice(*symbol)
	if err != nil {
		log.Fatalf("Error getting current price for %s: %v", *symbol, err)
	}

	// Determine available quote amount based on side
	var availableQuote float64
	if sideUpper == "BUY" {
		usdtBalance, err := client.GetUSDTBalance()
		if err != nil {
			log.Fatalf("Error getting %s balance: %v", quoteAsset, err)
		}
		availableQuote = usdtBalance
	} else {
		baseAsset := strings.TrimSuffix(*symbol, quoteAsset)
		baseBalance, err := client.GetAssetBalance(baseAsset)
		if err != nil {
			log.Fatalf("Error getting %s balance: %v", baseAsset, err)
		}
		availableQuote = baseBalance * currentPrice
	}

	// Determine the total quote amount to use
	amountToUse := availableQuote
	if *totalAmount > 0 {
		if *totalAmount > availableQuote {
			log.Fatalf("Specified total amount (%.8f) is greater than available %s amount (%.8f)", *totalAmount, quoteAsset, availableQuote)
		}
		amountToUse = *totalAmount
	}

	// Calculate per-second USDT amount
	totalSeconds := duration.Seconds()
	usdtPerSecond := math.Round((amountToUse/totalSeconds)*100) / 100

	log.Printf("Initial available %s (quote) amount: %.2f", quoteAsset, availableQuote)
	log.Printf("Total run time: %s (%.0f seconds)", duration, totalSeconds)
	log.Printf("%s amount per second: %.8f", quoteAsset, usdtPerSecond)
	log.Printf("Total %s to %s: %.8f", quoteAsset, strings.ToLower(sideUpper), usdtPerSecond*totalSeconds)
	log.Printf("Starting automated %s for %s at price %.8f", strings.ToLower(sideUpper), *symbol, currentPrice)

	if usdtPerSecond < 1.0 {
		// Calculate number of intervals (each interval trades 1 of quote asset)
		nIntervals := int(amountToUse)
		if nIntervals == 0 {
			log.Printf("%s amount to use is less than 1. Nothing to do.", quoteAsset)
			return
		}
		intervalDuration := time.Duration(duration.Seconds()/float64(nIntervals)) * time.Second
		log.Printf("Per-second amount < 1 %s. Will trade 1 %s every %s, %d times.", quoteAsset, quoteAsset, intervalDuration, nIntervals)
		for i := 0; i < nIntervals; i++ {
			if amountToUse < 1.0 {
				log.Printf("Insufficient %s amount to use (%.2f) for next order (1.0). Stopping.", quoteAsset, amountToUse)
				break
			}
			order, err := client.PlaceOrder(*symbol, sideUpper, 1.0)
			if err != nil {
				log.Printf("Error placing order: %v", err)
			} else {
				log.Printf("Order placed successfully: OrderID=%d, Status=%s, ExecutedQty=%s, Price=%s",
					order.OrderID, order.Status, order.ExecutedQty, order.Price)
				amountToUse -= 1.0
				log.Printf("Remaining %s amount to use: %.2f", quoteAsset, amountToUse)
			}
			if i < nIntervals-1 {
				time.Sleep(intervalDuration)
			}
		}
	} else {
		// Calculate number of trades to be made
		numberOfTrades := int(totalSeconds)
		if numberOfTrades == 0 {
			log.Printf("Total run time is less than 1 second. Nothing to buy.")
			return
		}
		usdtPerTrade := amountToUse / float64(numberOfTrades)
		log.Printf("Will make %d trades, %.8f %s per trade", numberOfTrades, usdtPerTrade, quoteAsset)

		// Main loop
		for i := 0; i < numberOfTrades; i++ {
			if amountToUse < usdtPerTrade {
				log.Printf("Insufficient %s amount to use (%.2f) for next order (%.8f). Stopping.", quoteAsset, amountToUse, usdtPerTrade)
				break
			}
			order, err := client.PlaceOrder(*symbol, sideUpper, usdtPerTrade)
			if err != nil {
				log.Printf("Error placing order: %v", err)
			} else {
				log.Printf("Order placed successfully: OrderID=%d, Status=%s, ExecutedQty=%s, Price=%s",
					order.OrderID, order.Status, order.ExecutedQty, order.Price)
				amountToUse -= usdtPerTrade
				log.Printf("Remaining %s amount to use: %.2f", quoteAsset, amountToUse)
			}
			time.Sleep(time.Second)
		}
	}

	log.Printf("Trading completed. Final %s amount remaining to use: %.2f", quoteAsset, amountToUse)
}
