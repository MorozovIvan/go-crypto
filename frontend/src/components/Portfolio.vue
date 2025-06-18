<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900 to-indigo-900 p-6">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-4xl font-bold text-white mb-2">
        <span class="bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
          Portfolio Overview
        </span>
      </h1>
      <p class="text-gray-300 text-lg">Manage and track your crypto portfolio across exchanges</p>
    </div>

    <!-- Exchange Tabs -->
    <div class="mb-8">
      <div class="flex space-x-1 bg-gray-800/50 p-1 rounded-lg backdrop-blur-sm">
        <button
          @click="activeTab = 'binance'"
          :class="[
            'flex-1 py-3 px-6 rounded-md font-medium transition-all duration-200',
            activeTab === 'binance'
              ? 'bg-yellow-500 text-black shadow-lg'
              : 'text-gray-300 hover:bg-gray-700/50 hover:text-white'
          ]"
        >
          <div class="flex items-center justify-center space-x-2">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2L8.5 5.5 12 9l3.5-3.5L12 2zm0 20l-3.5-3.5L12 15l3.5 3.5L12 22zm10-10l-3.5-3.5L15 12l3.5 3.5L22 12zM2 12l3.5-3.5L9 12l-3.5 3.5L2 12z"/>
            </svg>
            <span>Binance</span>
          </div>
        </button>
        <button
          @click="activeTab = 'okx'"
          :class="[
            'flex-1 py-3 px-6 rounded-md font-medium transition-all duration-200',
            activeTab === 'okx'
              ? 'bg-blue-500 text-white shadow-lg'
              : 'text-gray-300 hover:bg-gray-700/50 hover:text-white'
          ]"
        >
          <div class="flex items-center justify-center space-x-2">
            <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
            </svg>
            <span>OKX</span>
          </div>
        </button>
      </div>
    </div>

    <!-- Tab Content -->
    <div v-if="activeTab === 'binance'">
      <!-- Binance Portfolio -->
      <div v-if="loading" class="flex justify-center items-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-yellow-500"></div>
      </div>

      <div v-else-if="binancePortfolio">
        <!-- Portfolio Summary -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-gray-400 text-sm">Total Balance</p>
                <p class="text-2xl font-bold text-white">${{ formatNumber(binancePortfolio.totalValue || 0) }}</p>
              </div>
              <div class="p-3 bg-yellow-500/20 rounded-lg">
                <svg class="w-6 h-6 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
                </svg>
              </div>
            </div>
          </div>

          <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-gray-400 text-sm">24h Change</p>
                <p :class="[
                  'text-2xl font-bold',
                  (binancePortfolio.portfolioChange || 0) >= 0 ? 'text-green-400' : 'text-red-400'
                ]">
                  {{ formatPercentage(binancePortfolio.portfolioChange || 0) }}%
                </p>
              </div>
              <div :class="[
                'p-3 rounded-lg',
                (binancePortfolio.portfolioChange || 0) >= 0 ? 'bg-green-500/20' : 'bg-red-500/20'
              ]">
                <svg :class="[
                  'w-6 h-6',
                  (binancePortfolio.portfolioChange || 0) >= 0 ? 'text-green-400' : 'text-red-400'
                ]" fill="currentColor" viewBox="0 0 24 24">
                  <path v-if="(binancePortfolio.portfolioChange || 0) >= 0" d="M7 14l5-5 5 5z"/>
                  <path v-else d="M7 10l5 5 5-5z"/>
                </svg>
              </div>
            </div>
          </div>

          <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-gray-400 text-sm">Assets Count</p>
                <p class="text-2xl font-bold text-white">{{ binancePortfolio.assets?.length || 0 }}</p>
              </div>
              <div class="p-3 bg-blue-500/20 rounded-lg">
                <svg class="w-6 h-6 text-blue-400" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zM9 17H7v-7h2v7zm4 0h-2V7h2v10zm4 0h-2v-4h2v4z"/>
                </svg>
              </div>
            </div>
          </div>

          <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-gray-400 text-sm">24h Volume</p>
                <p class="text-2xl font-bold text-white">${{ formatNumber(binancePortfolio.volume24h || 0) }}</p>
              </div>
              <div class="p-3 bg-purple-500/20 rounded-lg">
                <svg class="w-6 h-6 text-purple-400" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/>
                </svg>
              </div>
            </div>
          </div>
        </div>

        <!-- Assets Grid -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div 
            v-for="asset in binancePortfolio.assets" 
            :key="asset.symbol"
            class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50 hover:border-gray-600/50 transition-all duration-200"
          >
            <!-- Asset Header -->
            <div class="flex items-center justify-between mb-4">
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-gradient-to-r from-yellow-400 to-yellow-600 rounded-full flex items-center justify-center">
                  <span class="text-black font-bold text-sm">{{ asset.symbol?.slice(0, 2) || 'N/A' }}</span>
                </div>
                <div>
                  <h3 class="text-white font-semibold">{{ asset.symbol || 'Unknown' }}</h3>
                  <p class="text-gray-400 text-sm">{{ asset.name || 'Unknown Asset' }}</p>
                </div>
              </div>
              <div class="text-right">
                <p class="text-white font-semibold">${{ formatNumber(asset.value || 0) }}</p>
                <p :class="[
                  'text-sm font-medium',
                  (asset.change || 0) >= 0 ? 'text-green-400' : 'text-red-400'
                ]">
                  {{ formatPercentage(asset.change || 0) }}%
                </p>
              </div>
            </div>

            <!-- Asset Details -->
            <div class="space-y-3">
              <div class="flex justify-between">
                <span class="text-gray-400">Amount</span>
                <span class="text-white font-medium">{{ formatNumber(asset.amount || 0) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-400">Price</span>
                <span class="text-white font-medium">${{ formatNumber(asset.price || 0) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-400">24h High</span>
                <span class="text-white font-medium">${{ formatNumber(asset.high24h || 0) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-400">24h Low</span>
                <span class="text-white font-medium">${{ formatNumber(asset.low24h || 0) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-400">24h Volume</span>
                <span class="text-white font-medium">${{ formatNumber(asset.volume24h || 0) }}</span>
              </div>
            </div>

            <!-- Price Chart Placeholder -->
            <div class="mt-6 h-20 bg-gray-700/30 rounded-lg flex items-center justify-center">
              <span class="text-gray-500 text-sm">Price Chart</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="text-center py-20">
        <div class="bg-red-500/20 border border-red-500/30 rounded-xl p-8 max-w-md mx-auto">
          <svg class="w-12 h-12 text-red-400 mx-auto mb-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
          </svg>
          <h3 class="text-red-400 font-semibold mb-2">Failed to Load Portfolio</h3>
          <p class="text-gray-400 text-sm mb-4">{{ error }}</p>
          <button 
            @click="fetchBinancePortfolio"
            class="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded-lg transition-colors duration-200"
          >
            Retry
          </button>
        </div>
      </div>
    </div>

    <!-- OKX Tab Content -->
    <div v-else-if="activeTab === 'okx'">
      <div class="text-center py-20">
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-8 max-w-md mx-auto border border-gray-700/50">
          <svg class="w-16 h-16 text-blue-400 mx-auto mb-4" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
          </svg>
          <h3 class="text-white font-semibold text-xl mb-2">OKX Integration</h3>
          <p class="text-gray-400 mb-6">OKX portfolio integration is coming soon. Connect your OKX account to view your portfolio data here.</p>
          <button class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg transition-colors duration-200">
            Connect OKX Account
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Portfolio',
  data() {
    return {
      activeTab: 'binance',
      loading: false,
      error: null,
      binancePortfolio: null,
      refreshInterval: null
    }
  },
  methods: {
    async fetchBinancePortfolio() {
      this.loading = true
      this.error = null
      
      try {
        const response = await fetch('/api/portfolio')
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const data = await response.json()
        this.binancePortfolio = data
        console.log('Binance Portfolio Data:', data)
      } catch (error) {
        console.error('Failed to fetch Binance portfolio:', error)
        this.error = error.message
      } finally {
        this.loading = false
      }
    },
    
    formatNumber(value) {
      if (value === null || value === undefined) return '0.00'
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(value)
    },
    
    formatPercentage(value) {
      if (value === null || value === undefined) return '0.00'
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
        signDisplay: 'always'
      }).format(value)
    },
    
    startAutoRefresh() {
      // Refresh portfolio data every 2 minutes
      this.refreshInterval = setInterval(() => {
        if (this.activeTab === 'binance') {
          this.fetchBinancePortfolio()
        }
      }, 120000)
    },
    
    stopAutoRefresh() {
      if (this.refreshInterval) {
        clearInterval(this.refreshInterval)
        this.refreshInterval = null
      }
    }
  },
  
  watch: {
    activeTab(newTab) {
      if (newTab === 'binance' && !this.binancePortfolio) {
        this.fetchBinancePortfolio()
      }
    }
  },
  
  mounted() {
    // Load Binance portfolio on component mount
    this.fetchBinancePortfolio()
    this.startAutoRefresh()
  },
  
  beforeUnmount() {
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
/* Additional custom styles if needed */
.backdrop-blur-sm {
  backdrop-filter: blur(4px);
}
</style> 