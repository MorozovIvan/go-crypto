<template>
  <div class="p-6 space-y-8">
    <h2 class="text-2xl font-bold mb-4">Market Analysis Dashboard</h2>
    <div v-if="error" class="mb-4 p-4 bg-red-100 text-red-700 rounded">
      <strong>Error:</strong> {{ error }}
    </div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <MetricCard
        v-for="metric in metrics"
        :key="metric.key"
        :title="metric.title"
        :value="metric.value"
        :indicator="metric.indicator"
        :chart-data="metric.chartData"
        :chart-labels="metric.chartLabels"
        :error="metric.error"
      />
    </div>
    <div class="mt-8 p-6 bg-white rounded-lg shadow">
      <h3 class="text-xl font-semibold mb-2">Algorithmic Signal</h3>
      <div class="flex items-center space-x-4 mb-2">
        <span class="text-lg font-bold">Signal:</span>
        <span :class="signalClass">{{ signal }}</span>
        <span v-if="asset" class="ml-2 text-gray-600">({{ asset }})</span>
      </div>
      <div class="text-gray-700">Weighted Score: <span class="font-mono">{{ totalScore.toFixed(2) }}</span></div>
      <div class="mt-4">
        <h4 class="font-semibold mb-1">Score Breakdown:</h4>
        <ul class="text-sm text-gray-600 grid grid-cols-2 md:grid-cols-3 gap-2">
          <li v-for="metric in metrics" :key="metric.key">
            <span class="font-bold">{{ metric.title }}:</span> {{ metric.error ? 'Error' : metric.score }} (w: {{ metric.weight * 100 }}%)
          </li>
        </ul>
      </div>
    </div>
    <div class="metric-card">
      <h3>RSI (Relative Strength Index)</h3>
      <div v-if="getMetricByKey('rsi').loading" class="loading">Loading...</div>
      <div v-else-if="getMetricByKey('rsi').error" class="error">{{ getMetricByKey('rsi').error }}</div>
      <div v-else>
        <div class="metric-value">{{ getMetricByKey('rsi').value ? getMetricByKey('rsi').value.toFixed(2) : 'N/A' }}</div>
        <div class="metric-indicator" :class="getRSIIndicator(getMetricByKey('rsi').value)">
          {{ getRSIIndicatorText(getMetricByKey('rsi').value) }}
        </div>
        <div class="metric-chart">
          <LineChart
            :data="getMetricByKey('rsi').historical || []"
            :labels="['5d', '4d', '3d', '2d', 'Now']"
            :color="getRSIColor(getMetricByKey('rsi').value)"
          />
        </div>
      </div>
    </div>
    <div class="metric-card">
      <h3>Google Trends</h3>
      <div v-if="getMetricByKey('google-trends').loading" class="loading">Loading...</div>
      <div v-else-if="getMetricByKey('google-trends').error" class="error">{{ getMetricByKey('google-trends').error }}</div>
      <div v-else>
        <div class="metric-value">{{ getMetricByKey('google-trends').value ? getMetricByKey('google-trends').value.toFixed(2) : 'N/A' }}</div>
        <div class="metric-indicator" :class="getTrendsIndicator(getMetricByKey('google-trends').value)">
          {{ getTrendsIndicatorText(getMetricByKey('google-trends').value) }}
        </div>
        <div class="metric-chart">
          <LineChart
            :data="getMetricByKey('google-trends').historical || []"
            :labels="['5d', '4d', '3d', '2d', 'Now']"
            :color="getTrendsColor(getMetricByKey('google-trends').value)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import { Line } from 'vue-chartjs';
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  LineElement,
  CategoryScale,
  LinearScale,
  PointElement,
  Filler
} from 'chart.js';
import MetricCard from './MetricCard.vue';

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement, Filler);

export default defineComponent({
  name: 'MarketAnalysis',
  components: { MetricCard },
  data() {
    console.log('MarketAnalysis: data() called');
    return {
      metrics: [
        {
          key: 'fear-greed',
          title: 'Fear & Greed Index',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.18,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'altcoin-season',
          title: 'Altcoin Season Index',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.12,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'btc-dominance',
          title: 'BTC Dominance',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.10,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'ssr',
          title: 'Stablecoin Supply Ratio',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.08,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'rsi',
          title: 'RSI',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.08,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true,
          historical: []
        },
        {
          key: 'market-cap',
          title: 'Total Market Cap Change (%)',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.08,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'google-trends',
          title: 'Google Trends',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.05,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true,
          historical: []
        },
        {
          key: 'ma-signal',
          title: 'Moving Averages',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.05,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'volume-trend',
          title: 'Volume Trend',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.05,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'exchange-flows',
          title: 'Exchange Flows',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.05,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'active-addresses',
          title: 'Active Addresses',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.04,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'whale-transactions',
          title: 'Whale Transactions',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.04,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'bollinger-bands',
          title: 'Bollinger Bands Width',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'funding-rate',
          title: 'Funding Rate',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'open-interest',
          title: 'Open Interest',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'eth-btc-ratio',
          title: 'ETH/BTC Ratio',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.02,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        }
      ],
      totalScore: 0,
      signal: 'Hold',
      asset: null,
      error: null
    };
  },
  created() {
    console.log('MarketAnalysis: created() hook called');
  },
  mounted() {
    console.log('MarketAnalysis: mounted() hook called');
    this.fetchAllMetrics();
  },
  beforeUnmount() {
    console.log('MarketAnalysis: beforeUnmount() hook called');
  },
  computed: {
    signalClass() {
      return {
        'text-green-600': this.signal === 'Buy',
        'text-red-600': this.signal === 'Sell',
        'text-yellow-600': this.signal === 'Hold'
      };
    }
  },
  methods: {
    formatWaitTime(seconds) {
      if (seconds < 60) {
        return `${seconds} seconds`;
      } else if (seconds < 3600) {
        return `${Math.ceil(seconds / 60)} minutes`;
      } else {
        return `${Math.ceil(seconds / 3600)} hours`;
      }
    },
    isRateLimitError(error) {
      return error && (
        error.includes('rate limited') ||
        error.includes('429') ||
        error.includes('Too Many Requests') ||
        error.includes('FLOOD_WAIT')
      )
    },
    isAuthError(error) {
      return error.includes('401') || error.includes('Unauthorized');
    },
    isNotFoundError(error) {
      return error.includes('404') || error.includes('Not Found');
    },
    getErrorMessage(error) {
      if (this.isRateLimitError(error)) {
        return 'API rate limit reached. Please try again in a few minutes.'
      }
      return error || 'An error occurred while fetching data'
    },
    async fetchMetric(key) {
      console.log(`MarketAnalysis: fetchMetric(${key}) called`);
      try {
        // Map of special endpoints that don't follow the standard pattern
        const endpointMap = {
          'fear-greed': '/api/fear-greed',
          'btc-dominance': '/api/btc-dominance',
          'market-cap': '/api/market-cap',
          'eth-btc-ratio': '/api/eth-btc-ratio',
          'altcoin-season': '/api/altcoin-season',
          'volume-trend': '/api/volume-trend',
          'bollinger-bands': '/api/bollinger-bands',
          'ssr': '/api/ssr',
          'exchange-flows': '/api/exchange-flows',
          'active-addresses': '/api/active-addresses',
          'whale-transactions': '/api/whale-transactions',
          'funding-rate': '/api/funding-rate',
          'open-interest': '/api/open-interest',
          'ma-signal': '/api/moving-averages',
          'google-trends': '/api/google-trends'
        };

        const endpoint = endpointMap[key] || `/api/${key}`;
        console.log(`MarketAnalysis: Using endpoint ${endpoint} for ${key}`);
        
        const response = await fetch(`http://localhost:8080${endpoint}`);
        console.log(`MarketAnalysis: fetchMetric(${key}) response received:`, response.status);
        
        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
        }
        
        const data = await response.json();
        console.log(`MarketAnalysis: fetchMetric(${key}) data:`, data);
        
        const metricIndex = this.metrics.findIndex(m => m.key === key);
        if (metricIndex === -1) {
          console.error(`MarketAnalysis: Metric ${key} not found in metrics array`);
          return;
        }
        
        if (data.error) {
          throw new Error(data.error);
        }

        // Special handling for SSR
        if (key === 'ssr') {
          const ssrValue = data.current_ssr;
          this.metrics[metricIndex].value = (ssrValue * 100).toFixed(2) + '%';
          this.metrics[metricIndex].chartData = data.historical.map(v => v * 100);
          this.metrics[metricIndex].chartLabels = data.labels;
          
          // Set indicator based on SSR value
          if (ssrValue < 0.5) {
            this.metrics[metricIndex].indicator = 'Buy';
            this.metrics[metricIndex].score = 1;
          } else if (ssrValue > 2.0) {
            this.metrics[metricIndex].indicator = 'Sell';
            this.metrics[metricIndex].score = -1;
          } else {
            this.metrics[metricIndex].indicator = 'Hold';
            this.metrics[metricIndex].score = 0;
          }
        } else {
          // Handle other metrics as before
          this.metrics[metricIndex].value = data.value;
          this.metrics[metricIndex].indicator = data.indicator;
          this.metrics[metricIndex].score = data.score;
          this.metrics[metricIndex].chartData = data.chart_data || [];
          this.metrics[metricIndex].chartLabels = data.chart_labels || [];
        }
        
        this.metrics[metricIndex].error = false;
        this.metrics[metricIndex].loading = false;
        this.calculateSignal();
      } catch (error) {
        console.error(`Error fetching ${key}:`, error);
        const metricIndex = this.metrics.findIndex(m => m.key === key);
        if (metricIndex !== -1) {
          this.metrics[metricIndex].error = true;
          this.metrics[metricIndex].value = this.getErrorMessage(error.message);
          this.metrics[metricIndex].loading = false;
        }
      }
    },
    async fetchExchangeFlows() {
      const idx = this.metrics.findIndex(m => m.key === 'exchange-flows');
      try {
        const res = await fetch('/api/exchange-flows');
        const data = await res.json();
        if (data.error) {
          this.metrics[idx].error = true;
          this.metrics[idx].value = this.getErrorMessage(data.error);
          return;
        }
        if (data && data.netFlow !== undefined) {
          this.metrics[idx].value = data.netFlow.toFixed(2);
          this.metrics[idx].chartData = data.history || [];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (data.netFlow < -1000) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (data.netFlow > 1000) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
          this.metrics[idx].value = 'RT';
        }
      } catch (e) {
        this.metrics[idx].error = true;
        this.metrics[idx].value = this.getErrorMessage(e.message);
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchVolumeTrend() {
      const idx = this.metrics.findIndex(m => m.key === 'volume-trend');
      try {
        const res = await fetch('/api/volume-trend');
        const data = await res.json();
        if (data.error) {
          this.metrics[idx].error = true;
          this.metrics[idx].value = this.getErrorMessage(data.error);
          return;
        }
        if (data && data.value !== undefined) {
          // Format the trend value as a percentage
          this.metrics[idx].value = `${(data.value * 100).toFixed(1)}%`;
          this.metrics[idx].chartData = data.chart_data || [];
          this.metrics[idx].chartLabels = data.chart_labels || ['5d', '4d', '3d', '2d', 'Now'];
          
          // Set indicator and score based on trend value
          if (data.value > 0.1) { // More than 10% increase
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (data.value < -0.1) { // More than 10% decrease
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
          this.metrics[idx].value = 'No data';
        }
      } catch (e) {
        this.metrics[idx].error = true;
        this.metrics[idx].value = this.getErrorMessage(e.message);
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchBollingerBands() {
      const idx = this.metrics.findIndex(m => m.key === 'bollinger-bands');
      try {
        const res = await fetch('/api/bollinger-bands');
        const data = await res.json();
        if (data.error) {
          this.metrics[idx].error = true;
          this.metrics[idx].value = this.getErrorMessage(data.error);
          return;
        }
        if (data && data.width !== undefined && data.bandwidth) {
          this.metrics[idx].value = data.width.toFixed(4);
          this.metrics[idx].chartData = data.bandwidth;
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (data.width < 0.02) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (data.width > 0.04) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
          this.metrics[idx].value = 'RT';
        }
      } catch (e) {
        this.metrics[idx].error = true;
        this.metrics[idx].value = this.getErrorMessage(e.message);
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchRSI() {
      try {
        this.metrics.rsi.loading = true;
        this.metrics.rsi.error = false;
        
        const response = await fetch('/api/rsi');
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to fetch RSI');
        }
        
        const data = await response.json();
        
        this.metrics.rsi = {
          ...this.metrics.rsi,
          value: data.current,
          historical: data.historical,
          loading: false,
          error: false
        };
      } catch (error) {
        console.error('Error fetching RSI:', error);
        this.metrics.rsi = {
          ...this.metrics.rsi,
          value: null,
          historical: [],
          loading: false,
          error: true
        };
      }
    },
    async fetchMovingAverages() {
      try {
        this.metrics.movingAverages.loading = true;
        this.metrics.movingAverages.error = false;
        
        const response = await fetch('/api/moving-averages');
        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Failed to fetch moving averages');
        }
        
        const data = await response.json();
        
        this.metrics.movingAverages = {
          ...this.metrics.movingAverages,
          value: data.value,
          indicator: data.indicator,
          score: data.score,
          chartData: data.chart_data || [],
          chartLabels: data.chart_labels || [],
          loading: false,
          error: false
        };
      } catch (error) {
        console.error('Error fetching moving averages:', error);
        this.metrics.movingAverages = {
          ...this.metrics.movingAverages,
          value: null,
          indicator: 'Hold',
          score: 0,
          chartData: [],
          chartLabels: [],
          loading: false,
          error: true
        };
      }
    },
    calculateSignal() {
      console.log('[calculateSignal] Starting signal calculation')
      const criticalKeys = [
        'fear-greed', 'altcoin-season', 'btc-dominance', 'ssr', 'rsi', 'market-cap', 'ma-signal', 'volume-trend'
      ];
      let error = null;
      for (const key of criticalKeys) {
        const m = this.metrics.find(m => m.key === key);
        if (!m || m.error) {
          console.log(`[calculateSignal] Critical metric ${key} is unavailable or has error`)
          error = 'One or more critical metrics are unavailable. Algorithmic signal may be inaccurate.';
          break;
        }
      }
      this.error = error;

      let totalWeight = 0;
      let weightedScore = 0;
      
      for (const m of this.metrics) {
        if (!m.error) {
          weightedScore += m.score * m.weight;
          totalWeight += m.weight;
        }
      }

      const totalScore = totalWeight > 0 ? weightedScore / totalWeight : 0;
      console.log(`[calculateSignal] Calculated total score: ${totalScore}`)
      
      let signal = 'Hold';
      let asset = null;
      
      if (!error && totalWeight > 0) {
        if (totalScore > 0.5) {
          signal = 'Buy';
          const altcoinMetric = this.metrics.find(m => m.key === 'altcoin-season');
          asset = altcoinMetric && !altcoinMetric.error && altcoinMetric.value > 75 ? 'Altcoins' : 'Bitcoin';
        } else if (totalScore < -0.5) {
          signal = 'Sell';
          asset = 'All';
        }
      }

      console.log(`[calculateSignal] Final signal: ${signal}, asset: ${asset}`)
      this.totalScore = totalScore;
      this.signal = signal;
      this.asset = asset;
    },
    async fetchAllMetrics() {
      console.log('MarketAnalysis: fetchAllMetrics() called');
      try {
        const promises = this.metrics.map(metric => this.fetchMetric(metric.key));
        await Promise.all(promises);
        console.log('MarketAnalysis: All metrics fetched successfully');
      } catch (error) {
        console.error('MarketAnalysis: Error fetching metrics:', error);
        this.error = 'Failed to fetch market data';
      }
    },
    mounted() {
      console.log('[mounted] MarketAnalysis component mounted')
      console.log('[mounted] Initial metrics state:', this.metrics.map(m => ({ 
        key: m.key, 
        value: m.value,
        loading: m.loading,
        error: m.error 
      })))
      this.fetchAllMetrics()
      // Refresh data every 5 minutes
      console.log('[mounted] Setting up refresh interval')
      setInterval(() => {
        console.log('[refresh] Starting scheduled refresh...')
        this.fetchAllMetrics()
      }, 5 * 60 * 1000)
    },
    getRSIIndicator(value) {
      if (!value) return 'neutral';
      if (value >= 70) return 'overbought';
      if (value <= 30) return 'oversold';
      return 'neutral';
    },
    getRSIIndicatorText(value) {
      if (!value) return 'N/A';
      if (value >= 70) return 'Overbought';
      if (value <= 30) return 'Oversold';
      return 'Neutral';
    },
    getRSIColor(value) {
      if (!value) return '#3b82f6';
      if (value >= 70) return '#ef4444';
      if (value <= 30) return '#22c55e';
      return '#3b82f6';
    },
    getTrendsIndicator(value) {
      if (!value) return 'medium';
      if (value >= 70) return 'high';
      if (value <= 30) return 'low';
      return 'medium';
    },
    getTrendsIndicatorText(value) {
      if (!value) return 'N/A';
      if (value >= 70) return 'High Interest';
      if (value <= 30) return 'Low Interest';
      return 'Moderate Interest';
    },
    getTrendsColor(value) {
      if (!value) return '#3b82f6';
      if (value >= 70) return '#ef4444';
      if (value <= 30) return '#22c55e';
      return '#3b82f6';
    },
    getMetricByKey(key) {
      return this.metrics.find(m => m.key === key) || {
        loading: true,
        error: null,
        value: null,
        historical: []
      };
    },
    async fetchGoogleTrends() {
      const idx = this.metrics.findIndex(m => m.key === 'google-trends');
      try {
        const res = await fetch('/api/google-trends');
        const data = await res.json();
        
        // Handle error response
        if (data.error) {
          console.error('Google Trends error:', data.error);
          this.metrics[idx].error = true;
          this.metrics[idx].value = data.error;
          this.metrics[idx].chartData = [];
          this.metrics[idx].chartLabels = [];
          this.metrics[idx].indicator = 'Hold';
          this.metrics[idx].score = 0;
          return;
        }

        // Validate data
        if (!data || data.value === undefined) {
          console.error('Invalid Google Trends data:', data);
          this.metrics[idx].error = true;
          this.metrics[idx].value = 'Invalid data received';
          this.metrics[idx].chartData = [];
          this.metrics[idx].chartLabels = [];
          this.metrics[idx].indicator = 'Hold';
          this.metrics[idx].score = 0;
          return;
        }

        // Validate value is within expected range (0-100)
        const value = Math.max(0, Math.min(100, parseFloat(data.value)));
        this.metrics[idx].value = value.toFixed(1);
        this.metrics[idx].chartData = data.historical || [];
        this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
        
        // Calculate average value for comparison
        const avgValue = this.metrics[idx].chartData.reduce((a, b) => a + b, 0) / this.metrics[idx].chartData.length;
        
        // Set indicator and score based on current value vs average
        if (value < avgValue * 0.75) {
          this.metrics[idx].indicator = 'Buy';
          this.metrics[idx].score = 1;
        } else if (value > avgValue * 1.25) {
          this.metrics[idx].indicator = 'Sell';
          this.metrics[idx].score = -1;
        } else {
          this.metrics[idx].indicator = 'Hold';
          this.metrics[idx].score = 0;
        }
        this.metrics[idx].error = false;
      } catch (e) {
        console.error('Error fetching Google Trends:', e);
        this.metrics[idx].error = true;
        this.metrics[idx].value = `Error: ${e.message}`;
        this.metrics[idx].chartData = [];
        this.metrics[idx].chartLabels = [];
        this.metrics[idx].indicator = 'Hold';
        this.metrics[idx].score = 0;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    }
  }
});
</script>

<style scoped>
.metric-indicator.overbought {
  color: #ef4444;
}

.metric-indicator.oversold {
  color: #22c55e;
}

.metric-indicator.neutral {
  color: #3b82f6;
}

.metric-indicator.high {
  color: #ef4444;
}

.metric-indicator.low {
  color: #22c55e;
}

.metric-indicator.medium {
  color: #3b82f6;
}
</style> 