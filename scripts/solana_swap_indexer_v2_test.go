package main

import (
	"testing"
)

type metaType struct {
	LogMessages      []string `json:"logMessages"`
	PreTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"preTokenBalances"`
	PostTokenBalances []struct {
		AccountIndex  int    `json:"accountIndex"`
		Mint          string `json:"mint"`
		Owner         string `json:"owner"`
		ProgramId     string `json:"programId"`
		UiTokenAmount struct {
			Amount   string  `json:"amount"`
			Decimals int     `json:"decimals"`
			UiAmount float64 `json:"uiAmount"`
		} `json:"uiTokenAmount"`
	} `json:"postTokenBalances"`
}

func TestParseSwapFromLog(t *testing.T) {
	mint := "7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump"
	tests := []struct {
		name        string
		logLine     string
		expectIn    float64
		expectOut   float64
		expectPrice float64
		direction   string
	}{
		{
			name:     "Basic buy log",
			logLine:  "Program log: Swap: amount_in: 100.0 amount_out: 0.01 token_in: USDC token_out: 7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump price: 10000.0",
			expectIn: 100.0, expectOut: 0.01, expectPrice: 10000.0, direction: "buy",
		},
		{
			name:     "Basic sell log",
			logLine:  "Program log: Swap: amount_in: 0.01 amount_out: 100.0 token_in: 7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump token_out: USDC price: 10000.0",
			expectIn: 0.01, expectOut: 100.0, expectPrice: 10000.0, direction: "sell",
		},
		{
			name:     "No price log, infer price",
			logLine:  "Program log: Swap: amount_in: 0.02 amount_out: 200.0 token_in: 7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump token_out: USDC",
			expectIn: 0.02, expectOut: 200.0, expectPrice: 10000.0, direction: "sell",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			meta := metaType{LogMessages: []string{tc.logLine}}
			event := ParseSwapFromLog(tc.logLine, "sig", 1, 1234567890, mint, meta)
			if event == nil {
				t.Fatalf("Expected event, got nil")
			}
			if event.AmountIn != tc.expectIn {
				t.Errorf("AmountIn: got %v, want %v", event.AmountIn, tc.expectIn)
			}
			if event.AmountOut != tc.expectOut {
				t.Errorf("AmountOut: got %v, want %v", event.AmountOut, tc.expectOut)
			}
			if event.Direction != tc.direction {
				t.Errorf("Direction: got %v, want %v", event.Direction, tc.direction)
			}
			if event.Price != tc.expectPrice {
				t.Errorf("Price: got %v, want %v", event.Price, tc.expectPrice)
			}
		})
	}
}

func TestInferSwapFromBalances(t *testing.T) {
	mint := "7UUUHUxGrDunmbL5g3veDMtb5ai7gCexiNLZtBJgpump"
	meta := metaType{
		PreTokenBalances: []struct {
			AccountIndex  int    `json:"accountIndex"`
			Mint          string `json:"mint"`
			Owner         string `json:"owner"`
			ProgramId     string `json:"programId"`
			UiTokenAmount struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			} `json:"uiTokenAmount"`
		}{
			{Mint: mint, UiTokenAmount: struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			}{UiAmount: 10.0}},
			{Mint: "USDC", UiTokenAmount: struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			}{UiAmount: 1000.0}},
		},
		PostTokenBalances: []struct {
			AccountIndex  int    `json:"accountIndex"`
			Mint          string `json:"mint"`
			Owner         string `json:"owner"`
			ProgramId     string `json:"programId"`
			UiTokenAmount struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			} `json:"uiTokenAmount"`
		}{
			{Mint: mint, UiTokenAmount: struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			}{UiAmount: 9.0}},
			{Mint: "USDC", UiTokenAmount: struct {
				Amount   string  `json:"amount"`
				Decimals int     `json:"decimals"`
				UiAmount float64 `json:"uiAmount"`
			}{UiAmount: 1100.0}},
		},
	}
	event := InferSwapFromBalances("sig", 1, 1234567890, mint, meta)
	if event == nil {
		t.Fatalf("Expected event, got nil")
	}
	if event.AmountIn != 1.0 {
		t.Errorf("AmountIn: got %v, want 1.0", event.AmountIn)
	}
	if event.AmountOut != 100.0 {
		t.Errorf("AmountOut: got %v, want 100.0", event.AmountOut)
	}
	if event.Direction != "sell" {
		t.Errorf("Direction: got %v, want sell", event.Direction)
	}
	if event.Price != 100.0 {
		t.Errorf("Price: got %v, want 100.0", event.Price)
	}
}
