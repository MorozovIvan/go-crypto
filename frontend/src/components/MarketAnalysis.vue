<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-50 p-6">
    <!-- Header Section with Enhanced Controls -->
    <div class="mb-8">
      <div class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 mb-6">
        <div>
          <h1 class="text-4xl font-bold bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
            Market Analysis Dashboard
          </h1>
          <p class="text-gray-600 mt-2">Real-time crypto market intelligence with 26 advanced metrics</p>
        </div>
        
        <!-- Enhanced Controls -->
        <div class="flex flex-col sm:flex-row gap-4">
          <!-- Data Source Switcher -->
          <div class="flex items-center space-x-3 bg-white p-3 rounded-xl shadow-lg border border-gray-200">
            <span class="text-sm font-medium text-gray-700">Data:</span>
            <div class="flex items-center space-x-2">
              <label class="relative inline-flex items-center cursor-pointer">
                <input 
                  type="checkbox" 
                  v-model="useMockData" 
                  @change="toggleDataSource"
                  class="sr-only peer"
                >
                <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
              </label>
              <div class="flex flex-col text-xs">
                <span :class="useMockData ? 'text-blue-600 font-semibold' : 'text-gray-500'">Mock</span>
                <span :class="!useMockData ? 'text-green-600 font-semibold' : 'text-gray-500'">Live</span>
              </div>
            </div>
            <div :class="useMockData ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'" 
                 class="px-3 py-1 text-xs font-medium rounded-full">
              {{ useMockData ? 'Enhanced Mock' : 'Real APIs' }}
            </div>
          </div>

          <!-- Auto-refresh Toggle -->
          <div class="flex items-center space-x-3 bg-white p-3 rounded-xl shadow-lg border border-gray-200">
            <span class="text-sm font-medium text-gray-700">Auto-refresh:</span>
            <label class="relative inline-flex items-center cursor-pointer">
              <input 
                type="checkbox" 
                v-model="autoRefresh" 
                @change="toggleAutoRefresh"
                class="sr-only peer"
              >
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-green-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"></div>
            </label>
            <span class="text-xs" :class="autoRefresh ? 'text-green-600' : 'text-gray-500'">
              {{ autoRefresh ? 'ON' : 'OFF' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Status Bar -->
      <div class="flex flex-wrap items-center gap-4 p-4 bg-white rounded-xl shadow-lg border border-gray-200">
        <div class="flex items-center space-x-2">
          <div :class="connectionStatus === 'connected' ? 'bg-green-500' : 'bg-red-500'" 
               class="w-3 h-3 rounded-full animate-pulse"></div>
          <span class="text-sm font-medium">
            {{ connectionStatus === 'connected' ? 'Connected' : 'Disconnected' }}
          </span>
        </div>
        <div class="text-sm text-gray-600">
          Coverage: <span class="font-semibold">{{ dataCoverage }}%</span>
        </div>
        <div class="text-sm text-gray-600">
          Last Update: <span class="font-mono">{{ lastUpdateTime }}</span>
        </div>
        <div class="text-sm text-gray-600">
          Next Refresh: <span class="font-mono">{{ nextRefreshIn }}s</span>
        </div>
      </div>
    </div>

    <!-- Alert Banner -->
    <div v-if="error" class="mb-6 p-4 bg-red-50 border-l-4 border-red-400 rounded-lg">
      <div class="flex">
        <div class="flex-shrink-0">
          <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="ml-3">
          <p class="text-sm text-red-700"><strong>Warning:</strong> {{ error }}</p>
        </div>
      </div>
    </div>

    <!-- Main Content Layout -->
    <div class="grid grid-cols-1 xl:grid-cols-4 gap-8">
      
      <!-- Left Column: Signal Summary & Key Metrics -->
      <div class="xl:col-span-1 space-y-6">
        
        <!-- AI Signal Card -->
        <div class="bg-white rounded-2xl shadow-xl border border-gray-200 overflow-hidden">
          <div class="bg-gradient-to-r from-blue-600 to-indigo-600 p-6 text-white">
            <h3 class="text-xl font-bold mb-2">🤖 AI Trading Signal</h3>
            <div class="text-blue-100 text-sm">Advanced algorithmic analysis</div>
          </div>
          
          <div class="p-6 space-y-4">
            <!-- Signal Display -->
            <div class="text-center">
              <div class="text-3xl font-bold mb-2" :class="signalColorClass">
                {{ signal }}
              </div>
              <div v-if="asset" class="text-gray-600 text-sm mb-3">
                Recommended Asset: <span class="font-semibold">{{ asset }}</span>
              </div>
              
              <!-- Confidence Meter -->
              <div class="mb-4">
                <div class="flex justify-between text-sm mb-1">
                  <span>Confidence</span>
                  <span class="font-semibold" :class="confidenceColorClass">{{ confidence }}</span>
                </div>
                <div class="w-full bg-gray-200 rounded-full h-3">
                  <div class="h-3 rounded-full transition-all duration-500" 
                       :class="confidenceBarClass" 
                       :style="{ width: confidencePercentage + '%' }"></div>
                </div>
                <div class="text-xs text-gray-500 mt-1">{{ confidencePercentage }}%</div>
              </div>
              
              <!-- Score Display -->
              <div class="bg-gray-50 rounded-lg p-3">
                <div class="text-sm text-gray-600 mb-1">Weighted Score</div>
                <div class="text-2xl font-mono font-bold" :class="scoreColorClass">
                  {{ totalScore.toFixed(3) }}
                </div>
                <div class="text-xs text-gray-500">Range: -1.0 to +1.0</div>
              </div>
            </div>

            <!-- Quick Stats -->
            <div class="grid grid-cols-3 gap-3 pt-4 border-t">
              <div class="text-center">
                <div class="text-lg font-bold text-green-600">{{ bullishCount }}</div>
                <div class="text-xs text-gray-500">Bullish</div>
              </div>
              <div class="text-center">
                <div class="text-lg font-bold text-gray-600">{{ neutralCount }}</div>
                <div class="text-xs text-gray-500">Neutral</div>
              </div>
              <div class="text-center">
                <div class="text-lg font-bold text-red-600">{{ bearishCount }}</div>
                <div class="text-xs text-gray-500">Bearish</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Market Health Overview -->
        <div class="bg-white rounded-2xl shadow-xl border border-gray-200 p-6">
          <h4 class="text-lg font-bold mb-4 flex items-center">
            <span class="mr-2">📊</span> Market Health
          </h4>
          
          <div class="space-y-3">
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-600">Data Quality</span>
              <div class="flex items-center space-x-2">
                <div class="w-16 bg-gray-200 rounded-full h-2">
                  <div class="bg-green-500 h-2 rounded-full transition-all duration-500" 
                       :style="{ width: dataQuality + '%' }"></div>
                </div>
                <span class="text-sm font-semibold">{{ dataQuality }}%</span>
              </div>
            </div>
            
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-600">Signal Strength</span>
              <div class="flex items-center space-x-2">
                <div class="w-16 bg-gray-200 rounded-full h-2">
                  <div class="bg-blue-500 h-2 rounded-full transition-all duration-500" 
                       :style="{ width: signalStrength * 100 + '%' }"></div>
                </div>
                <span class="text-sm font-semibold">{{ (signalStrength * 100).toFixed(0) }}%</span>
              </div>
            </div>
            
            <div class="flex justify-between items-center">
              <span class="text-sm text-gray-600">Consensus</span>
              <div class="flex items-center space-x-2">
                <div class="w-16 bg-gray-200 rounded-full h-2">
                  <div class="bg-purple-500 h-2 rounded-full transition-all duration-500" 
                       :style="{ width: consensusRatio * 100 + '%' }"></div>
                </div>
                <span class="text-sm font-semibold">{{ (consensusRatio * 100).toFixed(0) }}%</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Top Movers -->
        <div class="bg-white rounded-2xl shadow-xl border border-gray-200 p-6">
          <h4 class="text-lg font-bold mb-4 flex items-center">
            <span class="mr-2">🔥</span> Signal Drivers
          </h4>
          
          <div class="space-y-3">
            <div v-for="metric in topMovers" :key="metric.key" class="flex justify-between items-center">
              <div class="flex-1">
                <div class="text-sm font-medium">{{ metric.title }}</div>
                <div class="text-xs text-gray-500">Weight: {{ (metric.weight * 100).toFixed(0) }}%</div>
              </div>
              <div class="text-right">
                <div class="text-sm font-bold" :class="getScoreColor(metric.score)">
                  {{ metric.score.toFixed(2) }}
                </div>
                <div class="text-xs" :class="getIndicatorColor(metric.indicator)">
                  {{ metric.indicator }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Metrics Grid -->
      <div class="xl:col-span-3">
        
        <!-- Filter and Sort Controls -->
        <div class="mb-6 flex flex-wrap gap-4 items-center justify-between">
          <div class="flex flex-wrap gap-2">
            <button 
              v-for="category in categories" 
              :key="category"
              @click="selectedCategory = category"
              :class="selectedCategory === category ? 'bg-blue-600 text-white' : 'bg-white text-gray-700 hover:bg-gray-50'"
              class="px-4 py-2 rounded-lg border border-gray-300 text-sm font-medium transition-colors"
            >
              {{ category }}
            </button>
          </div>
          
          <div class="flex items-center space-x-3">
            <select v-model="sortBy" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
              <option value="weight">Sort by Weight</option>
              <option value="score">Sort by Score</option>
              <option value="alphabetical">Sort A-Z</option>
            </select>
            
            <button 
              @click="refreshAllMetrics" 
              :disabled="isRefreshing"
              class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors flex items-center space-x-2"
            >
              <svg :class="isRefreshing ? 'animate-spin' : ''" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
              </svg>
              <span>{{ isRefreshing ? 'Refreshing...' : 'Refresh' }}</span>
            </button>
          </div>
        </div>

        <!-- Enhanced Metrics Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
          <EnhancedMetricCard
            v-for="metric in filteredAndSortedMetrics"
            :key="metric.key"
            :title="metric.title"
            :value="metric.value"
            :indicator="metric.indicator"
            :score="metric.score"
            :weight="metric.weight"
            :chart-data="metric.chartData"
            :chart-labels="metric.chartLabels"
            :error="metric.error"
            :loading="metric.loading"
            :last-updated="metric.lastUpdated"
            @refresh="fetchMetric(metric.key)"
          />
        </div>
      </div>
    </div>

    <!-- Score Breakdown Modal -->
    <div v-if="showBreakdown" @click="showBreakdown = false" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div @click.stop class="bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-auto">
        <div class="p-6 border-b">
          <div class="flex justify-between items-center">
            <h3 class="text-2xl font-bold">Detailed Score Breakdown</h3>
            <button @click="showBreakdown = false" class="text-gray-500 hover:text-gray-700">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
              </svg>
            </button>
          </div>
        </div>
        
        <div class="p-6">
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b">
                  <th class="text-left py-2">Metric</th>
                  <th class="text-center py-2">Value</th>
                  <th class="text-center py-2">Score</th>
                  <th class="text-center py-2">Weight</th>
                  <th class="text-center py-2">Contribution</th>
                  <th class="text-center py-2">Signal</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="metric in metrics" :key="metric.key" class="border-b hover:bg-gray-50">
                  <td class="py-3">
                    <div class="font-medium">{{ metric.title }}</div>
                    <div class="text-xs text-gray-500">{{ metric.key }}</div>
                  </td>
                  <td class="text-center py-3">
                    <span v-if="!metric.error" class="font-mono">{{ formatValue(metric.value) }}</span>
                    <span v-else class="text-red-500">Error</span>
                  </td>
                  <td class="text-center py-3">
                    <span class="font-mono" :class="getScoreColor(metric.score)">
                      {{ metric.score.toFixed(3) }}
                    </span>
                  </td>
                  <td class="text-center py-3">{{ (metric.weight * 100).toFixed(1) }}%</td>
                  <td class="text-center py-3">
                    <span class="font-mono" :class="getScoreColor(metric.score * metric.weight)">
                      {{ (metric.score * metric.weight).toFixed(4) }}
                    </span>
                  </td>
                  <td class="text-center py-3">
                    <span class="px-2 py-1 rounded-full text-xs font-medium" :class="getIndicatorBadgeClass(metric.indicator)">
                      {{ metric.indicator }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Floating Action Button for Breakdown -->
    <button 
      @click="showBreakdown = true"
      class="fixed bottom-6 right-6 bg-blue-600 text-white p-4 rounded-full shadow-lg hover:bg-blue-700 transition-colors z-40"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
      </svg>
    </button>
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
import EnhancedMetricCard from './EnhancedMetricCard.vue';

ChartJS.register(Title, Tooltip, Legend, LineElement, CategoryScale, LinearScale, PointElement, Filler);

export default defineComponent({
  name: 'MarketAnalysis',
  components: { MetricCard, EnhancedMetricCard },
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
      lastUpdateTime: null,
      useMockData: false,
      autoRefresh: false,
      connectionStatus: 'disconnected',
      dataCoverage: 0,
      nextRefreshIn: 0,
      showBreakdown: false,
      selectedCategory: 'All',
      sortBy: 'weight',
      isRefreshing: false,
      signalStrength: 0,
      consensusRatio: 0,
      dataQuality: 0,
      bullishCount: 0,
      neutralCount: 0,
      bearishCount: 0,
      signalColorClass: {},
      confidenceColorClass: {},
      confidenceBarClass: {},
      scoreColorClass: {},
      confidencePercentage: 0,
      topMovers: [],
      categories: ['All', 'Fear & Greed', 'BTC Dominance', 'RSI', 'Market Cap', 'Google Trends'],
      filteredAndSortedMetrics: []
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
    
    // Initialize enhanced UX features
    this.nextRefreshIn = 30;
    this.lastUpdateTime = new Date().toLocaleTimeString();
    this.updateConnectionStatus();
    this.updateRefreshTimer();
    
    // Update UI state every second
    setInterval(() => {
      this.lastUpdateTime = new Date().toLocaleTimeString();
      this.updateConnectionStatus();
    }, 1000);
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
    
    // Toggle between mock and real data sources
    toggleDataSource() {
      console.log(`[toggleDataSource] Switching to ${this.useMockData ? 'Mock' : 'Real'} data`);
      
      // Reset all metrics to loading state
      this.metrics.forEach(metric => {
        metric.loading = true;
        metric.error = false;
        metric.value = null;
        metric.indicator = 'Hold';
        metric.score = 0;
        metric.chartData = [];
        metric.chartLabels = [];
      });
      
      // Close existing WebSocket connection
      this.closeWebSocket();
      
      // Refetch all metrics with new data source
      this.fetchAllMetrics();
      
      // Reinitialize WebSocket connection
      setTimeout(() => {
        this.initWebSocket();
      }, 1000);
      
      // Show notification about the switch
      this.showDataSourceNotification();
    },
    
    // Show notification about data source switch
    showDataSourceNotification() {
      const notification = document.createElement('div');
      notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 transition-all duration-300 ${
        this.useMockData 
          ? 'bg-blue-100 border-blue-400 text-blue-800' 
          : 'bg-green-100 border-green-400 text-green-800'
      }`;
      notification.innerHTML = `
        <div class="flex items-center space-x-2">
          <div class="font-semibold">
            ${this.useMockData ? '🧪 Mock Data Mode' : '📊 Live Data Mode'}
          </div>
          <div class="text-sm">
            Coverage: ${this.useMockData ? '100%' : '77%'}
          </div>
        </div>
        <div class="text-xs mt-1">
          ${this.useMockData 
            ? 'Using enhanced mock data for all metrics' 
            : 'Using real APIs where available, fallback for others'
          }
        </div>
      `;
      
      document.body.appendChild(notification);
      
      // Auto-remove after 4 seconds
      setTimeout(() => {
        notification.style.opacity = '0';
        setTimeout(() => {
          if (notification.parentNode) {
            notification.parentNode.removeChild(notification);
          }
        }, 300);
      }, 4000);
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
          console.log(`[fetchMetric] Fetching ${key} (attempt ${retryCount + 1}/${maxRetries}) - Mock: ${this.useMockData}`);
          const mockParam = this.useMockData ? '?mock=true' : '';
          const response = await fetch(`/api/${key}${mockParam}`, {
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
      this.signalStrength = signalStrength;
      this.consensusRatio = consensusRatio;
      this.dataQuality = dataQuality;
      this.bullishCount = bullishCount;
      this.neutralCount = neutralCount;
      this.bearishCount = bearishCount;
      this.confidencePercentage = Math.round(confidenceScore * 100);
      this.confidenceColorClass = {
        'text-green-600': confidence === 'High',
        'text-yellow-600': confidence === 'Medium',
        'text-red-600': confidence === 'Low'
      };
      this.scoreColorClass = {
        'text-green-600': totalScore > 0,
        'text-red-600': totalScore < 0
      };
      this.signalColorClass = {
        'text-green-600': signal === 'Strong Buy' || signal === 'Buy',
        'text-red-600': signal === 'Strong Sell' || signal === 'Sell',
        'text-yellow-600': signal === 'Hold'
      };
      this.topMovers = this.metrics.filter(m => m.weight > 0.05);
      this.filteredAndSortedMetrics = this.metrics.filter(m => this.selectedCategory === 'All' || m.title.includes(this.selectedCategory));
      this.filteredAndSortedMetrics.sort((a, b) => {
        if (this.sortBy === 'weight') {
          return b.weight - a.weight;
        } else if (this.sortBy === 'score') {
          return b.score - a.score;
        } else {
          return a.title.localeCompare(b.title);
        }
      });
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
    
    // Enhanced UX/UI Methods
    toggleDataSource() {
      this.fetchAllMetrics();
    },
    
    toggleAutoRefresh() {
      if (this.autoRefresh) {
        this.autoRefreshInterval = setInterval(() => {
          this.refreshAllMetrics();
        }, 30000); // Refresh every 30 seconds
      } else {
        if (this.autoRefreshInterval) {
          clearInterval(this.autoRefreshInterval);
          this.autoRefreshInterval = null;
        }
      }
    },
    
    async refreshAllMetrics() {
      this.isRefreshing = true;
      try {
        await this.fetchAllMetrics();
      } finally {
        this.isRefreshing = false;
      }
    },
    
    formatValue(value) {
      if (value === null || value === undefined) return 'N/A';
      if (typeof value === 'number') {
        if (value > 1000000) return (value / 1000000).toFixed(1) + 'M';
        if (value > 1000) return (value / 1000).toFixed(1) + 'K';
        return value.toFixed(2);
      }
      return String(value);
    },
    
    getScoreColor(score) {
      if (score > 0.3) return 'text-green-600';
      if (score < -0.3) return 'text-red-600';
      return 'text-gray-600';
    },
    
    getIndicatorColor(indicator) {
      if (indicator === 'Buy') return 'text-green-600';
      if (indicator === 'Sell') return 'text-red-600';
      return 'text-gray-600';
    },
    
    getIndicatorBadgeClass(indicator) {
      return {
        'bg-green-100 text-green-800': indicator === 'Buy',
        'bg-red-100 text-red-800': indicator === 'Sell',
        'bg-yellow-100 text-yellow-800': indicator === 'Hold'
      };
    },
    
    updateConnectionStatus() {
      this.connectionStatus = this.wsConnected ? 'connected' : 'disconnected';
      const availableMetrics = this.metrics.filter(m => !m.error).length;
      this.dataCoverage = Math.round((availableMetrics / this.metrics.length) * 100);
    },
    
    updateRefreshTimer() {
      if (this.autoRefresh) {
        this.refreshTimer = setInterval(() => {
          this.nextRefreshIn--;
          if (this.nextRefreshIn <= 0) {
            this.nextRefreshIn = 30;
            this.refreshAllMetrics();
          }
        }, 1000);
      } else {
        if (this.refreshTimer) {
          clearInterval(this.refreshTimer);
          this.refreshTimer = null;
        }
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