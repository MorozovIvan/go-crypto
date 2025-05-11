# Crypto Market Analysis Metrics & Algorithm

This document describes all the metrics used in the Market Analysis dashboard and explains how the buy/sell/hold algorithm works.

---

## Metrics Overview

### 1. **Fear & Greed Index**

- **What:** Measures market sentiment (0 = extreme fear, 100 = extreme greed).
- **Why:** Contrarian indicator. Buy when others are fearful, sell when greedy.
- **Signal:**
  - Buy: <30
  - Sell: >70
  - Hold: 30–70

### 2. **Altcoin Season Index**

- **What:** Tracks if altcoins are outperforming Bitcoin (score >75 = altcoin season).
- **Why:** Helps decide allocation between Bitcoin and altcoins.
- **Signal:**
  - Buy: >75 (altcoin season)
  - Sell: <25 (Bitcoin season)
  - Hold: 25–75

### 3. **BTC Dominance**

- **What:** Bitcoin's market cap as a % of total crypto market cap.
- **Why:** Indicates market direction and risk appetite.
- **Signal:**
  - Buy: Falling and <50%
  - Sell: Rising and >60%
  - Hold: Stable or 50–60%

### 4. **Stablecoin Supply Ratio (SSR)**

- **What:** Ratio of Bitcoin market cap to stablecoin market cap.
- **Why:** Low SSR = high buying power on sidelines.
- **Signal:**
  - Buy: Low (bottom 25% of historical range)
  - Sell: High (top 25%)
  - Hold: Middle 50%

### 5. **RSI (Relative Strength Index)**

- **What:** Momentum oscillator (0–100).
- **Why:** Identifies overbought/oversold conditions.
- **Signal:**
  - Buy: <30 (oversold)
  - Sell: >70 (overbought)
  - Hold: 30–70

### 6. **Total Market Cap Change (%)**

- **What:** 7-day % change in total crypto market cap.
- **Why:** Shows overall market health.
- **Signal:**
  - Buy: >+5%
  - Sell: <-5%
  - Hold: ±5%

### 7. **Google Trends**

- **What:** Search volume for "crypto"/"bitcoin".
- **Why:** Captures retail sentiment.
- **Signal:**
  - Buy: Low (bottom 25% of 90-day range)
  - Sell: High (top 25%)
  - Hold: Middle 50%

### 8. **Moving Averages (MA)**

- **What:** 50-day and 200-day simple moving averages.
- **Why:** Trend confirmation.
- **Signal:**
  - Buy: Golden cross (50d > 200d)
  - Sell: Death cross (50d < 200d)
  - Hold: No cross

### 9. **Volume Trend**

- **What:** Price movement with trading volume.
- **Why:** Confirms trend strength.
- **Signal:**
  - Buy: High and rising with price
  - Sell: Low/declining with price
  - Hold: Stable/uncorrelated

### 10. **Exchange Flows**

- **What:** Net inflow/outflow of coins to exchanges.
- **Why:** Inflows = sell pressure, outflows = accumulation.
- **Signal:**
  - Buy: Outflows rising
  - Sell: Inflows rising
  - Hold: Stable

### 11. **Active Addresses**

- **What:** Number of unique active wallets.
- **Why:** Proxy for network activity/adoption.
- **Signal:**
  - Buy: Rising
  - Sell: Falling
  - Hold: Stable

### 12. **Whale Transactions**

- **What:** Large on-chain transfers.
- **Why:** Whales can move markets.
- **Signal:**
  - Buy: Accumulation
  - Sell: Distribution
  - Hold: Neutral

### 13. **Bollinger Bands Width**

- **What:** Measures price volatility.
- **Why:** Squeeze = breakout likely, wide = reversal risk.
- **Signal:**
  - Buy: Squeeze + breakout
  - Sell: Wide + reversal
  - Hold: Neutral

### 14. **Funding Rate**

- **What:** Perpetual futures funding rate.
- **Why:** Shows long/short bias.
- **Signal:**
  - Buy: Negative/extreme negative
  - Sell: Positive/extreme positive
  - Hold: Neutral

### 15. **Open Interest**

- **What:** Total value of open derivatives contracts.
- **Why:** High/rising = strong trend, but can mean crowded trade.
- **Signal:**
  - Buy: High and rising with price
  - Sell: High and falling with price
  - Hold: Stable

### 16. **ETH/BTC Ratio**

- **What:** ETH price divided by BTC price.
- **Why:** Indicates risk-on/risk-off and altcoin rotation.
- **Signal:**
  - Buy: Rising
  - Sell: Falling
  - Hold: Stable

---

## Algorithm Overview

1. **Data Collection**

   - Fetch real-time data for all metrics (APIs, on-chain, etc.).
   - Normalize data as needed (e.g., RSI and Fear & Greed are 0–100).

2. **Scoring System**

   - Each metric is scored:
     - **Buy**: +1
     - **Sell**: -1
     - **Hold**: 0
   - Each metric has a weight (importance in the final score).

3. **Weighted Score Calculation**

   - `Total Score = Σ(Metric Score × Weight)`

4. **Signal Thresholds**

   - **Buy**: Total Score > 0.5
   - **Sell**: Total Score < -0.5
   - **Hold**: -0.5 ≤ Total Score ≤ 0.5

5. **Asset Allocation**

   - If Altcoin Season Index >75, favor altcoins.
   - If BTC Dominance rising and Altcoin Index <25, favor Bitcoin.
   - Otherwise, allocate based on RSI and Fear & Greed.

6. **Error Handling**
   - If any critical metric is missing, the signal is set to "Hold" and a warning is displayed.

---

## Example

If most metrics are bullish (Buy), the algorithm issues a **Buy** signal and suggests whether to focus on Bitcoin or altcoins. If metrics are mixed or missing, the signal is **Hold**. If most are bearish, the signal is **Sell**.

---

## Notes

- This framework is extensible: add/remove metrics as needed.
- Backtest and adjust weights/thresholds for your strategy.
- Always use risk management and do not rely solely on algorithmic signals for trading decisions.
