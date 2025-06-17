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
      <div class="flex items-center space-x-4 mb-2">
        <span class="text-sm font-semibold">Confidence:</span>
        <span :class="{
          'text-green-600': confidence === 'High',
          'text-yellow-600': confidence === 'Medium', 
          'text-red-600': confidence === 'Low'
        }">{{ confidence }}</span>
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
          weight: 0.15,
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
          weight: 0.10,
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
          weight: 0.12,
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
          weight: 0.12,
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
          weight: 0.10,
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
          weight: 0.04,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true,
          historical: []
        },
        {
          key: 'moving-averages',
          title: 'Moving Averages',
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
          key: 'volume-trend',
          title: 'Volume Trend',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.06,
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
          weight: 0.06,
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
        },
        // Advanced DeFi & Crypto Metrics
        {
          key: 'defi-tvl',
          title: 'DeFi TVL',
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
          key: 'social-sentiment',
          title: 'Social Sentiment',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.06,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'options-flow',
          title: 'Options Flow (Put/Call)',
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
          key: 'stablecoin-flows',
          title: 'Stablecoin Flows',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.07,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'network-health',
          title: 'Network Health',
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
          key: 'institutional-flows',
          title: 'Institutional Flows',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.09,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        },
        {
          key: 'yield-curves',
          title: 'DeFi Yield Premium',
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
          key: 'correlation-matrix',
          title: 'BTC-Stock Correlation',
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
          key: 'volatility-surface',
          title: 'Implied Volatility',
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
          key: 'liquidation-heatmap',
          title: 'Liquidation Risk',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.06,
          chartData: [],
          chartLabels: [],
          error: false,
          loading: true
        }
      ],
      totalScore: 0,
      signal: 'Hold',
      asset: null,
      error: null,
      confidence: 'Low',
      websocket: null,
      wsConnected: false,
      wsReconnectAttempts: 0,
      maxReconnectAttempts: 5,
      reconnectInterval: 5000,
      lastUpdateTime: null
    };
  },
  created() {
    console.log('MarketAnalysis: created() hook called');
  },
  mounted() {
    console.log('[mounted] MarketAnalysis component mounted')
    console.log('[mounted] Initial metrics state:', this.metrics.map(m => ({ 
      key: m.key, 
      value: m.value,
      loading: m.loading,
      error: m.error 
    })))
    
    // Initial fetch
    this.fetchAllMetrics()
    
    // Initialize WebSocket connection
    this.initWebSocket()
    
    // Fallback polling for non-critical metrics (reduced frequency)
    this.setupFallbackPolling()
    
    // Performance monitoring
    this.startPerformanceMonitoring()
  },
  beforeUnmount() {
    console.log('MarketAnalysis: beforeUnmount() hook called');
    this.closeWebSocket()
  },
  computed: {
    signalClass() {
      return {
        'text-green-600 font-bold': this.signal === 'Strong Buy',
        'text-green-500': this.signal === 'Buy',
        'text-red-600 font-bold': this.signal === 'Strong Sell',
        'text-red-500': this.signal === 'Sell',
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
      const idx = this.metrics.findIndex(m => m.key === key);
      if (idx === -1) return;
      
      this.metrics[idx].loading = true;
      this.metrics[idx].error = false;
      
      const maxRetries = 3;
      let retryCount = 0;
      
      while (retryCount < maxRetries) {
        try {
          console.log(`[fetchMetric] Fetching ${key} (attempt ${retryCount + 1}/${maxRetries})`);
          const response = await fetch(`/api/${key}`, {
            timeout: 10000, // 10 second timeout
            headers: {
              'Cache-Control': 'no-cache',
              'Pragma': 'no-cache'
            }
          });
          
          if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
          }
          
          const data = await response.json();
          
          // Validate data structure
          if (!data || (data.value === undefined && data.netFlow === undefined)) {
            throw new Error('Invalid data structure received');
          }
          
          // Handle different response formats
          let value = data.value !== undefined ? data.value : data.netFlow;
          let indicator = data.indicator || 'Hold';
          let score = data.score !== undefined ? data.score : 0;
          
          // Additional validation for specific metrics
          if (key === 'fear-greed' && (value < 0 || value > 100)) {
            throw new Error(`Invalid Fear & Greed value: ${value}`);
          }
          
          if (key === 'rsi' && (value < 0 || value > 100)) {
            throw new Error(`Invalid RSI value: ${value}`);
          }
          
          this.metrics[idx].value = value;
          this.metrics[idx].indicator = indicator;
          this.metrics[idx].score = score;
          this.metrics[idx].chartData = data.chart_data || data.historical || [];
          this.metrics[idx].chartLabels = data.chart_labels || data.labels || [];
          this.metrics[idx].loading = false;
          this.metrics[idx].error = false;
          this.metrics[idx].lastUpdated = new Date().toISOString();
          
          console.log(`[fetchMetric] Successfully fetched ${key}:`, {
            value: value,
            indicator: indicator,
            score: score
          });
          
          // Success - break retry loop
          break;
          
        } catch (error) {
          retryCount++;
          console.error(`[fetchMetric] Error fetching ${key} (attempt ${retryCount}):`, error.message);
          
          if (retryCount >= maxRetries) {
            // Final failure - set error state
            this.metrics[idx].loading = false;
            this.metrics[idx].error = true;
            this.metrics[idx].value = `Error: ${error.message}`;
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
            this.metrics[idx].chartData = [];
            this.metrics[idx].chartLabels = [];
            console.error(`[fetchMetric] Failed to fetch ${key} after ${maxRetries} attempts`);
          } else {
            // Wait before retry (exponential backoff)
            const delay = Math.pow(2, retryCount) * 1000; // 2s, 4s, 8s
            console.log(`[fetchMetric] Retrying ${key} in ${delay}ms...`);
            await new Promise(resolve => setTimeout(resolve, delay));
          }
        }
      }
      
      // Recalculate signal after each metric update
      this.calculateSignal();
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
        'fear-greed', 'altcoin-season', 'btc-dominance', 'ssr', 'rsi', 'market-cap', 'moving-averages', 'volume-trend'
      ];
      let error = null;
      let criticalMetricsAvailable = 0;
      
      for (const key of criticalKeys) {
        const m = this.metrics.find(m => m.key === key);
        if (!m || m.error) {
          console.log(`[calculateSignal] Critical metric ${key} is unavailable or has error`)
          error = 'One or more critical metrics are unavailable. Algorithmic signal may be inaccurate.';
        } else {
          criticalMetricsAvailable++;
        }
      }
      this.error = error;

      let totalWeight = 0;
      let weightedScore = 0;
      let bullishCount = 0;
      let bearishCount = 0;
      let neutralCount = 0;
      let strongSignals = 0; // Count of signals with absolute score >= 0.8
      let availableMetrics = 0;
      
      for (const m of this.metrics) {
        if (!m.error) {
          weightedScore += m.score * m.weight;
          totalWeight += m.weight;
          availableMetrics++;
          
          // Count signal types for consensus analysis
          if (m.score > 0.1) {
            bullishCount++;
            if (Math.abs(m.score) >= 0.8) strongSignals++;
          } else if (m.score < -0.1) {
            bearishCount++;
            if (Math.abs(m.score) >= 0.8) strongSignals++;
          } else {
            neutralCount++;
          }
        }
      }

      const totalScore = totalWeight > 0 ? weightedScore / totalWeight : 0;
      const signalStrength = Math.abs(totalScore);
      const totalSignals = bullishCount + bearishCount + neutralCount;
      
      // Enhanced confidence calculation
      let confidenceScore = 0;
      
      // 1. Signal strength (40% of confidence)
      confidenceScore += Math.min(signalStrength * 2.5, 1.0) * 0.4;
      
      // 2. Signal consensus (30% of confidence)
      const dominantSignals = Math.max(bullishCount, bearishCount);
      const consensusRatio = totalSignals > 0 ? dominantSignals / totalSignals : 0;
      confidenceScore += consensusRatio * 0.3;
      
      // 3. Data quality (20% of confidence)
      const dataQuality = availableMetrics / this.metrics.length;
      const criticalDataQuality = criticalMetricsAvailable / criticalKeys.length;
      confidenceScore += (dataQuality * 0.5 + criticalDataQuality * 0.5) * 0.2;
      
      // 4. Strong signal presence (10% of confidence)
      const strongSignalRatio = totalSignals > 0 ? strongSignals / totalSignals : 0;
      confidenceScore += strongSignalRatio * 0.1;
      
      console.log(`[calculateSignal] Confidence components:`, {
        signalStrength: signalStrength,
        consensusRatio: consensusRatio,
        dataQuality: dataQuality,
        strongSignalRatio: strongSignalRatio,
        finalConfidenceScore: confidenceScore
      });
      
      let signal = 'Hold';
      let asset = null;
      let confidence = 'Low';
      
      // Determine confidence level
      if (confidenceScore >= 0.75) {
        confidence = 'High';
      } else if (confidenceScore >= 0.5) {
        confidence = 'Medium';
      } else {
        confidence = 'Low';
      }
      
      if (!error && totalWeight > 0) {
        // More nuanced thresholds based on signal strength and consensus
        const strongThreshold = 0.4;
        const moderateThreshold = 0.25;
        
        if (totalScore > strongThreshold && bullishCount >= 3 && consensusRatio >= 0.6) {
          signal = 'Strong Buy';
          const altcoinMetric = this.metrics.find(m => m.key === 'altcoin-season');
          const btcDominanceMetric = this.metrics.find(m => m.key === 'btc-dominance');
          
          // More sophisticated asset allocation
          if (altcoinMetric && !altcoinMetric.error && altcoinMetric.value > 75) {
            asset = 'Altcoins';
          } else if (btcDominanceMetric && !btcDominanceMetric.error && btcDominanceMetric.value < 50) {
            asset = 'Mixed (BTC + Altcoins)';
          } else {
            asset = 'Bitcoin';
          }
        } else if (totalScore > moderateThreshold && bullishCount >= 2) {
          signal = 'Buy';
          asset = 'Bitcoin'; // Conservative allocation for moderate signals
        } else if (totalScore < -strongThreshold && bearishCount >= 3 && consensusRatio >= 0.6) {
          signal = 'Strong Sell';
          asset = 'All';
        } else if (totalScore < -moderateThreshold && bearishCount >= 2) {
          signal = 'Sell';
          asset = 'All';
        } else {
          signal = 'Hold';
          asset = null;
          // For hold signals, confidence depends more on neutrality consensus
          if (neutralCount >= totalSignals * 0.5) {
            confidence = Math.max(confidence, 'Medium');
          }
        }
      }

      console.log(`[calculateSignal] Final signal: ${signal}, asset: ${asset}, confidence: ${confidence}`)
      console.log(`[calculateSignal] Signal breakdown: Bull: ${bullishCount}, Bear: ${bearishCount}, Neutral: ${neutralCount}`)
      
      this.totalScore = totalScore;
      this.signal = signal;
      this.asset = asset;
      this.confidence = confidence;
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
    startPerformanceMonitoring() {
      setInterval(() => {
        const now = new Date()
        const staleMetrics = this.metrics.filter(m => {
          if (!m.lastUpdated) return true
          const lastUpdate = new Date(m.lastUpdated)
          const ageMinutes = (now - lastUpdate) / (1000 * 60)
          return ageMinutes > 15 // Consider stale if older than 15 minutes
        })
        
        if (staleMetrics.length > 0) {
          console.warn('[performance] Stale metrics detected:', staleMetrics.map(m => m.key))
          // Trigger refresh of stale metrics
          staleMetrics.forEach(m => this.fetchMetric(m.key))
        }
        
        const errorMetrics = this.metrics.filter(m => m.error)
        if (errorMetrics.length > 0) {
          console.warn('[performance] Metrics with errors:', errorMetrics.map(m => m.key))
        }
        
        // WebSocket health check
        const wsHealthy = this.wsConnected && this.lastUpdateTime && 
                         (now - this.lastUpdateTime) < 2 * 60 * 1000 // Last update within 2 minutes
        
        console.log('[performance] System status:', {
          totalMetrics: this.metrics.length,
          errorCount: errorMetrics.length,
          staleCount: staleMetrics.length,
          wsConnected: this.wsConnected,
          wsHealthy: wsHealthy,
          lastSignalUpdate: this.totalScore,
          lastWSUpdate: this.lastUpdateTime
        })
        
        // Reconnect WebSocket if unhealthy
        if (!wsHealthy && this.wsReconnectAttempts < this.maxReconnectAttempts) {
          console.warn('[performance] WebSocket appears unhealthy, attempting reconnect')
          this.initWebSocket()
        }
      }, 5 * 60 * 1000) // Check every 5 minutes
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
    },
    initWebSocket() {
      if (this.websocket) {
        this.closeWebSocket()
      }
      
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${protocol}//${window.location.host}/api/ws`
      
      console.log('[WebSocket] Connecting to:', wsUrl)
      
      try {
        this.websocket = new WebSocket(wsUrl)
        
        this.websocket.onopen = this.onWebSocketOpen
        this.websocket.onmessage = this.onWebSocketMessage
        this.websocket.onclose = this.onWebSocketClose
        this.websocket.onerror = this.onWebSocketError
        
      } catch (error) {
        console.error('[WebSocket] Connection failed:', error)
        this.scheduleReconnect()
      }
    },
    
    onWebSocketOpen(event) {
      console.log('[WebSocket] Connected successfully')
      this.wsConnected = true
      this.wsReconnectAttempts = 0
      this.lastUpdateTime = new Date()
    },
    
    onWebSocketMessage(event) {
      try {
        const message = JSON.parse(event.data)
        console.log('[WebSocket] Received message:', message)
        
        if (message.type === 'market_update') {
          this.updateMetricFromWebSocket(message)
        }
        
        this.lastUpdateTime = new Date()
        
      } catch (error) {
        console.error('[WebSocket] Failed to parse message:', error)
      }
    },
    
    onWebSocketClose(event) {
      console.log('[WebSocket] Connection closed:', event.code, event.reason)
      this.wsConnected = false
      
      if (event.code !== 1000) { // Not a normal closure
        this.scheduleReconnect()
      }
    },
    
    onWebSocketError(error) {
      console.error('[WebSocket] Error:', error)
      this.wsConnected = false
    },
    
    scheduleReconnect() {
      if (this.wsReconnectAttempts >= this.maxReconnectAttempts) {
        console.error('[WebSocket] Max reconnection attempts reached')
        return
      }
      
      this.wsReconnectAttempts++
      const delay = this.reconnectInterval * Math.pow(2, this.wsReconnectAttempts - 1)
      
      console.log(`[WebSocket] Scheduling reconnect attempt ${this.wsReconnectAttempts} in ${delay}ms`)
      
      setTimeout(() => {
        if (!this.wsConnected) {
          this.initWebSocket()
        }
      }, delay)
    },
    
    closeWebSocket() {
      if (this.websocket) {
        this.websocket.close(1000, 'Component unmounting')
        this.websocket = null
        this.wsConnected = false
      }
    },
    
    updateMetricFromWebSocket(message) {
      const idx = this.metrics.findIndex(m => m.key === message.metric)
      if (idx === -1) return
      
      console.log(`[WebSocket] Updating ${message.metric}:`, message)
      
      this.metrics[idx].value = message.value
      this.metrics[idx].indicator = message.indicator
      this.metrics[idx].score = message.score
      this.metrics[idx].chartData = message.chart_data || []
      this.metrics[idx].loading = false
      this.metrics[idx].error = false
      this.metrics[idx].lastUpdated = message.timestamp
      
      // Recalculate signal after WebSocket update
      this.calculateSignal()
    },
    
    setupFallbackPolling() {
      // Reduced frequency polling for non-critical metrics only
      const nonCriticalMetrics = ['google-trends', 'whale-transactions', 'bollinger-bands']
      
      setInterval(() => {
        if (!this.wsConnected) {
          console.log('[Fallback] WebSocket disconnected, using polling for critical metrics')
          const criticalMetrics = ['fear-greed', 'btc-dominance', 'rsi', 'moving-averages']
          criticalMetrics.forEach(key => this.fetchMetric(key))
        }
        
        // Always poll non-critical metrics (they're not sent via WebSocket)
        nonCriticalMetrics.forEach(key => this.fetchMetric(key))
      }, 2 * 60 * 1000) // Every 2 minutes
    },
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