<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-indigo-900 p-6">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-4xl font-bold text-white mb-2">
            <span class="bg-gradient-to-r from-yellow-400 to-orange-500 bg-clip-text text-transparent">
              Sniping New Tokens
            </span>
          </h1>
          <p class="text-gray-300 text-lg">Pump.fun Memecoin Sniping with Social Signal Filtering</p>
        </div>
        <div class="flex items-center space-x-4">
          <div class="bg-green-500/20 px-4 py-2 rounded-lg border border-green-500/30">
            <div class="flex items-center space-x-2">
              <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span class="text-green-400 font-medium">Live Monitoring</span>
            </div>
          </div>
          <button
            @click="toggleAutoSnipe"
            :class="[
              'px-6 py-2 rounded-lg font-medium transition-all duration-200',
              autoSnipeEnabled
                ? 'bg-red-500 hover:bg-red-600 text-white'
                : 'bg-green-500 hover:bg-green-600 text-white'
            ]"
          >
            {{ autoSnipeEnabled ? 'Stop Auto-Snipe' : 'Start Auto-Snipe' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <div class="bg-gradient-to-br from-blue-500/20 to-blue-600/20 backdrop-blur-sm rounded-xl p-6 border border-blue-500/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-blue-200 text-sm font-medium">Tokens Monitored</p>
            <p class="text-3xl font-bold text-white">{{ stats.tokensMonitored }}</p>
          </div>
          <div class="p-3 bg-blue-500/30 rounded-lg">
            <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center">
          <span class="text-green-400 text-sm">+12 new today</span>
        </div>
      </div>

      <div class="bg-gradient-to-br from-green-500/20 to-green-600/20 backdrop-blur-sm rounded-xl p-6 border border-green-500/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-green-200 text-sm font-medium">Successful Snipes</p>
            <p class="text-3xl font-bold text-white">{{ stats.successfulSnipes }}</p>
          </div>
          <div class="p-3 bg-green-500/30 rounded-lg">
            <svg class="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center">
          <span class="text-green-400 text-sm">{{ stats.successRate }}% success rate</span>
        </div>
      </div>

      <div class="bg-gradient-to-br from-purple-500/20 to-purple-600/20 backdrop-blur-sm rounded-xl p-6 border border-purple-500/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-purple-200 text-sm font-medium">Total Profit</p>
            <p class="text-3xl font-bold text-white">${{ stats.totalProfit.toLocaleString() }}</p>
          </div>
          <div class="p-3 bg-purple-500/30 rounded-lg">
            <svg class="w-6 h-6 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"/>
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center">
          <span class="text-green-400 text-sm">+{{ stats.avgReturn }}% avg return</span>
        </div>
      </div>

      <div class="bg-gradient-to-br from-orange-500/20 to-orange-600/20 backdrop-blur-sm rounded-xl p-6 border border-orange-500/30">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-orange-200 text-sm font-medium">Social Mentions</p>
            <p class="text-3xl font-bold text-white">{{ stats.socialMentions }}</p>
          </div>
          <div class="p-3 bg-orange-500/30 rounded-lg">
            <svg class="w-6 h-6 text-orange-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-1l-4 4z"/>
            </svg>
          </div>
        </div>
        <div class="mt-2 flex items-center">
          <span class="text-orange-400 text-sm">{{ stats.trendingTokens }} trending</span>
        </div>
      </div>
    </div>

    <!-- Filters and Controls -->
    <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 mb-8 border border-gray-700/50">
      <h3 class="text-xl font-semibold text-white mb-4">Sniping Filters & Settings</h3>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Social Signal Threshold -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Min Social Mentions/Hour</label>
          <select v-model="filters.socialThreshold" class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white">
            <option value="100">100+ mentions</option>
            <option value="250">250+ mentions</option>
            <option value="500">500+ mentions</option>
            <option value="1000">1000+ mentions</option>
          </select>
        </div>

        <!-- Liquidity Filter -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Min Initial Liquidity</label>
          <select v-model="filters.minLiquidity" class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white">
            <option value="1000">$1,000+</option>
            <option value="5000">$5,000+</option>
            <option value="10000">$10,000+</option>
            <option value="25000">$25,000+</option>
          </select>
        </div>

        <!-- Max Buy Amount -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Max Buy Amount</label>
          <select v-model="filters.maxBuyAmount" class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white">
            <option value="100">$100</option>
            <option value="250">$250</option>
            <option value="500">$500</option>
            <option value="1000">$1,000</option>
          </select>
        </div>

        <!-- Platform Filter -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Platform Focus</label>
          <select v-model="filters.platform" class="w-full bg-gray-700 border border-gray-600 rounded-lg px-3 py-2 text-white">
            <option value="all">All Platforms</option>
            <option value="pump.fun">Pump.fun Only</option>
            <option value="raydium">Raydium</option>
            <option value="orca">Orca</option>
          </select>
        </div>
      </div>

      <!-- Advanced Filters -->
      <div class="mt-4 pt-4 border-t border-gray-700">
        <div class="flex flex-wrap gap-4">
          <label class="flex items-center">
            <input type="checkbox" v-model="filters.requireInfluencerMention" class="mr-2 text-blue-500">
            <span class="text-gray-300">Require Influencer Mention</span>
          </label>
          <label class="flex items-center">
            <input type="checkbox" v-model="filters.avoidRugPulls" class="mr-2 text-blue-500">
            <span class="text-gray-300">Rug Pull Protection</span>
          </label>
          <label class="flex items-center">
            <input type="checkbox" v-model="filters.onlyVerifiedDevs" class="mr-2 text-blue-500">
            <span class="text-gray-300">Verified Developers Only</span>
          </label>
        </div>
      </div>
    </div>

    <!-- Live Token Feed -->
    <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl border border-gray-700/50 overflow-hidden">
      <div class="p-6 border-b border-gray-700/50">
        <div class="flex items-center justify-between">
          <h3 class="text-xl font-semibold text-white">Live Token Feed</h3>
          <div class="flex items-center space-x-2">
            <button
              @click="refreshFeed"
              class="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors"
            >
              Refresh
            </button>
            <div class="flex items-center space-x-2 text-sm text-gray-400">
              <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span>Auto-refresh: 5s</span>
            </div>
          </div>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-700/50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Token</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Created</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Social Score</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Liquidity</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Volume</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Risk Level</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-700/50">
            <tr v-for="token in filteredTokens" :key="token.address" class="hover:bg-gray-700/30 transition-colors">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <img :src="token.logo" :alt="token.name" class="w-10 h-10 rounded-full mr-3">
                  <div>
                    <div class="text-sm font-medium text-white">{{ token.name }}</div>
                    <div class="text-sm text-gray-400">${{ token.symbol }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                {{ token.createdAt }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="flex-1 bg-gray-700 rounded-full h-2 mr-2">
                    <div 
                      class="bg-gradient-to-r from-orange-400 to-red-500 h-2 rounded-full"
                      :style="{ width: Math.min(token.socialScore / 10, 100) + '%' }"
                    ></div>
                  </div>
                  <span class="text-sm font-medium text-white">{{ token.socialScore }}</span>
                </div>
                <div class="text-xs text-gray-400 mt-1">
                  X: {{ token.xMentions }} | TG: {{ token.telegramMentions }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                ${{ token.liquidity.toLocaleString() }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-300">
                ${{ token.volume.toLocaleString() }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="[
                  'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                  token.riskLevel === 'Low' ? 'bg-green-100 text-green-800' :
                  token.riskLevel === 'Medium' ? 'bg-yellow-100 text-yellow-800' :
                  'bg-red-100 text-red-800'
                ]">
                  {{ token.riskLevel }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex space-x-2">
                  <button
                    @click="snipeToken(token)"
                    :disabled="token.sniped"
                    :class="[
                      'px-3 py-1 rounded-lg text-xs font-medium transition-colors',
                      token.sniped
                        ? 'bg-gray-600 text-gray-400 cursor-not-allowed'
                        : 'bg-green-500 hover:bg-green-600 text-white'
                    ]"
                  >
                    {{ token.sniped ? 'Sniped' : 'Snipe Now' }}
                  </button>
                  <button
                    @click="viewTokenDetails(token)"
                    class="px-3 py-1 bg-blue-500 hover:bg-blue-600 text-white rounded-lg text-xs font-medium transition-colors"
                  >
                    Details
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Recent Snipes -->
    <div class="mt-8 bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50">
      <h3 class="text-xl font-semibold text-white mb-4">Recent Snipes</h3>
      <div class="space-y-4">
        <div v-for="snipe in recentSnipes" :key="snipe.id" class="flex items-center justify-between p-4 bg-gray-700/30 rounded-lg">
          <div class="flex items-center space-x-4">
            <img :src="snipe.logo" :alt="snipe.name" class="w-12 h-12 rounded-full">
            <div>
              <div class="text-white font-medium">{{ snipe.name }} ({{ snipe.symbol }})</div>
              <div class="text-gray-400 text-sm">{{ snipe.timestamp }}</div>
            </div>
          </div>
          <div class="text-right">
            <div class="text-white font-medium">${{ snipe.buyAmount }}</div>
            <div :class="[
              'text-sm font-medium',
              snipe.currentValue > snipe.buyAmount ? 'text-green-400' : 'text-red-400'
            ]">
              {{ snipe.currentValue > snipe.buyAmount ? '+' : '' }}{{ ((snipe.currentValue - snipe.buyAmount) / snipe.buyAmount * 100).toFixed(1) }}%
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SnipingDashboard',
  data() {
    return {
      autoSnipeEnabled: false,
      stats: {
        tokensMonitored: 1247,
        successfulSnipes: 89,
        successRate: 73.2,
        totalProfit: 45280,
        avgReturn: 187,
        socialMentions: 2847,
        trendingTokens: 23
      },
      filters: {
        socialThreshold: 500,
        minLiquidity: 10000,
        maxBuyAmount: 500,
        platform: 'pump.fun',
        requireInfluencerMention: true,
        avoidRugPulls: true,
        onlyVerifiedDevs: false
      },
      tokens: [
        {
          address: '0x1234...5678',
          name: 'PumpCat',
          symbol: 'PCAT',
          logo: 'https://via.placeholder.com/40/FF6B35/FFFFFF?text=PC',
          createdAt: '2 min ago',
          socialScore: 847,
          xMentions: 623,
          telegramMentions: 224,
          liquidity: 15420,
          volume: 8750,
          riskLevel: 'Low',
          sniped: false
        },
        {
          address: '0x2345...6789',
          name: 'MoonDoge',
          symbol: 'MDOGE',
          logo: 'https://via.placeholder.com/40/4ECDC4/FFFFFF?text=MD',
          createdAt: '5 min ago',
          socialScore: 1203,
          xMentions: 891,
          telegramMentions: 312,
          liquidity: 22100,
          volume: 12400,
          riskLevel: 'Low',
          sniped: true
        },
        {
          address: '0x3456...7890',
          name: 'SolanaShiba',
          symbol: 'SSHIB',
          logo: 'https://via.placeholder.com/40/45B7D1/FFFFFF?text=SS',
          createdAt: '8 min ago',
          socialScore: 654,
          xMentions: 456,
          telegramMentions: 198,
          liquidity: 8900,
          volume: 5200,
          riskLevel: 'Medium',
          sniped: false
        },
        {
          address: '0x4567...8901',
          name: 'PumpPepe',
          symbol: 'PPEPE',
          logo: 'https://via.placeholder.com/40/96CEB4/FFFFFF?text=PP',
          createdAt: '12 min ago',
          socialScore: 1456,
          xMentions: 1124,
          telegramMentions: 332,
          liquidity: 31200,
          volume: 18700,
          riskLevel: 'Low',
          sniped: true
        },
        {
          address: '0x5678...9012',
          name: 'SafeMoon2',
          symbol: 'SAFE2',
          logo: 'https://via.placeholder.com/40/FFEAA7/000000?text=S2',
          createdAt: '15 min ago',
          socialScore: 432,
          xMentions: 298,
          telegramMentions: 134,
          liquidity: 6800,
          volume: 3100,
          riskLevel: 'High',
          sniped: false
        },
        {
          address: '0x6789...0123',
          name: 'ElonCoin',
          symbol: 'ELON',
          logo: 'https://via.placeholder.com/40/DDA0DD/000000?text=EC',
          createdAt: '18 min ago',
          socialScore: 2103,
          xMentions: 1654,
          telegramMentions: 449,
          liquidity: 45600,
          volume: 28900,
          riskLevel: 'Low',
          sniped: true
        }
      ],
      recentSnipes: [
        {
          id: 1,
          name: 'MoonDoge',
          symbol: 'MDOGE',
          logo: 'https://via.placeholder.com/48/4ECDC4/FFFFFF?text=MD',
          timestamp: '3 minutes ago',
          buyAmount: 500,
          currentValue: 847
        },
        {
          id: 2,
          name: 'PumpPepe',
          symbol: 'PPEPE',
          logo: 'https://via.placeholder.com/48/96CEB4/FFFFFF?text=PP',
          timestamp: '8 minutes ago',
          buyAmount: 250,
          currentValue: 623
        },
        {
          id: 3,
          name: 'ElonCoin',
          symbol: 'ELON',
          logo: 'https://via.placeholder.com/48/DDA0DD/000000?text=EC',
          timestamp: '15 minutes ago',
          buyAmount: 1000,
          currentValue: 1456
        }
      ]
    }
  },
  computed: {
    filteredTokens() {
      return this.tokens.filter(token => {
        if (token.socialScore < this.filters.socialThreshold) return false;
        if (token.liquidity < this.filters.minLiquidity) return false;
        if (this.filters.platform !== 'all' && token.platform !== this.filters.platform) return false;
        return true;
      });
    }
  },
  methods: {
    toggleAutoSnipe() {
      this.autoSnipeEnabled = !this.autoSnipeEnabled;
      if (this.autoSnipeEnabled) {
        this.$toast?.success('Auto-snipe enabled! Monitoring for opportunities...');
      } else {
        this.$toast?.info('Auto-snipe disabled.');
      }
    },
    refreshFeed() {
      // Simulate refresh
      this.$toast?.info('Refreshing token feed...');
    },
    snipeToken(token) {
      if (token.sniped) return;
      
      token.sniped = true;
      this.stats.successfulSnipes++;
      
      // Add to recent snipes
      this.recentSnipes.unshift({
        id: Date.now(),
        name: token.name,
        symbol: token.symbol,
        logo: token.logo,
        timestamp: 'Just now',
        buyAmount: this.filters.maxBuyAmount,
        currentValue: this.filters.maxBuyAmount
      });
      
      // Keep only last 10 snipes
      if (this.recentSnipes.length > 10) {
        this.recentSnipes = this.recentSnipes.slice(0, 10);
      }
      
      this.$toast?.success(`Successfully sniped ${token.name} for $${this.filters.maxBuyAmount}!`);
    },
    viewTokenDetails(token) {
      this.$toast?.info(`Viewing details for ${token.name}...`);
    }
  },
  mounted() {
    // Simulate real-time updates
    setInterval(() => {
      // Update social scores and volumes randomly
      this.tokens.forEach(token => {
        if (Math.random() > 0.7) {
          token.socialScore += Math.floor(Math.random() * 50);
          token.volume += Math.floor(Math.random() * 1000);
        }
      });
      
      // Update stats
      this.stats.socialMentions += Math.floor(Math.random() * 10);
    }, 5000);
  }
}
</script> 