package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	http.HandleFunc("/api/cmc/global", func(w http.ResponseWriter, r *http.Request) {
		apiKey := os.Getenv("VITE_CMC_API_KEY")
		if apiKey == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "API key not set"}`))
			return
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://pro-api.coinmarketcap.com/v1/global-metrics/quotes/latest", nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Failed to create request"}`))
			return
		}
		req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{"error": "Failed to fetch from CoinMarketCap"}`))
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})
	log.Println("CMC proxy running on :5002/api/cmc/global")
	http.ListenAndServe(":5002", nil)
}
