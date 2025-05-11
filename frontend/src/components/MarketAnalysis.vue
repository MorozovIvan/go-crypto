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
  PointElement
} from 'chart.js';
import MetricCard from './MetricCard.vue';

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement);

export default defineComponent({
  name: 'MarketAnalysis',
  components: { MetricCard },
  data() {
    // Initial state with loading/error for real metrics
    return {
      metrics: [
        {
          key: 'fear_greed',
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
          key: 'altcoin_index',
          title: 'Altcoin Season Index',
          value: 80,
          indicator: 'Buy',
          score: 1,
          weight: 0.12,
          chartData: [60, 65, 70, 75, 80],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'btc_dominance',
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
          value: 0.5,
          indicator: 'Buy',
          score: 1,
          weight: 0.08,
          chartData: [0.7, 0.65, 0.6, 0.55, 0.5],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'rsi',
          title: 'RSI',
          value: 28,
          indicator: 'Buy',
          score: 1,
          weight: 0.08,
          chartData: [35, 33, 31, 29, 28],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'market_cap',
          title: 'Total Market Cap Change (%)',
          value: 6,
          indicator: 'Buy',
          score: 1,
          weight: 0.08,
          chartData: [2, 3, 4, 5, 6],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'trends',
          title: 'Google Trends',
          value: 20,
          indicator: 'Buy',
          score: 1,
          weight: 0.05,
          chartData: [30, 28, 25, 22, 20],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'ma_signal',
          title: 'Moving Averages',
          value: 'Golden Cross',
          indicator: 'Buy',
          score: 1,
          weight: 0.05,
          chartData: [0, 0, 1, 1, 1],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        {
          key: 'volume',
          title: 'Volume Trend',
          value: 'High Rising',
          indicator: 'Buy',
          score: 1,
          weight: 0.05,
          chartData: [0.5, 0.6, 0.7, 0.8, 1],
          chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
          error: false
        },
        // New metrics below
        {
          key: 'exchange_flows',
          title: 'Exchange Flows',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.05,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'active_addresses',
          title: 'Active Addresses',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.04,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'whale_tx',
          title: 'Whale Transactions',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.04,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'bollinger',
          title: 'Bollinger Bands Width',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'funding_rate',
          title: 'Funding Rate',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'open_interest',
          title: 'Open Interest',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.03,
          chartData: [],
          chartLabels: [],
          error: true
        },
        {
          key: 'eth_btc_ratio',
          title: 'ETH/BTC Ratio',
          value: null,
          indicator: 'Hold',
          score: 0,
          weight: 0.02,
          chartData: [],
          chartLabels: [],
          error: true
        }
      ],
      totalScore: 0,
      signal: 'Hold',
      asset: null,
      error: null
    };
  },
  computed: {
    signalClass() {
      return {
        'text-green-600': this.signal === 'Buy',
        'text-red-600': this.signal === 'Sell',
        'text-gray-600': this.signal === 'Hold',
        'font-bold': true,
        'text-lg': true
      };
    }
  },
  mounted() {
    this.fetchFearGreed();
    this.fetchBTCDominance();
    this.fetchSSR();
    this.fetchMarketCapChange();
    this.fetchETHBTCRatio();
    this.fetchBTCPriceHistory();
    this.fetchGoogleTrends();
    this.calculateSignal();
  },
  methods: {
    async fetchFearGreed() {
      const idx = this.metrics.findIndex(m => m.key === 'fear_greed');
      try {
        const res = await fetch('https://api.alternative.me/fng/?limit=5');
        const data = await res.json();
        if (data && data.data && data.data.length > 0) {
          const values = data.data.map(d => parseInt(d.value));
          this.metrics[idx].value = values[0];
          this.metrics[idx].chartData = values.reverse();
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          // Scoring
          if (values[0] < 30) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (values[0] > 70) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchBTCDominance() {
      const idx = this.metrics.findIndex(m => m.key === 'btc_dominance');
      try {
        const res = await fetch('http://localhost:5002/api/cmc/global');
        const data = await res.json();
        if (data && data.data && data.data.btc_dominance) {
          const value = data.data.btc_dominance;
          this.metrics[idx].value = value.toFixed(2);
          this.metrics[idx].chartData = [value, value, value, value, value];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (value < 50) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (value > 60) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchSSR() {
      const idx = this.metrics.findIndex(m => m.key === 'ssr');
      try {
        // Get BTC and stablecoin market caps from CoinGecko
        const res = await fetch('https://api.coingecko.com/api/v3/global');
        const data = await res.json();
        if (data && data.data && data.data.total_market_cap && data.data.total_market_cap.btc && data.data.total_market_cap.usdt) {
          const btcCap = data.data.total_market_cap.btc;
          const stableCap = data.data.total_market_cap.usdt + (data.data.total_market_cap.usdc || 0) + (data.data.total_market_cap.dai || 0);
          const ssr = btcCap / stableCap;
          this.metrics[idx].value = ssr.toFixed(2);
          this.metrics[idx].chartData = [ssr, ssr, ssr, ssr, ssr];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          // Scoring (mock thresholds)
          if (ssr < 0.6) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (ssr > 1.2) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchMarketCapChange() {
      const idx = this.metrics.findIndex(m => m.key === 'market_cap');
      try {
        const res = await fetch('https://api.coingecko.com/api/v3/global');
        const data = await res.json();
        if (data && data.data && data.data.total_market_cap && data.data.total_market_cap.usd && data.data.market_cap_change_percentage_24h_usd) {
          const change = data.data.market_cap_change_percentage_24h_usd;
          this.metrics[idx].value = change.toFixed(2);
          this.metrics[idx].chartData = [change, change, change, change, change];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (change > 5) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (change < -5) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchETHBTCRatio() {
      const idx = this.metrics.findIndex(m => m.key === 'eth_btc_ratio');
      try {
        const res = await fetch('https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum&vs_currencies=usd');
        const data = await res.json();
        if (data && data.bitcoin && data.ethereum) {
          const ratio = data.ethereum.usd / data.bitcoin.usd;
          this.metrics[idx].value = ratio.toFixed(4);
          this.metrics[idx].chartData = [ratio, ratio, ratio, ratio, ratio];
          this.metrics[idx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          if (ratio > 0.06) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (ratio < 0.05) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    async fetchBTCPriceHistory() {
      // For RSI, MA, Volume, Bollinger Bands
      const rsiIdx = this.metrics.findIndex(m => m.key === 'rsi');
      const maIdx = this.metrics.findIndex(m => m.key === 'ma_signal');
      const volIdx = this.metrics.findIndex(m => m.key === 'volume');
      const bollIdx = this.metrics.findIndex(m => m.key === 'bollinger');
      try {
        const res = await fetch('https://api.coingecko.com/api/v3/coins/bitcoin/ohlc?vs_currency=usd&days=7');
        const data = await res.json();
        if (Array.isArray(data) && data.length > 0) {
          // data: [timestamp, open, high, low, close]
          const closes = data.map(d => d[4]);
          // RSI (14): use closes, simple calculation for demo
          const rsi = closes.length > 14 ? (100 - (100 / (1 + (closes.slice(-15, -1).reduce((a, b) => a + b, 0) / closes.slice(-14).reduce((a, b) => a + b, 0))))) : 50;
          this.metrics[rsiIdx].value = rsi.toFixed(2);
          this.metrics[rsiIdx].chartData = closes.slice(-5);
          this.metrics[rsiIdx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          this.metrics[rsiIdx].indicator = rsi < 30 ? 'Buy' : rsi > 70 ? 'Sell' : 'Hold';
          this.metrics[rsiIdx].score = rsi < 30 ? 1 : rsi > 70 ? -1 : 0;
          this.metrics[rsiIdx].error = false;
          // MA (simple 5/20 cross for demo)
          const ma5 = closes.slice(-5).reduce((a, b) => a + b, 0) / 5;
          const ma20 = closes.length >= 20 ? closes.slice(-20).reduce((a, b) => a + b, 0) / 20 : ma5;
          this.metrics[maIdx].value = ma5 > ma20 ? 'Golden Cross' : 'Death Cross';
          this.metrics[maIdx].chartData = closes.slice(-5);
          this.metrics[maIdx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          this.metrics[maIdx].indicator = ma5 > ma20 ? 'Buy' : ma5 < ma20 ? 'Sell' : 'Hold';
          this.metrics[maIdx].score = ma5 > ma20 ? 1 : ma5 < ma20 ? -1 : 0;
          this.metrics[maIdx].error = false;
          // Volume (mock, as CoinGecko OHLC does not provide volume)
          this.metrics[volIdx].value = 'High Rising';
          this.metrics[volIdx].chartData = [1, 1, 1, 1, 1];
          this.metrics[volIdx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          this.metrics[volIdx].indicator = 'Buy';
          this.metrics[volIdx].score = 1;
          this.metrics[volIdx].error = false;
          // Bollinger Bands (mock, as CoinGecko OHLC does not provide stddev)
          this.metrics[bollIdx].value = 'N/A';
          this.metrics[bollIdx].chartData = closes.slice(-5);
          this.metrics[bollIdx].chartLabels = ['5d', '4d', '3d', '2d', 'Now'];
          this.metrics[bollIdx].indicator = 'Hold';
          this.metrics[bollIdx].score = 0;
          this.metrics[bollIdx].error = false;
        } else {
          this.metrics[rsiIdx].error = true;
          this.metrics[maIdx].error = true;
          this.metrics[volIdx].error = true;
          this.metrics[bollIdx].error = true;
        }
      } catch (e) {
        this.metrics[rsiIdx].error = true;
        this.metrics[maIdx].error = true;
        this.metrics[volIdx].error = true;
        this.metrics[bollIdx].error = true;
      } finally {
        this.calculateSignal();
      }
    },
    async fetchGoogleTrends() {
      const idx = this.metrics.findIndex(m => m.key === 'trends');
      try {
        const res = await fetch('http://localhost:5001/api/trends?keyword=bitcoin&timeframe=now 7-d');
        const data = await res.json();
        if (data && data.values && data.values.length > 0) {
          this.metrics[idx].value = data.values[data.values.length - 1];
          this.metrics[idx].chartData = data.values;
          this.metrics[idx].chartLabels = data.labels;
          // Scoring logic (example):
          const val = data.values[data.values.length - 1];
          if (val < 25) {
            this.metrics[idx].indicator = 'Buy';
            this.metrics[idx].score = 1;
          } else if (val > 75) {
            this.metrics[idx].indicator = 'Sell';
            this.metrics[idx].score = -1;
          } else {
            this.metrics[idx].indicator = 'Hold';
            this.metrics[idx].score = 0;
          }
          this.metrics[idx].error = false;
        } else {
          this.metrics[idx].error = true;
        }
      } catch (e) {
        this.metrics[idx].error = true;
      } finally {
        this.metrics[idx].loading = false;
        this.calculateSignal();
      }
    },
    calculateSignal() {
      // If any critical metric is missing, set error
      const criticalKeys = [
        'fear_greed', 'altcoin_index', 'btc_dominance', 'ssr', 'rsi', 'market_cap', 'ma_signal', 'volume'
      ];
      let error = null;
      for (const key of criticalKeys) {
        const m = this.metrics.find(m => m.key === key);
        if (!m || m.error) {
          error = 'One or more critical metrics are unavailable. Algorithmic signal may be inaccurate.';
          break;
        }
      }
      this.error = error;
      // Calculate weighted score
      const totalScore = error
        ? 0
        : this.metrics.reduce((sum, m) => sum + m.score * m.weight, 0);
      let signal = 'Hold';
      let asset = null;
      if (!error) {
        if (totalScore > 0.5) {
          signal = 'Buy';
          asset = this.metrics.find(m => m.key === 'altcoin_index').value > 75 ? 'Altcoins' : 'Bitcoin';
        } else if (totalScore < -0.5) {
          signal = 'Sell';
          asset = 'All';
        }
      }
      this.totalScore = totalScore;
      this.signal = signal;
      this.asset = asset;
    }
  }
});
</script> 