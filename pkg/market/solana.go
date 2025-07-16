package market

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaService struct {
	clients     []*rpc.Client
	currentIdx  int
	rateLimiter *time.Ticker
	mu          sync.Mutex
}

type TokenData struct {
	Time           time.Time
	Price          float64
	Volume         float64
	Performance    float64
	MaxPerformance float64
}

func NewSolanaService() *SolanaService {
	// List of public RPC endpoints
	endpoints := []string{
		"https://api.mainnet-beta.solana.com",
		"https://solana-api.projectserum.com",
		"https://rpc.ankr.com/solana",
		"https://solana-mainnet.g.alchemy.com/v2/demo",
	}

	clients := make([]*rpc.Client, len(endpoints))
	for i, endpoint := range endpoints {
		clients[i] = rpc.New(endpoint)
	}

	return &SolanaService{
		clients:     clients,
		currentIdx:  0,
		rateLimiter: time.NewTicker(100 * time.Millisecond), // 10 requests per second
		mu:          sync.Mutex{},
	}
}

func (s *SolanaService) getNextClient() *rpc.Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := s.clients[s.currentIdx]
	s.currentIdx = (s.currentIdx + 1) % len(s.clients)
	return client
}

func (s *SolanaService) GetTokenData(contractAddress string) (*TokenData, error) {
	// Parse the public key
	pubKey, err := solana.PublicKeyFromBase58(contractAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid contract address: %v", err)
	}

	// Try each client until we get a successful response
	var sigs []*rpc.TransactionSignature
	var lastError error

	for i := 0; i < len(s.clients); i++ {
		<-s.rateLimiter.C // Rate limiting
		client := s.getNextClient()

		sigs, err = client.GetSignaturesForAddress(context.Background(), pubKey)
		if err == nil {
			break
		}
		lastError = err
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get signatures after trying all endpoints: %v", lastError)
	}

	if len(sigs) == 0 {
		return nil, fmt.Errorf("no transactions found for this token")
	}

	// Calculate metrics from recent transactions
	var totalVolume float64
	var maxPrice float64
	var minPrice float64
	var firstPrice float64
	var lastPrice float64

	// Process only the last 100 transactions if we have more
	maxTx := 100
	if len(sigs) > maxTx {
		sigs = sigs[:maxTx]
	}

	for i, sig := range sigs {
		<-s.rateLimiter.C // Rate limiting
		client := s.getNextClient()

		tx, err := client.GetTransaction(
			context.Background(),
			sig.Signature,
			&rpc.GetTransactionOpts{
				Encoding: solana.EncodingBase64,
			},
		)
		if err != nil {
			continue
		}

		// Extract price and volume from transaction
		if tx.Meta != nil && len(tx.Meta.PreTokenBalances) > 0 && len(tx.Meta.PostTokenBalances) > 0 {
			pre := tx.Meta.PreTokenBalances[0].UiTokenAmount.UiAmount
			post := tx.Meta.PostTokenBalances[0].UiTokenAmount.UiAmount
			if pre != nil && post != nil {
				volume := *post - *pre
				totalVolume += volume

				if i == 0 {
					lastPrice = *post
				}
				if i == len(sigs)-1 {
					firstPrice = *pre
				}
				if *post > maxPrice {
					maxPrice = *post
				}
				if minPrice == 0 || *post < minPrice {
					minPrice = *post
				}
			}
		}
	}

	// Calculate performance
	var performance float64
	var maxPerformance float64
	if firstPrice > 0 {
		performance = ((lastPrice - firstPrice) / firstPrice) * 100
		maxPerformance = ((maxPrice - firstPrice) / firstPrice) * 100
	}

	return &TokenData{
		Time:           time.Now(),
		Price:          lastPrice,
		Volume:         totalVolume,
		Performance:    performance,
		MaxPerformance: maxPerformance,
	}, nil
}

func (s *SolanaService) Close() {
	s.rateLimiter.Stop()
}
