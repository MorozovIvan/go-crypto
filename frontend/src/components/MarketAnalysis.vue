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
          loading: true
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
          key: 'trends',
          title: 'Google Trends',
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
          'ma-signal': '/api/moving-averages'
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
        if (data && data.trend !== undefined) {
          this.metrics[idx].value = data.trend;
          this.metrics[idx].chartData = data.history || [0.5, 0.6, 0.7, 0.8, 1];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (data.trend === 'High Rising') {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (data.trend === 'Low Falling') {
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
          value: data.value,
          indicator: data.indicator,
          score: data.score,
          chartData: data.chart_data || [],
          chartLabels: data.chart_labels || [],
          loading: false,
          error: false
        };
      } catch (error) {
        console.error('Error fetching RSI:', error);
        this.metrics.rsi = {
          ...this.metrics.rsi,
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
    }
  }
});
</script> 