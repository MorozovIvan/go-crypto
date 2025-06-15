<template>
  <div class="p-8">
    <h2 class="text-2xl font-bold mb-6 flex items-center gap-2">
      <svg class="w-7 h-7 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <rect x="3" y="13" width="4" height="8" rx="1" stroke-width="2" stroke="currentColor"/>
        <rect x="9" y="9" width="4" height="12" rx="1" stroke-width="2" stroke="currentColor"/>
        <rect x="15" y="5" width="4" height="16" rx="1" stroke-width="2" stroke="currentColor"/>
      </svg>
      Adequate Binance Statistic
    </h2>
    <div class="mb-4 flex gap-4 items-center">
      <button @click="showConnectModal = true" class="px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600">Connect Binance</button>
      <span v-if="binanceConnected" class="text-green-600 font-semibold">Connected</span>
    </div>
    <!-- Connect Modal -->
    <div v-if="showConnectModal" class="fixed inset-0 bg-black bg-opacity-40 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg shadow-lg p-8 max-w-sm w-full relative">
        <button @click="showConnectModal = false" class="absolute top-2 right-2 text-gray-400 hover:text-gray-700">&times;</button>
        <h3 class="text-xl font-bold mb-4">Connect to Binance</h3>
        <form @submit.prevent="submitBinanceKeys">
          <div class="mb-4">
            <label class="block text-sm font-semibold mb-1">API Key</label>
            <input v-model="apiKey" type="text" class="w-full px-3 py-2 border rounded" required />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-semibold mb-1">API Secret</label>
            <input v-model="apiSecret" type="password" class="w-full px-3 py-2 border rounded" required />
          </div>
          <div class="flex justify-end gap-2">
            <button type="button" @click="showConnectModal = false" class="px-4 py-2 border rounded">Cancel</button>
            <button type="submit" class="px-4 py-2 bg-yellow-500 text-white rounded hover:bg-yellow-600">Connect</button>
          </div>
        </form>
        <div v-if="connectError" class="text-red-600 mt-2 text-sm">{{ connectError }}</div>
      </div>
    </div>
    <div class="mb-6 flex flex-wrap gap-4 items-end">
      <div>
        <label class="block text-xs font-semibold mb-1">From</label>
        <input type="date" v-model="from" class="px-2 py-1 border rounded" />
      </div>
      <div>
        <label class="block text-xs font-semibold mb-1">To</label>
        <input type="date" v-model="to" class="px-2 py-1 border rounded" />
      </div>
      <button @click="fetchStats" class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Fetch Statistics</button>
    </div>
    <div v-if="loading" class="text-center py-8 text-blue-600">Loading...</div>
    <div v-else>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div class="bg-white rounded shadow p-4">
          <h3 class="font-semibold mb-2">Total Trades</h3>
          <div class="text-2xl font-bold">{{ stats.totalTrades }}</div>
        </div>
        <div class="bg-white rounded shadow p-4">
          <h3 class="font-semibold mb-2">Total PNL</h3>
          <div class="text-2xl font-bold" :class="stats.totalPNL >= 0 ? 'text-green-600' : 'text-red-600'">{{ stats.totalPNL }}</div>
        </div>
        <div class="bg-white rounded shadow p-4">
          <h3 class="font-semibold mb-2">Win Rate</h3>
          <div class="text-2xl font-bold">{{ stats.winRate }}%</div>
        </div>
        <div class="bg-white rounded shadow p-4">
          <h3 class="font-semibold mb-2">Most Traded Pair</h3>
          <div class="text-2xl font-bold">{{ stats.mostTradedPair }}</div>
        </div>
      </div>
      <!-- Revenue Graph (Total) -->
      <div class="bg-white rounded shadow p-6 mb-8">
        <h3 class="font-semibold mb-4">Revenue Over Time</h3>
        <div class="h-64">
          <canvas ref="revenueChart"></canvas>
        </div>
      </div>
      <!-- Revenue Graph (Per Token) -->
      <div class="bg-white rounded shadow p-6 mb-8">
        <h3 class="font-semibold mb-4">Revenue by Token</h3>
        <div class="h-64">
          <canvas ref="tokenRevenueChart"></canvas>
        </div>
      </div>
      <!-- Table: Price of Buying Each Coin -->
      <div class="bg-white rounded shadow p-6 mb-8">
        <h3 class="font-semibold mb-4">Price of Buying Each Coin</h3>
        <table class="min-w-full divide-y divide-gray-200">
          <thead>
            <tr>
              <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Token</th>
              <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Avg Buy Price</th>
              <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase">Total Bought</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in stats.buyPrices" :key="row.token">
              <td class="px-4 py-2 font-mono">{{ row.token }}</td>
              <td class="px-4 py-2">{{ row.avgPrice }}</td>
              <td class="px-4 py-2">{{ row.total }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import { Chart, registerables } from 'chart.js';
Chart.register(...registerables);

export default {
  name: 'AdequateBinanceStatistic',
  data() {
    return {
      from: '',
      to: '',
      loading: false,
      binanceConnected: false,
      showConnectModal: false,
      apiKey: '',
      apiSecret: '',
      connectError: '',
      stats: {
        totalTrades: 123,
        totalPNL: 2500.75,
        winRate: 67,
        mostTradedPair: 'BTC/USDT',
        buyPrices: [
          { token: 'BTC', avgPrice: 42000, total: 0.5 },
          { token: 'ETH', avgPrice: 2500, total: 2 },
          { token: 'BNB', avgPrice: 350, total: 10 }
        ],
        revenueHistory: [100, 300, 500, 800, 1200, 1800, 2500],
        revenueLabels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
        tokenRevenue: {
          BTC: [50, 100, 200, 400, 600, 900, 1200],
          ETH: [20, 80, 150, 200, 300, 500, 800],
          BNB: [30, 120, 150, 200, 300, 400, 500]
        }
      },
      revenueChart: null,
      tokenRevenueChart: null
    }
  },
  methods: {
    connectBinance() {
      this.showConnectModal = true;
    },
    async submitBinanceKeys() {
      this.connectError = '';
      try {
        const res = await fetch('/api/binance/connect', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ apiKey: this.apiKey, apiSecret: this.apiSecret })
        });
        if (!res.ok) throw new Error('Failed to connect to Binance API');
        this.binanceConnected = true;
        this.showConnectModal = false;
        this.apiKey = '';
        this.apiSecret = '';
      } catch (err) {
        this.connectError = err.message || 'Connection failed';
      }
    },
    fetchStats() {
      this.loading = true;
      setTimeout(() => {
        this.loading = false;
        this.$nextTick(() => {
          this.renderCharts();
        });
      }, 1000);
    },
    renderCharts() {
      // Revenue Over Time (Total)
      if (this.revenueChart) this.revenueChart.destroy();
      const ctx = this.$refs.revenueChart.getContext('2d');
      this.revenueChart = new Chart(ctx, {
        type: 'line',
        data: {
          labels: this.stats.revenueLabels,
          datasets: [{
            label: 'Revenue',
            data: this.stats.revenueHistory,
            borderColor: '#2563eb',
            backgroundColor: 'rgba(37,99,235,0.1)',
            fill: true
          }]
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          plugins: { legend: { display: false } }
        }
      });
      // Revenue by Token (Multi-line)
      if (this.tokenRevenueChart) this.tokenRevenueChart.destroy();
      const ctx2 = this.$refs.tokenRevenueChart.getContext('2d');
      this.tokenRevenueChart = new Chart(ctx2, {
        type: 'line',
        data: {
          labels: this.stats.revenueLabels,
          datasets: Object.keys(this.stats.tokenRevenue).map(token => ({
            label: token,
            data: this.stats.tokenRevenue[token],
            borderColor: token === 'BTC' ? '#f7931a' : token === 'ETH' ? '#627eea' : '#f3ba2f',
            backgroundColor: 'rgba(0,0,0,0)',
            fill: false
          }))
        },
        options: {
          responsive: true,
          maintainAspectRatio: false
        }
      });
    }
  },
  mounted() {
    this.renderCharts();
  }
}
</script> 