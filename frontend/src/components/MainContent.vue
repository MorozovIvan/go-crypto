<template>
  <main class="flex-1 p-6 bg-gray-50 min-h-screen">
    <div v-if="!isConnected" class="text-center">
      <TelegramConnectForm
        @connected="handleConnected"
        @2fa-required="handle2FARequired"
      />
    </div>
    <div v-else class="max-w-7xl mx-auto">
      <!-- Header Section -->
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">Signal Analytics</h1>
          <p class="text-gray-600 mt-1">Track and analyze your Telegram trading signals</p>
        </div>
        <div class="flex items-center space-x-4">
          <!-- User Profile Dropdown -->
          <div v-if="userInfo" class="relative">
            <button
              @click="showUserDropdown = !showUserDropdown"
              class="flex items-center space-x-3 bg-white px-4 py-2 rounded-lg shadow-sm hover:shadow-md transition-shadow border border-gray-200"
            >
              <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white font-bold text-sm">
                {{ getUserInitials(userInfo) }}
              </div>
              <div class="text-sm text-left">
                <div class="font-medium text-gray-900">{{ getUserDisplayName(userInfo) }}</div>
                <div class="text-gray-500">@{{ userInfo.username || 'telegram' }}</div>
              </div>
              <svg class="w-4 h-4 text-gray-400 transition-transform" :class="{ 'rotate-180': showUserDropdown }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
              </svg>
            </button>
            
            <!-- Dropdown Menu -->
            <div v-if="showUserDropdown" class="absolute right-0 mt-2 w-64 bg-white rounded-lg shadow-lg border border-gray-200 py-2 z-50">
              <div class="px-4 py-3 border-b border-gray-100">
                <div class="flex items-center space-x-3">
                  <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-full flex items-center justify-center text-white font-bold">
                    {{ getUserInitials(userInfo) }}
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ getUserDisplayName(userInfo) }}</div>
                    <div class="text-sm text-gray-500">@{{ userInfo.username || 'telegram' }}</div>
                    <div v-if="userInfo.phone" class="text-xs text-gray-400">{{ userInfo.phone }}</div>
                  </div>
                </div>
              </div>
              
              <div class="py-1">
                <button
                  @click="refreshUserInfo"
                  class="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 flex items-center space-x-2"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
                  </svg>
                  <span>Refresh Profile</span>
                </button>
                
                <button
                  @click="handleLogout"
                  class="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 flex items-center space-x-2"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"></path>
                  </svg>
                  <span>Logout</span>
                </button>
              </div>
            </div>
          </div>
          
          <!-- Loading state when user info is not available -->
          <div v-else class="flex items-center space-x-3 bg-white px-4 py-2 rounded-lg shadow-sm border border-gray-200">
            <div class="w-10 h-10 bg-gray-200 rounded-full animate-pulse"></div>
            <div class="text-sm">
              <div class="w-20 h-4 bg-gray-200 rounded animate-pulse mb-1"></div>
              <div class="w-16 h-3 bg-gray-200 rounded animate-pulse"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Stats Overview -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-600">Total Channels</p>
              <p class="text-2xl font-bold text-gray-900">{{ signalChannels.length }}</p>
            </div>
            <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-600">Total Signals</p>
              <p class="text-2xl font-bold text-gray-900">{{ totalSignals }}</p>
            </div>
            <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-600">Win Rate</p>
              <p class="text-2xl font-bold text-green-600">{{ overallWinRate }}%</p>
            </div>
            <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </div>
          </div>
        </div>
        
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-600">Total PNL</p>
              <p class="text-2xl font-bold" :class="totalPnl >= 0 ? 'text-green-600' : 'text-red-600'">
                {{ totalPnl >= 0 ? '+' : '' }}{{ totalPnl.toFixed(2) }}%
              </p>
            </div>
            <div class="w-12 h-12 rounded-lg flex items-center justify-center" :class="totalPnl >= 0 ? 'bg-green-100' : 'bg-red-100'">
              <svg class="w-6 h-6" :class="totalPnl >= 0 ? 'text-green-600' : 'text-red-600'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" :d="totalPnl >= 0 ? 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6' : 'M13 17h8m0 0V9m0 8l-8-8-4 4-6-6'"></path>
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Channel Analysis Section -->
      <div class="bg-white rounded-xl shadow-sm border border-gray-100 mb-8">
        <div class="p-6 border-b border-gray-100">
          <div class="flex justify-between items-center">
            <h2 class="text-xl font-semibold text-gray-900">Signal Channels</h2>
            <div class="flex items-center space-x-3">
              <button
                @click="loadGroups"
                :disabled="loadingGroups"
                class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 flex items-center space-x-2"
              >
                <svg v-if="loadingGroups" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>{{ loadingGroups ? 'Loading...' : 'Refresh Channels' }}</span>
              </button>
              <select v-model="sortBy" class="px-3 py-2 border border-gray-300 rounded-lg text-sm">
                <option value="winRate">Sort by Win Rate</option>
                <option value="pnl">Sort by PNL</option>
                <option value="signals">Sort by Signals</option>
                <option value="name">Sort by Name</option>
              </select>
            </div>
          </div>
        </div>
        
        <div v-if="signalChannels.length === 0 && !loadingGroups" class="p-12 text-center">
          <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-900 mb-2">No Signal Channels Found</h3>
          <p class="text-gray-600 mb-4">Click "Refresh Channels" to load your Telegram channels and start analyzing trading signals.</p>
        </div>
        
        <div v-else class="divide-y divide-gray-100">
          <div 
            v-for="channel in sortedChannels" 
            :key="channel.id"
            class="p-6 hover:bg-gray-50 transition-colors cursor-pointer"
            @click="selectChannel(channel)"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4">
                <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold">
                  {{ getChannelInitials(channel.title) }}
                </div>
                <div>
                  <h3 class="font-semibold text-gray-900">{{ channel.title }}</h3>
                  <p v-if="channel.username" class="text-blue-600 text-sm">@{{ channel.username }}</p>
                  <div class="flex items-center space-x-4 mt-1 text-sm text-gray-500">
                    <span class="capitalize">{{ channel.type }}</span>
                    <span v-if="channel.members">{{ formatNumber(channel.members) }} members</span>
                  </div>
                </div>
              </div>
              
              <div class="flex items-center space-x-6">
                <!-- Signals Count -->
                <div class="text-center">
                  <div class="text-lg font-bold text-gray-900">{{ channel.signalCount || 0 }}</div>
                  <div class="text-xs text-gray-500">Signals</div>
                </div>
                
                <!-- Win Rate -->
                <div class="text-center">
                  <div class="text-lg font-bold" :class="getWinRateColor(channel.winRate)">
                    {{ channel.winRate || 0 }}%
                  </div>
                  <div class="text-xs text-gray-500">Win Rate</div>
                </div>
                
                <!-- PNL -->
                <div class="text-center">
                  <div class="text-lg font-bold" :class="channel.pnl >= 0 ? 'text-green-600' : 'text-red-600'">
                    {{ channel.pnl >= 0 ? '+' : '' }}{{ (channel.pnl || 0).toFixed(1) }}%
                  </div>
                  <div class="text-xs text-gray-500">PNL</div>
                </div>
                
                <!-- Performance Badge -->
                <div class="flex flex-col items-end">
                  <span class="px-3 py-1 rounded-full text-xs font-medium" :class="getPerformanceBadge(channel)">
                    {{ getPerformanceLabel(channel) }}
                  </span>
                  <div class="text-xs text-gray-500 mt-1">{{ getLastSignalTime(channel) }}</div>
                </div>
                
                <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Channel Detail Modal -->
      <div v-if="selectedChannel" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="closeChannelDetail">
        <div class="bg-white rounded-xl shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-hidden" @click.stop>
          <div class="p-6 border-b border-gray-100">
            <div class="flex justify-between items-center">
              <div class="flex items-center space-x-4">
                <div class="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center text-white font-bold text-xl">
                  {{ getChannelInitials(selectedChannel.title) }}
                </div>
                <div>
                  <h2 class="text-2xl font-bold text-gray-900">{{ selectedChannel.title }}</h2>
                  <p v-if="selectedChannel.username" class="text-blue-600">@{{ selectedChannel.username }}</p>
                  <p v-if="selectedChannel.description" class="text-gray-600 mt-1">{{ selectedChannel.description }}</p>
                </div>
              </div>
              <button @click="closeChannelDetail" class="text-gray-400 hover:text-gray-600">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
              </button>
            </div>
          </div>
          
          <div class="p-6 overflow-y-auto max-h-[calc(90vh-120px)]">
            <!-- Channel Stats -->
            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
              <div class="bg-gray-50 p-4 rounded-lg">
                <div class="text-2xl font-bold text-gray-900">{{ selectedChannel.signalCount || 0 }}</div>
                <div class="text-sm text-gray-600">Total Signals</div>
              </div>
              <div class="bg-gray-50 p-4 rounded-lg">
                <div class="text-2xl font-bold" :class="getWinRateColor(selectedChannel.winRate)">
                  {{ selectedChannel.winRate || 0 }}%
                </div>
                <div class="text-sm text-gray-600">Win Rate</div>
              </div>
              <div class="bg-gray-50 p-4 rounded-lg">
                <div class="text-2xl font-bold" :class="selectedChannel.pnl >= 0 ? 'text-green-600' : 'text-red-600'">
                  {{ selectedChannel.pnl >= 0 ? '+' : '' }}{{ (selectedChannel.pnl || 0).toFixed(2) }}%
                </div>
                <div class="text-sm text-gray-600">Total PNL</div>
              </div>
              <div class="bg-gray-50 p-4 rounded-lg">
                <div class="text-2xl font-bold text-gray-900">{{ formatNumber(selectedChannel.members || 0) }}</div>
                <div class="text-sm text-gray-600">Members</div>
              </div>
            </div>
            
            <!-- Recent Signals -->
            <div class="mb-6">
              <h3 class="text-lg font-semibold text-gray-900 mb-4">Recent Signals</h3>
              <div class="space-y-3">
                <div v-for="signal in getRecentSignals(selectedChannel)" :key="signal.id" 
                     class="p-4 border border-gray-200 rounded-lg">
                  <div class="flex justify-between items-start">
                    <div>
                      <div class="flex items-center space-x-2 mb-2">
                        <span class="font-medium text-gray-900">{{ signal.token }}</span>
                        <span class="px-2 py-1 rounded text-xs font-medium" 
                              :class="signal.type === 'BUY' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'">
                          {{ signal.type }}
                        </span>
                      </div>
                      <div class="text-sm text-gray-600">
                        Entry: ${{ signal.entry }} | Target: ${{ signal.target }}
                      </div>
                    </div>
                    <div class="text-right">
                      <div class="font-medium" :class="signal.pnl >= 0 ? 'text-green-600' : 'text-red-600'">
                        {{ signal.pnl >= 0 ? '+' : '' }}{{ signal.pnl.toFixed(2) }}%
                      </div>
                      <div class="text-xs text-gray-500">{{ signal.date }}</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>

<script>
import TelegramConnectForm from './TelegramConnectForm.vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export default {
  name: 'MainContent',
  components: {
    TelegramConnectForm
  },
  props: {
    isConnected: {
      type: Boolean,
      default: false
    },
    userId: {
      type: [Number, String],
      default: null
    }
  },
  data() {
    return {
      userInfo: null,
      groups: [],
      loadingGroups: false,
      selectedChannel: null,
      sortBy: 'winRate',
      showUserDropdown: false,
      // Real channels data fetched from API
      signalChannels: [],
      loadingChannels: false
    }
  },
  emits: ['connect', '2fa-required', 'logout'],
  computed: {
    sortedChannels() {
      const channels = [...this.signalChannels]
      switch (this.sortBy) {
        case 'winRate':
          return channels.sort((a, b) => (b.winRate || 0) - (a.winRate || 0))
        case 'pnl':
          return channels.sort((a, b) => (b.pnl || 0) - (a.pnl || 0))
        case 'signals':
          return channels.sort((a, b) => (b.signalCount || 0) - (a.signalCount || 0))
        case 'name':
          return channels.sort((a, b) => a.title.localeCompare(b.title))
        default:
          return channels
      }
    },
    totalSignals() {
      return this.signalChannels.reduce((sum, channel) => sum + (channel.signalCount || 0), 0)
    },
    overallWinRate() {
      if (this.signalChannels.length === 0) return 0
      const totalWinRate = this.signalChannels.reduce((sum, channel) => sum + (channel.winRate || 0), 0)
      return (totalWinRate / this.signalChannels.length).toFixed(1)
    },
    totalPnl() {
      return this.signalChannels.reduce((sum, channel) => sum + (channel.pnl || 0), 0)
    }
  },
  watch: {
    isConnected: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.loadUserInfo()
        } else {
          this.userInfo = null
          this.groups = []
          this.signalChannels = []
        }
      }
    },
    userInfo: {
      handler(newVal) {
        if (newVal && newVal.id) {
          this.loadGroups()
        }
      }
    }
  },
  methods: {
    async loadUserInfo() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/telegram/status`)
        const data = await response.json()
        
        if (data.status && data.status.authenticated && data.status.user) {
          this.userInfo = data.status.user
        } else {
          this.userInfo = null
        }
      } catch (error) {
        console.error('Failed to load user info:', error)
        this.userInfo = null
      }
    },
    
    async loadGroups() {
      if (!this.userInfo || !this.userInfo.id) return
      
      this.loadingGroups = true
      this.loadingChannels = true
      try {
        const response = await fetch(`${API_BASE_URL}/api/telegram/groups?user_id=${this.userInfo.id}`)
        const data = await response.json()
        
        if (data.groups) {
          this.groups = data.groups
          
          // Transform real Telegram channels/groups into signal channel format
          this.signalChannels = data.groups.map(group => ({
            id: group.id,
            title: group.title,
            username: group.username || '',
            type: group.type,
            members: group.members || 0,
            description: group.description || '',
            // For now, we'll use placeholder values for trading metrics
            // In a real implementation, you would analyze messages to calculate these
            signalCount: Math.floor(Math.random() * 200) + 50,
            winRate: Math.floor(Math.random() * 40) + 60, // 60-100%
            pnl: Math.floor(Math.random() * 300) + 50, // 50-350%
            lastSignal: this.getRandomTimeAgo(),
            recentSignals: this.generateMockSignals()
          }))
          
          console.log('Loaded real channels:', this.signalChannels)
        }
      } catch (error) {
        console.error('Failed to load groups:', error)
        // Fallback to empty array on error
        this.signalChannels = []
      } finally {
        this.loadingGroups = false
        this.loadingChannels = false
      }
    },
    
    selectChannel(channel) {
      this.selectedChannel = channel
    },
    
    closeChannelDetail() {
      this.selectedChannel = null
    },
    
    getUserDisplayName(user) {
      if (user.first_name && user.last_name) {
        return `${user.first_name} ${user.last_name}`
      } else if (user.first_name) {
        return user.first_name
      } else if (user.username) {
        return user.username
      } else {
        return 'Telegram User'
      }
    },
    
    getUserInitials(user) {
      if (user.first_name && user.last_name) {
        return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
      } else if (user.first_name) {
        return user.first_name[0].toUpperCase()
      } else if (user.username) {
        return user.username[0].toUpperCase()
      } else {
        return 'T'
      }
    },
    
    getChannelInitials(title) {
      const words = title.split(' ').filter(word => /^[A-Za-z]/.test(word))
      if (words.length >= 2) {
        return `${words[0][0]}${words[1][0]}`.toUpperCase()
      } else if (words.length === 1) {
        return words[0].substring(0, 2).toUpperCase()
      } else {
        return title.substring(0, 2).toUpperCase()
      }
    },
    
    getWinRateColor(winRate) {
      if (winRate >= 80) return 'text-green-600'
      if (winRate >= 60) return 'text-yellow-600'
      return 'text-red-600'
    },
    
    getPerformanceBadge(channel) {
      const winRate = channel.winRate || 0
      const pnl = channel.pnl || 0
      
      if (winRate >= 80 && pnl >= 200) return 'bg-green-100 text-green-800'
      if (winRate >= 70 && pnl >= 100) return 'bg-blue-100 text-blue-800'
      if (winRate >= 60 && pnl >= 50) return 'bg-yellow-100 text-yellow-800'
      return 'bg-gray-100 text-gray-800'
    },
    
    getPerformanceLabel(channel) {
      const winRate = channel.winRate || 0
      const pnl = channel.pnl || 0
      
      if (winRate >= 80 && pnl >= 200) return 'Elite'
      if (winRate >= 70 && pnl >= 100) return 'Premium'
      if (winRate >= 60 && pnl >= 50) return 'Good'
      return 'Average'
    },
    
    getLastSignalTime(channel) {
      return channel.lastSignal || 'No recent signals'
    },
    
    getRecentSignals(channel) {
      return channel.recentSignals || []
    },
    
    formatNumber(num) {
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M'
      } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
      }
      return num.toString()
    },
    
    handleConnected() {
      this.$emit('connect')
    },
    
    handle2FARequired(data) {
      this.$emit('2fa-required', data)
    },
    
    handleLogout() {
      this.$emit('logout')
    },
    
    refreshUserInfo() {
      this.showUserDropdown = false
      this.loadUserInfo()
    },
    
    getRandomTimeAgo() {
      const times = ['30 minutes ago', '1 hour ago', '2 hours ago', '4 hours ago', '6 hours ago', '1 day ago']
      return times[Math.floor(Math.random() * times.length)]
    },
    
    generateMockSignals() {
      const tokens = ['BTC', 'ETH', 'SOL', 'BONK', 'WIF', 'PEPE', 'MATIC', 'AVAX', 'UNI', 'AAVE']
      const types = ['BUY', 'SELL']
      const signals = []
      
      for (let i = 0; i < Math.floor(Math.random() * 3) + 1; i++) {
        const token = tokens[Math.floor(Math.random() * tokens.length)]
        const type = types[Math.floor(Math.random() * types.length)]
        const entry = Math.random() * 100
        const pnl = (Math.random() - 0.5) * 100 // Can be negative or positive
        
        signals.push({
          id: i + 1,
          token,
          type,
          entry: entry.toFixed(6),
          target: (entry * (1 + Math.random() * 0.5)).toFixed(6),
          pnl: pnl.toFixed(1),
          date: this.getRandomTimeAgo()
        })
      }
      
      return signals
    },
    
    formatNumber(num) {
      if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M'
      } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K'
      }
      return num.toString()
    },
    
    getPerformanceLabel(channel) {
      const winRate = channel.winRate || 0
      const pnl = channel.pnl || 0
      
      if (winRate >= 80 && pnl >= 200) return 'Excellent'
      if (winRate >= 70 && pnl >= 100) return 'Good'
      if (winRate >= 60 && pnl >= 50) return 'Average'
      return 'Poor'
    },
    
    getLastSignalTime(channel) {
      return channel.lastSignal || 'No recent signals'
    },
    
    getRecentSignals(channel) {
      return channel.recentSignals || []
    }
  },
  
  mounted() {
    // Close dropdown when clicking outside
    document.addEventListener('click', (e) => {
      if (!this.$el.contains(e.target)) {
        this.showUserDropdown = false
      }
    })
  },
  
  beforeUnmount() {
    // Clean up event listener
    document.removeEventListener('click', this.handleClickOutside)
  }
}
</script> 