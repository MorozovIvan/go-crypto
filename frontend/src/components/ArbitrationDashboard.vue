<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-50 p-6">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-4xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            Solana DEX Arbitrage
          </h1>
          <p class="text-gray-600 mt-2">Real-time arbitrage opportunities across Solana DEXs</p>
        </div>
        <div class="flex items-center space-x-4">
          <div class="flex items-center space-x-2">
            <div class="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
            <span class="text-sm text-gray-600">Live</span>
          </div>
          <button 
            @click="refreshData"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center space-x-2"
          >
            <svg class="w-4 h-4" :class="{ 'animate-spin': isRefreshing }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span>Refresh</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-600">Active Opportunities</p>
            <p class="text-3xl font-bold text-green-600">{{ stats.activeOpportunities }}</p>
          </div>
          <div class="p-3 bg-green-100 rounded-lg">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center text-sm">
          <span class="text-green-500">+12%</span>
          <span class="text-gray-500 ml-1">vs last hour</span>
        </div>
      </div>

      <div class="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-600">Avg Profit Margin</p>
            <p class="text-3xl font-bold text-blue-600">{{ stats.avgProfitMargin }}%</p>
          </div>
          <div class="p-3 bg-blue-100 rounded-lg">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center text-sm">
          <span class="text-blue-500">5.2%</span>
          <span class="text-gray-500 ml-1">24h avg</span>
        </div>
      </div>

      <div class="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-600">Total Volume</p>
            <p class="text-3xl font-bold text-purple-600">${{ formatNumber(stats.totalVolume) }}</p>
          </div>
          <div class="p-3 bg-purple-100 rounded-lg">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center text-sm">
          <span class="text-purple-500">+8.3%</span>
          <span class="text-gray-500 ml-1">last 24h</span>
        </div>
      </div>

      <div class="bg-white rounded-xl shadow-lg p-6 border border-gray-100">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-gray-600">Success Rate</p>
            <p class="text-3xl font-bold text-orange-600">{{ stats.successRate }}%</p>
          </div>
          <div class="p-3 bg-orange-100 rounded-lg">
            <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center text-sm">
          <span class="text-orange-500">+2.1%</span>
          <span class="text-gray-500 ml-1">this week</span>
        </div>
      </div>
    </div>

    <!-- Filters and Controls -->
    <div class="bg-white rounded-xl shadow-lg p-6 mb-8 border border-gray-100">
      <div class="flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center space-x-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Min Profit</label>
            <select v-model="filters.minProfit" class="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option value="0">All</option>
              <option value="2">2%+</option>
              <option value="5">5%+</option>
              <option value="10">10%+</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">DEX Pair</label>
            <select v-model="filters.dexPair" class="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option value="all">All Pairs</option>
              <option value="raydium-orca">Raydium ↔ Orca</option>
              <option value="raydium-jupiter">Raydium ↔ Jupiter</option>
              <option value="orca-jupiter">Orca ↔ Jupiter</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-1">Token Type</label>
            <select v-model="filters.tokenType" class="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option value="all">All Tokens</option>
              <option value="meme">Memecoins</option>
              <option value="defi">DeFi Tokens</option>
              <option value="stable">Stablecoins</option>
            </select>
          </div>
        </div>
        <div class="flex items-center space-x-2">
          <button 
            @click="toggleAutoRefresh"
            :class="[
              'px-4 py-2 rounded-lg transition-colors flex items-center space-x-2',
              autoRefresh ? 'bg-green-600 text-white' : 'bg-gray-200 text-gray-700'
            ]"
          >
            <div class="w-2 h-2 rounded-full" :class="autoRefresh ? 'bg-white animate-pulse' : 'bg-gray-400'"></div>
            <span>Auto Refresh</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Arbitrage Opportunities Table -->
    <div class="bg-white rounded-xl shadow-lg border border-gray-100 overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-800">Live Arbitrage Opportunities</h2>
        <p class="text-sm text-gray-600 mt-1">{{ filteredOpportunities.length }} opportunities found</p>
      </div>
      
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Token</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">DEX Pair</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Prices</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Profit</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Volume</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Liquidity</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="opportunity in filteredOpportunities" :key="opportunity.id" class="hover:bg-gray-50 transition-colors">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="flex-shrink-0 h-10 w-10">
                    <img class="h-10 w-10 rounded-full" :src="opportunity.token.logo" :alt="opportunity.token.symbol">
                  </div>
                  <div class="ml-4">
                    <div class="text-sm font-medium text-gray-900">{{ opportunity.token.symbol }}</div>
                    <div class="text-sm text-gray-500">{{ opportunity.token.name }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center space-x-2">
                  <span class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded">{{ opportunity.buyDex }}</span>
                  <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 8l4 4m0 0l-4 4m4-4H3" />
                  </svg>
                  <span class="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded">{{ opportunity.sellDex }}</span>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900">
                  <div class="flex items-center space-x-2">
                    <span class="text-red-600">${{ opportunity.buyPrice.toFixed(6) }}</span>
                    <span class="text-gray-400">→</span>
                    <span class="text-green-600">${{ opportunity.sellPrice.toFixed(6) }}</span>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <span class="text-lg font-bold text-green-600">{{ opportunity.profitPercent.toFixed(2) }}%</span>
                  <div class="ml-2 w-16 bg-gray-200 rounded-full h-2">
                    <div 
                      class="bg-green-500 h-2 rounded-full transition-all duration-300" 
                      :style="{ width: Math.min(opportunity.profitPercent * 5, 100) + '%' }"
                    ></div>
                  </div>
                </div>
                <div class="text-sm text-gray-500">${{ opportunity.profitAmount.toFixed(2) }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                ${{ formatNumber(opportunity.volume24h) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <span class="text-sm text-gray-900">${{ formatNumber(opportunity.liquidity) }}</span>
                  <div class="ml-2">
                    <span 
                      :class="[
                        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                        opportunity.liquidityLevel === 'High' ? 'bg-green-100 text-green-800' :
                        opportunity.liquidityLevel === 'Medium' ? 'bg-yellow-100 text-yellow-800' :
                        'bg-red-100 text-red-800'
                      ]"
                    >
                      {{ opportunity.liquidityLevel }}
                    </span>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <button 
                  @click="executeArbitrage(opportunity)"
                  class="bg-gradient-to-r from-blue-500 to-purple-600 text-white px-4 py-2 rounded-lg hover:from-blue-600 hover:to-purple-700 transition-all duration-200 flex items-center space-x-2"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                  <span>Execute</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Recent Trades -->
    <div class="mt-8 bg-white rounded-xl shadow-lg border border-gray-100">
      <div class="px-6 py-4 border-b border-gray-200">
        <h2 class="text-xl font-semibold text-gray-800">Recent Arbitrage Trades</h2>
      </div>
      <div class="p-6">
        <div class="space-y-4">
          <div 
            v-for="trade in recentTrades" 
            :key="trade.id"
            class="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
          >
            <div class="flex items-center space-x-4">
              <img class="h-8 w-8 rounded-full" :src="trade.token.logo" :alt="trade.token.symbol">
              <div>
                <div class="font-medium text-gray-900">{{ trade.token.symbol }}</div>
                <div class="text-sm text-gray-500">{{ trade.dexPair }}</div>
              </div>
            </div>
            <div class="text-right">
              <div class="font-medium text-green-600">+{{ trade.profit.toFixed(2) }}%</div>
              <div class="text-sm text-gray-500">${{ trade.amount.toFixed(2) }}</div>
            </div>
            <div class="text-sm text-gray-500">{{ trade.timestamp }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ArbitrationDashboard',
  data() {
    return {
      isRefreshing: false,
      autoRefresh: true,
      filters: {
        minProfit: 0,
        dexPair: 'all',
        tokenType: 'all'
      },
      stats: {
        activeOpportunities: 47,
        avgProfitMargin: 7.8,
        totalVolume: 2450000,
        successRate: 94.2
      },
      opportunities: [
        {
          id: 1,
          token: {
            symbol: 'BONK',
            name: 'Bonk',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263/logo.png'
          },
          buyDex: 'Raydium',
          sellDex: 'Orca',
          buyPrice: 0.000012,
          sellPrice: 0.000013,
          profitPercent: 8.33,
          profitAmount: 125.50,
          volume24h: 850000,
          liquidity: 2500000,
          liquidityLevel: 'High',
          type: 'meme'
        },
        {
          id: 2,
          token: {
            symbol: 'WIF',
            name: 'dogwifhat',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/EKpQGSJtjMFqKZ9KQanSqYXRcF8fBopzLHYxdM65zcjm/logo.png'
          },
          buyDex: 'Jupiter',
          sellDex: 'Raydium',
          buyPrice: 1.245,
          sellPrice: 1.334,
          profitPercent: 7.15,
          profitAmount: 89.00,
          volume24h: 1200000,
          liquidity: 3200000,
          liquidityLevel: 'High',
          type: 'meme'
        },
        {
          id: 3,
          token: {
            symbol: 'PEPE',
            name: 'Pepe',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/BxzBCZJp8KFGhJwzWcQW1bwKqjYHbqKnWoaXWKR8Zxz/logo.png'
          },
          buyDex: 'Orca',
          sellDex: 'Jupiter',
          buyPrice: 0.0000085,
          sellPrice: 0.0000091,
          profitPercent: 7.06,
          profitAmount: 70.60,
          volume24h: 650000,
          liquidity: 1800000,
          liquidityLevel: 'Medium',
          type: 'meme'
        },
        {
          id: 4,
          token: {
            symbol: 'SAMO',
            name: 'Samoyedcoin',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU/logo.png'
          },
          buyDex: 'Raydium',
          sellDex: 'Orca',
          buyPrice: 0.0145,
          sellPrice: 0.0154,
          profitPercent: 6.21,
          profitAmount: 62.10,
          volume24h: 420000,
          liquidity: 1200000,
          liquidityLevel: 'Medium',
          type: 'meme'
        },
        {
          id: 5,
          token: {
            symbol: 'FIDA',
            name: 'Bonfida',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/EchesyfXePKdLtoiZSL8pBe8Myagyy8ZRqsACNCFGnvp/logo.png'
          },
          buyDex: 'Jupiter',
          sellDex: 'Raydium',
          buyPrice: 0.234,
          sellPrice: 0.248,
          profitPercent: 5.98,
          profitAmount: 59.80,
          volume24h: 180000,
          liquidity: 800000,
          liquidityLevel: 'Low',
          type: 'defi'
        },
        {
          id: 6,
          token: {
            symbol: 'RAY',
            name: 'Raydium',
            logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/4k3Dyjzvzp8eMZWUXbBCjEvwSkkk59S5iCNLY3QrkX6R/logo.png'
          },
          buyDex: 'Orca',
          sellDex: 'Jupiter',
          buyPrice: 1.89,
          sellPrice: 1.99,
          profitPercent: 5.29,
          profitAmount: 52.90,
          volume24h: 950000,
          liquidity: 4500000,
          liquidityLevel: 'High',
          type: 'defi'
        }
      ],
      recentTrades: [
        {
          id: 1,
          token: { symbol: 'BONK', logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263/logo.png' },
          dexPair: 'Raydium → Orca',
          profit: 8.33,
          amount: 125.50,
          timestamp: '2 min ago'
        },
        {
          id: 2,
          token: { symbol: 'WIF', logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/EKpQGSJtjMFqKZ9KQanSqYXRcF8fBopzLHYxdM65zcjm/logo.png' },
          dexPair: 'Jupiter → Raydium',
          profit: 7.15,
          amount: 89.00,
          timestamp: '5 min ago'
        },
        {
          id: 3,
          token: { symbol: 'PEPE', logo: 'https://raw.githubusercontent.com/solana-labs/token-list/main/assets/mainnet/BxzBCZJp8KFGhJwzWcQW1bwKqjYHbqKnWoaXWKR8Zxz/logo.png' },
          dexPair: 'Orca → Jupiter',
          profit: 7.06,
          amount: 70.60,
          timestamp: '8 min ago'
        }
      ]
    }
  },
  computed: {
    filteredOpportunities() {
      return this.opportunities.filter(opp => {
        if (this.filters.minProfit > 0 && opp.profitPercent < this.filters.minProfit) return false;
        if (this.filters.tokenType !== 'all' && opp.type !== this.filters.tokenType) return false;
        if (this.filters.dexPair !== 'all') {
          const dexPair = `${opp.buyDex.toLowerCase()}-${opp.sellDex.toLowerCase()}`;
          if (!dexPair.includes(this.filters.dexPair.replace('↔', '-').replace(' ', ''))) return false;
        }
        return true;
      });
    }
  },
  methods: {
    formatNumber(num) {
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
      } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
      }
      return num.toString();
    },
    refreshData() {
      this.isRefreshing = true;
      setTimeout(() => {
        this.isRefreshing = false;
        // Simulate data update
        this.stats.activeOpportunities = Math.floor(Math.random() * 20) + 40;
        this.stats.avgProfitMargin = (Math.random() * 5 + 5).toFixed(1);
      }, 1000);
    },
    toggleAutoRefresh() {
      this.autoRefresh = !this.autoRefresh;
    },
    executeArbitrage(opportunity) {
      alert(`Executing arbitrage for ${opportunity.token.symbol}\nBuy on ${opportunity.buyDex} at $${opportunity.buyPrice.toFixed(6)}\nSell on ${opportunity.sellDex} at $${opportunity.sellPrice.toFixed(6)}\nExpected profit: ${opportunity.profitPercent.toFixed(2)}%`);
    }
  },
  mounted() {
    // Auto refresh every 30 seconds if enabled
    setInterval(() => {
      if (this.autoRefresh) {
        this.refreshData();
      }
    }, 30000);
  }
}
</script> 