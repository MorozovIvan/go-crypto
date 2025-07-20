package solana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func getSolanaBalance(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(400, gin.H{"error": "Wallet address is required"})
		return
	}

	// Validate Solana address format (basic check)
	if len(address) < 32 || len(address) > 44 {
		c.JSON(400, gin.H{"error": "Invalid Solana address format"})
		return
	}

	// Try multiple RPC endpoints to get balance
	rpcEndpoints := []string{
		"https://solana-mainnet.g.alchemy.com/v2/demo",
		"https://api.mainnet-beta.solana.com",
		"https://solana-api.projectserum.com",
		"https://rpc.ankr.com/solana",
	}

	for _, endpoint := range rpcEndpoints {
		balance, err := fetchSolanaBalanceFromRPC(endpoint, address)
		if err == nil {
			c.JSON(200, gin.H{
				"address":   address,
				"balance":   balance,
				"endpoint":  endpoint,
				"timestamp": time.Now().Unix(),
			})
			return
		}
		log.Printf("RPC endpoint %s failed for address %s: %v", endpoint, address, err)
	}

	// If all endpoints fail, return error
	c.JSON(500, gin.H{
		"error":     "Failed to fetch balance from all RPC endpoints",
		"address":   address,
		"timestamp": time.Now().Unix(),
	})
}

// fetchSolanaBalanceFromRPC fetches balance from a specific RPC endpoint
func fetchSolanaBalanceFromRPC(endpoint, address string) (float64, error) {
	// Create JSON-RPC request
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getBalance",
		"params":  []string{address},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make POST request
	resp, err := client.Post(endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Parse response
	var response struct {
		Result struct {
			Value int64 `json:"value"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %v", err)
	}

	if response.Error != nil {
		return 0, fmt.Errorf("RPC error: %s", response.Error.Message)
	}

	// Convert lamports to SOL
	balanceInSol := float64(response.Result.Value) / 1000000000.0
	return balanceInSol, nil
}
