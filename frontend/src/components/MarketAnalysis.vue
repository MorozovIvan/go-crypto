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

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement);

const MetricCard = defineComponent({
  name: 'MetricCard',
  props: ['title', 'value', 'indicator', 'chartData', 'chartLabels', 'error'],
  components: { Line },
  template: `
    <div class="bg-white rounded-lg shadow p-4 flex flex-col">
      <div class="flex justify-between items-center mb-2">
        <h4 class="font-semibold">{{ title }}</h4>
        <span :class="indicatorClass">{{ error ? 'Error' : indicator }}</span>
      </div>
      <div class="text-2xl font-bold mb-2">
        <span v-if="!error">{{ value }}</span>
        <span v-else class="text-red-600">Error</span>
      </div>
      <Line v-if="!error" :data="chartConfig" :options="chartOptions" height="120" />
      <div v-else class="text-xs text-red-500">Data unavailable</div>
    </div>
  `,
  computed: {
    indicatorClass() {
      return {
        'text-green-600': this.indicator === 'Buy' && !this.error,
        'text-red-600': this.indicator === 'Sell' && !this.error,
        'text-gray-500': this.indicator === 'Hold' && !this.error,
        'text-red-600': this.error,
        'font-bold': true
      };
    },
    chartConfig() {
      return {
        labels: this.chartLabels,
        datasets: [
          {
            label: this.title,
            data: this.chartData,
            borderColor: '#0a78b9',
            backgroundColor: 'rgba(10,120,185,0.1)',
            tension: 0.3,
            fill: true
          }
        ]
      };
    },
    chartOptions() {
      return {
        responsive: true,
        plugins: {
          legend: { display: false },
          title: { display: false }
        },
        scales: {
          x: { display: false },
          y: { display: false }
        }
      };
    }
  }
});

export default defineComponent({
  name: 'MarketAnalysis',
  components: { MetricCard },
  data() {
    // Mock data for demonstration, with some errors
    const metrics = [
      {
        key: 'fear_greed',
        title: 'Fear & Greed Index',
        value: 25,
        indicator: 'Buy',
        score: 1,
        weight: 0.18,
        chartData: [40, 35, 30, 28, 25],
        chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
        error: false
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
        value: 45,
        indicator: 'Buy',
        score: 1,
        weight: 0.10,
        chartData: [55, 52, 50, 48, 45],
        chartLabels: ['5d', '4d', '3d', '2d', 'Now'],
        error: false
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
    ];

    // If any critical metric is missing, set error
    const criticalKeys = [
      'fear_greed', 'altcoin_index', 'btc_dominance', 'ssr', 'rsi', 'market_cap', 'ma_signal', 'volume'
    ];
    let error = null;
    for (const key of criticalKeys) {
      const m = metrics.find(m => m.key === key);
      if (!m || m.error) {
        error = 'One or more critical metrics are unavailable. Algorithmic signal may be inaccurate.';
        break;
      }
    }

    // Calculate weighted score
    const totalScore = error
      ? 0
      : metrics.reduce((sum, m) => sum + m.score * m.weight, 0);
    let signal = 'Hold';
    let asset = null;
    if (!error) {
      if (totalScore > 0.5) {
        signal = 'Buy';
        asset = metrics.find(m => m.key === 'altcoin_index').value > 75 ? 'Altcoins' : 'Bitcoin';
      } else if (totalScore < -0.5) {
        signal = 'Sell';
        asset = 'All';
      }
    }
    return {
      metrics,
      totalScore,
      signal,
      asset,
      error
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
  }
});
</script> 