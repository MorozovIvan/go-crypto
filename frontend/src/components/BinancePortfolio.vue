<template>
  <div class="portfolio-container">
    <h2>Binance Portfolio Analysis</h2>
    
    <!-- Portfolio Summary -->
    <div class="portfolio-summary">
      <div class="summary-card">
        <h3>Total Portfolio Value</h3>
        <div class="value">{{ formatCurrency(totalValue) }}</div>
        <div class="change" :class="{ 'positive': portfolioChange >= 0, 'negative': portfolioChange < 0 }">
          {{ portfolioChange >= 0 ? '+' : '' }}{{ portfolioChange.toFixed(2) }}%
        </div>
      </div>
      
      <div class="summary-card">
        <h3>24h Volume</h3>
        <div class="value">{{ formatCurrency(volume24h) }}</div>
      </div>
      
      <div class="summary-card">
        <h3>Number of Assets</h3>
        <div class="value">{{ assets.length }}</div>
      </div>
    </div>

    <!-- Asset List -->
    <div class="assets-list">
      <div v-for="asset in assets" :key="asset.symbol" class="asset-card">
        <div class="asset-header">
          <div class="asset-info">
            <h4>{{ asset.symbol }}</h4>
            <div class="asset-balance">{{ formatNumber(asset.amount) }} {{ asset.symbol }}</div>
          </div>
          <div class="asset-value">
            <div class="value">{{ formatCurrency(asset.value) }}</div>
            <div class="change" :class="{ 'positive': asset.change >= 0, 'negative': asset.change < 0 }">
              {{ asset.change >= 0 ? '+' : '' }}{{ asset.change.toFixed(2) }}%
            </div>
          </div>
        </div>
        
        <!-- Price Chart -->
        <div class="price-chart">
          <canvas :id="'chart-' + asset.symbol"></canvas>
        </div>
        
        <!-- Asset Metrics -->
        <div class="asset-metrics">
          <div class="metric">
            <span>24h High:</span>
            <span>{{ formatCurrency(asset.high24h) }}</span>
          </div>
          <div class="metric">
            <span>24h Low:</span>
            <span>{{ formatCurrency(asset.low24h) }}</span>
          </div>
          <div class="metric">
            <span>24h Volume:</span>
            <span>{{ formatCurrency(asset.volume24h) }}</span>
          </div>
          <div class="metric">
            <span>Market Cap:</span>
            <span>{{ formatCurrency(asset.marketCap) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import Chart from 'chart.js/auto'
import axios from 'axios'

export default {
  name: 'BinancePortfolio',
  setup() {
    const assets = ref([])
    const totalValue = ref(0)
    const portfolioChange = ref(0)
    const volume24h = ref(0)
    const charts = {}

    const fetchPortfolioData = async () => {
      try {
        const response = await axios.get('/api/binance/portfolio')
        assets.value = response.data.assets
        totalValue.value = response.data.totalValue
        portfolioChange.value = response.data.portfolioChange
        volume24h.value = response.data.volume24h
        
        // Initialize charts after data is loaded
        assets.value.forEach(asset => {
          initializeChart(asset)
        })
      } catch (error) {
        console.error('Error fetching portfolio data:', error)
      }
    }

    const initializeChart = (asset) => {
      const ctx = document.getElementById(`chart-${asset.symbol}`)
      if (!ctx) return

      if (charts[asset.symbol]) {
        charts[asset.symbol].destroy()
      }

      charts[asset.symbol] = new Chart(ctx, {
        type: 'line',
        data: {
          labels: asset.priceHistory.map(p => new Date(p.timestamp).toLocaleDateString()),
          datasets: [{
            label: `${asset.symbol} Price`,
            data: asset.priceHistory.map(p => p.price),
            borderColor: asset.change >= 0 ? '#4CAF50' : '#F44336',
            tension: 0.4,
            fill: false
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: {
            legend: {
              display: false
            }
          },
          scales: {
            y: {
              beginAtZero: false
            }
          }
        }
      })
    }

    const formatCurrency = (value) => {
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
      }).format(value)
    }

    const formatNumber = (value) => {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 8
      }).format(value)
    }

    onMounted(() => {
      fetchPortfolioData()
    })

    return {
      assets,
      totalValue,
      portfolioChange,
      volume24h,
      formatCurrency,
      formatNumber
    }
  }
}
</script>

<style scoped>
.portfolio-container {
  padding: 20px;
}

.portfolio-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.summary-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.summary-card h3 {
  margin: 0 0 10px 0;
  color: #666;
  font-size: 1rem;
}

.summary-card .value {
  font-size: 1.5rem;
  font-weight: bold;
  color: #333;
}

.change {
  font-size: 0.9rem;
  margin-top: 5px;
}

.change.positive {
  color: #4CAF50;
}

.change.negative {
  color: #F44336;
}

.assets-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.asset-card {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.asset-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 15px;
}

.asset-info h4 {
  margin: 0 0 5px 0;
  font-size: 1.2rem;
}

.asset-balance {
  color: #666;
  font-size: 0.9rem;
}

.price-chart {
  height: 200px;
  margin: 15px 0;
}

.asset-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-top: 15px;
}

.metric {
  display: flex;
  justify-content: space-between;
  font-size: 0.9rem;
  color: #666;
}

.metric span:last-child {
  font-weight: 500;
  color: #333;
}
</style> 