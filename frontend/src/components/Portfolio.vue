<template>
  <div class="portfolio-container">
    <div class="portfolio-header">
      <h2>Portfolio Overview</h2>
      <div class="portfolio-summary">
        <div class="total-value">
          <span class="label">Total Value:</span>
          <span class="value">${{ formatNumber(portfolio.totalValue) }}</span>
        </div>
        <div class="portfolio-change" :class="{ 'positive': portfolio.portfolioChange > 0, 'negative': portfolio.portfolioChange < 0 }">
          <span class="label">24h Change:</span>
          <span class="value">{{ formatNumber(portfolio.portfolioChange) }}%</span>
        </div>
      </div>
    </div>

    <div class="assets-grid">
      <div v-for="asset in portfolio.assets" :key="asset.symbol" class="asset-card">
        <div class="asset-header">
          <h3>{{ asset.symbol }}</h3>
          <div class="asset-value">${{ formatNumber(asset.value) }}</div>
        </div>
        
        <div class="asset-details">
          <div class="detail-row">
            <span class="label">Amount:</span>
            <span class="value">{{ formatNumber(asset.amount) }}</span>
          </div>
          <div class="detail-row">
            <span class="label">24h Change:</span>
            <span class="value" :class="{ 'positive': asset.change > 0, 'negative': asset.change < 0 }">
              {{ formatNumber(asset.change) }}%
            </span>
          </div>
          <div class="detail-row">
            <span class="label">24h High:</span>
            <span class="value">${{ formatNumber(asset.high24h) }}</span>
          </div>
          <div class="detail-row">
            <span class="label">24h Low:</span>
            <span class="value">${{ formatNumber(asset.low24h) }}</span>
          </div>
          <div class="detail-row">
            <span class="label">24h Volume:</span>
            <span class="value">${{ formatNumber(asset.volume24h) }}</span>
          </div>
        </div>

        <div class="price-chart">
          <LineChart
            :chart-data="getChartData(asset)"
            :options="chartOptions"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Line as LineChart } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

export default {
  name: 'Portfolio',
  components: {
    LineChart
  },
  data() {
    return {
      portfolio: {
        assets: [],
        totalValue: 0,
        portfolioChange: 0,
        volume24h: 0
      },
      chartOptions: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            display: false
          }
        },
        scales: {
          x: {
            display: false
          },
          y: {
            display: false
          }
        }
      }
    }
  },
  methods: {
    async fetchPortfolio() {
      try {
        const response = await fetch('/api/portfolio')
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        this.portfolio = await response.json()
      } catch (error) {
        console.error('Failed to fetch portfolio:', error)
      }
    },
    formatNumber(value) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(value)
    },
    getChartData(asset) {
      const labels = asset.priceHistory.map(point => 
        new Date(point.timestamp).toLocaleDateString()
      )
      const data = asset.priceHistory.map(point => point.price)

      return {
        labels,
        datasets: [
          {
            data,
            borderColor: asset.change >= 0 ? '#4CAF50' : '#F44336',
            borderWidth: 2,
            tension: 0.4,
            fill: false
          }
        ]
      }
    }
  },
  mounted() {
    this.fetchPortfolio()
    // Refresh portfolio data every minute
    setInterval(this.fetchPortfolio, 60000)
  }
}
</script>

<style scoped>
.portfolio-container {
  padding: 20px;
}

.portfolio-header {
  margin-bottom: 30px;
}

.portfolio-header h2 {
  margin: 0 0 15px 0;
  color: #333;
}

.portfolio-summary {
  display: flex;
  gap: 30px;
}

.total-value, .portfolio-change {
  display: flex;
  flex-direction: column;
}

.label {
  font-size: 14px;
  color: #666;
  margin-bottom: 5px;
}

.value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
}

.positive {
  color: #4CAF50;
}

.negative {
  color: #F44336;
}

.assets-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.asset-card {
  background: white;
  border-radius: 10px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.asset-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.asset-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.asset-value {
  font-size: 20px;
  font-weight: bold;
  color: #333;
}

.asset-details {
  margin-bottom: 20px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.detail-row .label {
  font-size: 14px;
  color: #666;
}

.detail-row .value {
  font-size: 14px;
  font-weight: 500;
}

.price-chart {
  height: 100px;
  margin-top: 15px;
}
</style> 