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

      <!-- Time Period Filter -->
      <div class="mb-6">
        <h3 class="text-lg font-semibold text-gray-900 mb-3">Time Period</h3>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="period in timePeriods"
            :key="period.value"
            @click="selectedPeriod = period.value"
            :class="[
              'px-4 py-2 rounded-full text-sm font-medium transition-all duration-200',
              selectedPeriod === period.value
                ? 'bg-blue-500 text-white shadow-md'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            ]"
          >
            {{ period.label }}
          </button>
        </div>
      </div>

      <!-- Stats Overview -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-gray-600">Total Channels</p>
              <p class="text-2xl font-bold text-gray-900">{{ filteredChannels.length }}</p>
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
                {{ totalPnl >= 0 ? '+' : '' }}{{ Number(totalPnl || 0).toFixed(2) }}%
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
                @click="loadSignalChannels"
                :disabled="loadingChannels"
                class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors disabled:opacity-50 flex items-center space-x-2"
              >
                <svg v-if="loadingChannels" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>{{ loadingChannels ? 'Loading...' : 'Refresh Channels' }}</span>
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
        
        <div v-if="filteredChannels.length === 0 && !loadingChannels" class="p-12 text-center">
          <div class="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-900 mb-2">No Solana Signal Channels Found</h3>
          <p class="text-gray-600 mb-4">No channels with Solana signals found for the selected time period. Try selecting a different time range or refresh your channels.</p>
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
            
            <!-- Channel Productivity Chart -->
            <div class="mb-6">
              <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold text-gray-900">
                  Token Performance Chart ({{ selectedPeriod.toUpperCase() }})
                </h3>
                <!-- Chart Sorting Controls -->
                <div class="flex items-center space-x-2">
                  <span class="text-sm text-gray-600">Sort by:</span>
                  <select 
                    v-model="chartSortBy" 
                    class="px-3 py-1 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="performance_asc">Current Performance (Low to High)</option>
                    <option value="performance_desc">Current Performance (High to Low)</option>
                    <option value="peak_asc">Peak Performance (Low to High)</option>
                    <option value="peak_desc">Peak Performance (High to Low)</option>
                    <option value="time_asc">Time (Oldest First)</option>
                    <option value="time_desc">Time (Newest First)</option>
                  </select>
                </div>
              </div>
              <div v-if="getFilteredContracts(selectedChannel).length === 0" class="text-center py-8">
                <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                  </svg>
                </div>
                <p class="text-gray-500">No performance data available for the selected period</p>
              </div>
              <div v-else class="bg-gray-50 p-6 rounded-lg">
                <!-- Chart Container -->
                <div class="relative pl-12">
                  <div class="flex items-end space-x-3 h-64 overflow-x-auto pb-4">
                    <div 
                      v-for="(contract, index) in sortedContracts" 
                      :key="contract.address"
                      class="flex flex-col items-center min-w-0 flex-shrink-0"
                      style="width: 100px;"
                    >
                      <!-- Dual Bar Container -->
                      <div class="relative flex flex-col justify-end h-48 mb-2">
                        <!-- Peak Performance Bar (Background/Shadow) -->
                        <div v-if="contract.max_performance !== undefined && contract.max_performance !== 'N/A'"
                          class="absolute bottom-0 w-8 rounded-t-lg opacity-40 border-2 border-dashed"
                          :style="getPeakBarStyle(contract.max_performance, sortedContracts)"
                          :title="`${contract.token} Peak: ${getPeakMultiplierText(contract.max_performance)}`"
                        >
                          <!-- Peak performance label -->
                          <div class="absolute -top-8 left-1/2 transform -translate-x-1/2 text-xs font-medium whitespace-nowrap text-blue-600">
                            {{ getPeakMultiplierText(contract.max_performance) }}
                          </div>
                        </div>
                        
                        <!-- Current Performance Bar (Foreground) -->
                        <div 
                          class="w-6 rounded-t-lg transition-all duration-300 hover:opacity-80 cursor-pointer relative z-10"
                          :style="getBarStyle(contract.performance, sortedContracts)"
                          :title="`${contract.token} Current: ${getMultiplierText(contract.performance)}`"
                        >
                          <!-- Current performance label -->
                          <div class="absolute -top-6 left-1/2 transform -translate-x-1/2 text-xs font-semibold whitespace-nowrap"
                               :class="contract.performance === 'N/A' ? 'text-gray-500' : contract.performance >= 0 ? 'text-green-700' : 'text-red-700'">
                            {{ contract.performance === 'N/A' ? 'N/A' : getMultiplierText(contract.performance) }}
                          </div>
                        </div>
                        
                        <!-- Useful Signal Indicator -->
                        <div v-if="contract.is_useful_signal" 
                             class="absolute -right-2 top-0 w-3 h-3 bg-green-500 rounded-full border-2 border-white"
                             title="Useful Signal - Reached profitable levels">
                          <div class="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-1 h-1 bg-white rounded-full"></div>
                        </div>
                      </div>
                      
                      <!-- Token name -->
                      <div class="text-xs font-medium text-gray-700 text-center truncate w-full px-1" 
                           :title="contract.token">
                        {{ contract.token.length > 10 ? contract.token.substring(0, 10) + '...' : contract.token }}
                      </div>
                      
                      <!-- Performance percentages -->
                      <div class="text-xs text-center mt-1 space-y-0.5">
                        <!-- Current Performance -->
                        <div class="text-gray-600">
                          <span class="text-gray-400">Now:</span>
                          {{ contract.performance === 'N/A' ? 'N/A' : (contract.performance >= 0 ? '+' : '') + Number(contract.performance || 0).toFixed(1) + '%' }}
                        </div>
                        <!-- Peak Performance -->
                        <div v-if="contract.max_performance !== undefined && contract.max_performance !== 'N/A'" 
                             class="text-blue-600 font-medium">
                          <span class="text-gray-400">Peak:</span>
                          {{ (contract.max_performance >= 0 ? '+' : '') + Number(contract.max_performance || 0).toFixed(1) + '%' }}
                        </div>
                      </div>
                    </div>
                  </div>
                  
                  <!-- Y-axis labels (now inside the chart area) -->
                  <div class="absolute left-0 top-0 h-48 flex flex-col justify-between text-xs text-gray-500 pr-2">
                    <span v-for="label in getYAxisLabels(sortedContracts)" 
                          :key="label" 
                          class="text-right">
                      {{ label }}
                    </span>
                  </div>
                </div>
                
                <!-- Chart Legend -->
                <div class="mt-4 flex flex-wrap gap-4 text-sm">
                  <div class="flex items-center space-x-2">
                    <div class="w-4 h-4 bg-gradient-to-r from-green-400 to-green-600 rounded"></div>
                    <span class="text-gray-600">Current Profit (>0%)</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <div class="w-4 h-4 bg-gradient-to-r from-red-400 to-red-600 rounded"></div>
                    <span class="text-gray-600">Current Loss (<0%)</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <div class="w-4 h-4 bg-gradient-to-r from-blue-300 to-blue-500 rounded opacity-40 border border-dashed border-blue-400"></div>
                    <span class="text-gray-600">Peak Performance (Historical High)</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <div class="w-3 h-3 bg-green-500 rounded-full"></div>
                    <span class="text-gray-600">Useful Signal (Reached Profit)</span>
                  </div>
                  <div class="flex items-center space-x-2">
                    <div class="w-4 h-4 bg-gradient-to-r from-gray-300 to-gray-500 rounded"></div>
                    <span class="text-gray-600">No data / N/A</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Contract Addresses Section -->
            <div class="mb-6">
              <h3 class="text-lg font-semibold text-gray-900 mb-4">
                Solana Contract Addresses ({{ selectedPeriod.toUpperCase() }})
              </h3>
              <div v-if="getFilteredContracts(selectedChannel).length === 0" class="text-center py-8">
                <div class="w-12 h-12 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-3">
                  <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                  </svg>
                </div>
                <p class="text-gray-500">No contract addresses found for the selected period</p>
              </div>
              <div v-else class="grid grid-cols-1 gap-3">
                <div v-for="contract in getFilteredContracts(selectedChannel)" :key="contract.address" 
                     class="p-4 bg-gray-50 rounded-lg border hover:bg-gray-100 transition-colors">
                  <div class="flex items-center justify-between">
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center space-x-3 mb-2">
                        <div class="w-8 h-8 bg-gradient-to-br from-purple-500 to-pink-500 rounded-lg flex items-center justify-center text-white font-bold text-sm">
                          {{ contract.token.substring(0, 2).toUpperCase() }}
                        </div>
                        <div>
                          <h4 class="font-medium text-gray-900">{{ contract.token }}</h4>
                          <p class="text-sm text-gray-500">{{ formatTimeAgo(contract.timestamp) }}</p>
                        </div>
                      </div>
                      <div class="flex items-center space-x-2 mb-2">
                        <code class="bg-white px-3 py-1 rounded border text-sm font-mono text-gray-700 flex-1 min-w-0 truncate">
                          {{ contract.address }}
                        </code>
                        <button 
                          @click="copyToClipboard(contract.address)"
                          class="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                          title="Copy address"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                          </svg>
                        </button>
                      </div>
                      <div class="flex items-center space-x-4 text-sm">
                        <span class="px-2 py-1 rounded text-xs font-medium" 
                              :class="contract.type === 'BUY' ? 'bg-green-100 text-green-800' : contract.type === 'SELL' ? 'bg-red-100 text-red-800' : 'bg-blue-100 text-blue-800'">
                          {{ contract.type || 'SIGNAL' }}
                        </span>
                        <div class="space-y-1">
                          <!-- Current Performance -->
                          <div>
                            <span class="text-xs text-gray-500">Current:</span>
                            <span v-if="contract.performance !== undefined && contract.performance !== null && contract.performance !== 'N/A'" class="font-medium ml-1" 
                                  :class="contract.performance >= 0 ? 'text-green-600' : 'text-red-600'">
                              {{ contract.performance >= 0 ? '+' : '' }}{{ Number(contract.performance || 0).toFixed(1) }}%
                            </span>
                            <span v-else class="font-medium text-gray-500 ml-1">N/A</span>
                          </div>
                          
                          <!-- Peak Performance -->
                          <div v-if="contract.max_performance !== undefined && contract.max_performance !== 'N/A'">
                            <span class="text-xs text-gray-500">Peak:</span>
                            <span class="font-medium ml-1" 
                                  :class="contract.max_performance >= 0 ? 'text-green-600' : 'text-red-600'">
                              {{ contract.max_performance >= 0 ? '+' : '' }}{{ Number(contract.max_performance || 0).toFixed(1) }}%
                            </span>
                            <span v-if="contract.is_useful_signal" class="ml-1 text-xs bg-green-100 text-green-800 px-1 rounded">✓ Useful</span>
                          </div>
                          
                          <!-- Optimal Exit -->
                          <div v-if="contract.optimal_exit !== undefined && contract.optimal_exit !== 'N/A' && contract.optimal_exit !== contract.performance">
                            <span class="text-xs text-gray-500">Best Exit:</span>
                            <span class="font-medium ml-1" 
                                  :class="contract.optimal_exit >= 0 ? 'text-green-600' : 'text-red-600'">
                              {{ contract.optimal_exit >= 0 ? '+' : '' }}{{ Number(contract.optimal_exit || 0).toFixed(1) }}%
                            </span>
                          </div>
                        </div>
                        <span v-if="contract.price" class="text-gray-600">
                          ${{ contract.price }}
                        </span>
                      </div>
                    </div>
                    <div class="ml-4">
                      <a 
                        :href="getSolscanLink(contract.address)" 
                        target="_blank" 
                        rel="noopener noreferrer"
                        class="inline-flex items-center px-3 py-1 bg-purple-100 text-purple-700 rounded-lg hover:bg-purple-200 transition-colors text-sm font-medium"
                      >
                        <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"></path>
                        </svg>
                        Solscan
                      </a>
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
      selectedChannel: null,
      sortBy: 'winRate',
      showUserDropdown: false,
      // Real channels data fetched from API
      signalChannels: [],
      loadingChannels: false,
      // Time period filtering
      selectedPeriod: '24h',
      timePeriods: [
        { label: '24H', value: '24h' },
        { label: '3D', value: '3d' },
        { label: '7D', value: '7d' },
        { label: '1M', value: '1m' }
      ],
      // Chart sorting
      chartSortBy: 'performance_asc'
    }
  },
  emits: ['connect', '2fa-required', 'logout'],
  computed: {
    filteredChannels() {
      // Filter channels based on selected time period and Solana signals
      return this.signalChannels.filter(channel => {
        if (!channel.recentSignals || channel.recentSignals.length === 0) return false
        
        const periodHours = this.getPeriodHours(this.selectedPeriod)
        const cutoffTime = new Date(Date.now() - periodHours * 60 * 60 * 1000)
        
        // Check if channel has Solana signals in the selected period
        return channel.recentSignals.some(signal => {
          const signalTime = new Date(signal.signal_time || signal.timestamp)
          const isInPeriod = signalTime > cutoffTime
          const isSolana = signal.blockchain === 'solana' || !signal.blockchain // Default to solana if no blockchain specified
          return isInPeriod && isSolana
        })
      })
    },
    sortedChannels() {
      const channels = [...this.filteredChannels]
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
      return this.filteredChannels.reduce((sum, channel) => sum + (channel.signalCount || 0), 0)
    },
    overallWinRate() {
      if (this.filteredChannels.length === 0) return 0
      const totalWinRate = this.filteredChannels.reduce((sum, channel) => sum + (channel.winRate || 0), 0)
      return (totalWinRate / this.filteredChannels.length).toFixed(1)
    },
    totalPnl() {
      return this.filteredChannels.reduce((sum, channel) => sum + (channel.pnl || 0), 0)
    },
    
    // Computed property for sorted contracts to avoid multiple calls and ensure consistency
    sortedContracts() {
      if (!this.selectedChannel) return []
      return this.getSortedContracts(this.selectedChannel)
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
          this.signalChannels = []
        }
      }
    },
    userInfo: {
      handler(newVal) {
        if (newVal && newVal.id) {
          this.loadSignalChannels()
        }
      }
    },
    selectedPeriod() {
      if (this.isConnected && this.userInfo) {
        this.loadSignalChannels()
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
    
    async loadSignalChannels() {
      if (!this.isConnected || !this.userInfo) return
      
      this.loadingChannels = true
      try {
        const response = await fetch(`${API_BASE_URL}/api/telegram/signal-channels?period=${this.selectedPeriod}&user_id=${this.userInfo.id}`)
        const data = await response.json()
        
        console.log('API Response:', data) // Debug log
        
        if (data.success && data.channels) {
          // Convert signal channels data to the format expected by the UI
          this.signalChannels = data.channels.map(signalChannel => ({
            id: signalChannel.id,
            title: signalChannel.title,
            username: signalChannel.username,
            type: signalChannel.type || 'channel',
            members: signalChannel.members || 0,
            description: signalChannel.description || 'Solana Signal Channel',
            signalCount: signalChannel.signalCount || 0,
            winRate: signalChannel.successRate || signalChannel.winRate || 0,
            pnl: signalChannel.totalPnl || this.calculatePnlFromSignals(signalChannel.contractAddresses),
            lastSignal: signalChannel.lastActivity,
            recentSignals: signalChannel.contractAddresses || [], // Map contractAddresses to recentSignals
            status: signalChannel.status || 'active'
          }))
          
          console.log('Loaded signal channels:', this.signalChannels)
        } else {
          console.log('No channels or success=false:', data)
          this.signalChannels = []
        }
      } catch (error) {
        console.error('Failed to load signal channels:', error)
        this.signalChannels = []
      } finally {
        this.loadingChannels = false
      }
    },

    calculatePnlFromSignals(signals) {
      if (!signals || signals.length === 0) return 0
      
      let totalPnl = 0
      let validSignals = 0
      
      signals.forEach(signal => {
        const performance = signal.performance
        if (performance !== 'N/A' && performance !== null && performance !== undefined) {
          totalPnl += Number(performance)
          validSignals++
        }
      })
      
      return validSignals > 0 ? totalPnl / validSignals : 0
    },

    getPeriodHours(period) {
      switch (period) {
        case '24h': return 24
        case '3d': return 72
        case '7d': return 168
        case '1m': return 720
        default: return 24
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
    
    formatTimeAgo(timestamp) {
      const now = new Date()
      const diff = Math.floor((now - new Date(timestamp)) / 1000)
      
      if (diff < 60) {
        return 'Just now'
      } else if (diff < 120) {
        return '1 minute ago'
      } else if (diff < 3600) {
        return Math.floor(diff / 60) + ' minutes ago'
      } else if (diff < 7200) {
        return '1 hour ago'
      } else if (diff < 86400) {
        return Math.floor(diff / 3600) + ' hours ago'
      } else if (diff < 172800) {
        return '1 day ago'
      } else {
        return Math.floor(diff / 86400) + ' days ago'
      }
    },
    
    getFilteredContracts(channel) {
      if (!channel.recentSignals || channel.recentSignals.length === 0) return []
      
      const periodHours = this.getPeriodHours(this.selectedPeriod)
      const cutoffTime = new Date(Date.now() - periodHours * 60 * 60 * 1000)
      
      return channel.recentSignals.filter(signal => {
        const signalTime = new Date(signal.signal_time || signal.timestamp)
        const isInPeriod = signalTime > cutoffTime
        const isSolana = signal.blockchain === 'solana' || !signal.blockchain // Default to solana if not specified
        return isInPeriod && isSolana
      })
    },
    
    getSortedContracts(channel) {
      const contracts = this.getFilteredContracts(channel)
      if (!contracts || contracts.length === 0) return []
      
      const sortedContracts = [...contracts]
      
      switch (this.chartSortBy) {
        case 'performance_asc':
          return sortedContracts.sort((a, b) => {
            const aPerf = a.performance === 'N/A' ? -Infinity : (a.performance || 0)
            const bPerf = b.performance === 'N/A' ? -Infinity : (b.performance || 0)
            return aPerf - bPerf
          })
        case 'performance_desc':
          return sortedContracts.sort((a, b) => {
            const aPerf = a.performance === 'N/A' ? -Infinity : (a.performance || 0)
            const bPerf = b.performance === 'N/A' ? -Infinity : (b.performance || 0)
            return bPerf - aPerf
          })
        case 'peak_asc':
          return sortedContracts.sort((a, b) => {
            const aPeak = a.max_performance === 'N/A' || a.max_performance === undefined ? -Infinity : (a.max_performance || 0)
            const bPeak = b.max_performance === 'N/A' || b.max_performance === undefined ? -Infinity : (b.max_performance || 0)
            return aPeak - bPeak
          })
        case 'peak_desc':
          return sortedContracts.sort((a, b) => {
            const aPeak = a.max_performance === 'N/A' || a.max_performance === undefined ? -Infinity : (a.max_performance || 0)
            const bPeak = b.max_performance === 'N/A' || b.max_performance === undefined ? -Infinity : (b.max_performance || 0)
            return bPeak - aPeak
          })
        case 'time_asc':
          return sortedContracts.sort((a, b) => {
            const timeA = new Date(a.timestamp).getTime()
            const timeB = new Date(b.timestamp).getTime()
            
            // Handle invalid timestamps
            if (isNaN(timeA) || isNaN(timeB)) {
              return 0
            }
            
            return timeA - timeB // Oldest first
          })
        case 'time_desc':
          return sortedContracts.sort((a, b) => {
            const timeA = new Date(a.timestamp).getTime()
            const timeB = new Date(b.timestamp).getTime()
            
            // Handle invalid timestamps
            if (isNaN(timeA) || isNaN(timeB)) {
              return 0
            }
            
            return timeB - timeA // Newest first
          })
        default:
          return sortedContracts
      }
    },
    
    getSolscanLink(address) {
      return `https://solscan.io/account/${address}`
    },
    
    copyToClipboard(text) {
      const tempInput = document.createElement('input')
      tempInput.value = text
      document.body.appendChild(tempInput)
      tempInput.select()
      document.execCommand('copy')
      document.body.removeChild(tempInput)
    },
    
    // Chart helper methods
    getMultiplierText(performance) {
      if (performance === null || performance === undefined || performance === 'N/A') return 'N/A'
      
      const multiplier = (100 + performance) / 100
      if (multiplier >= 100) return 'x100+'
      if (multiplier >= 10) return `x${multiplier.toFixed(0)}`
      if (multiplier >= 1) return `x${multiplier.toFixed(2)}`
      return `x${multiplier.toFixed(3)}`
    },
    
    getBarStyle(performance, allContracts) {
      if (performance === null || performance === undefined || performance === 'N/A') {
        return {
          height: '8px',
          background: 'linear-gradient(to top, #9CA3AF, #6B7280)',
          minHeight: '8px'
        }
      }
      
      // Find the max and min performance to scale dynamically
      const maxPerformance = this.getMaxPerformance(allContracts)
      const minPerformance = this.getMinPerformance(allContracts)
      const range = maxPerformance - minPerformance
      
      // Calculate height based on relative performance within the data range
      let heightPercent = 0
      
      if (range === 0) {
        // All values are the same
        heightPercent = 50
      } else {
        // Scale performance to 10-90% range for better visibility
        const normalizedPerf = (performance - minPerformance) / range
        heightPercent = 10 + (normalizedPerf * 80) // Maps to 10-90% range
      }
      
      // Ensure minimum visibility
      heightPercent = Math.max(8, Math.min(95, heightPercent))
      const height = `${heightPercent}%`
      
      // Color based on performance
      let background
      if (performance > 0) {
        // Green gradient for profits
        background = `linear-gradient(to top, #10B981, #059669)`
      } else if (performance < 0) {
        // Red gradient for losses
        background = `linear-gradient(to top, #EF4444, #DC2626)`
      } else {
        // Gray for no change
        background = 'linear-gradient(to top, #9CA3AF, #6B7280)'
      }
      
      return {
        height,
        background,
        minHeight: '8px'
      }
    },
    
    getMaxPerformance(contracts) {
      if (!contracts || contracts.length === 0) return 100
      const performances = contracts
        .map(c => c.performance)
        .filter(p => p !== 'N/A' && p !== null && p !== undefined)
        .map(p => Number(p) || 0)
      return performances.length > 0 ? Math.max(...performances) : 100
    },
    
    getMinPerformance(contracts) {
      if (!contracts || contracts.length === 0) return -50
      const performances = contracts
        .map(c => c.performance)
        .filter(p => p !== 'N/A' && p !== null && p !== undefined)
        .map(p => Number(p) || 0)
      return performances.length > 0 ? Math.min(...performances) : -50
    },
    
    getYAxisLabels(contracts) {
      if (!contracts || contracts.length === 0) {
        return ['x2.00', 'x1.50', 'x1.00', 'x0.75', 'x0.50']
      }
      
      const maxPerformance = this.getMaxPerformance(contracts)
      const minPerformance = this.getMinPerformance(contracts)
      const maxPeakPerformance = this.getMaxPeakPerformance(contracts)
      const minPeakPerformance = this.getMinPeakPerformance(contracts)
      
      // Use the maximum range from both current and peak performance
      const overallMax = Math.max(maxPerformance, maxPeakPerformance)
      const overallMin = Math.min(minPerformance, minPeakPerformance)
      
      const maxMultiplier = (100 + overallMax) / 100
      const minMultiplier = (100 + overallMin) / 100
      
      // Generate 5 evenly spaced labels from max to min
      const labels = []
      const range = maxMultiplier - minMultiplier
      
      // Ensure minimum range for better display
      const effectiveRange = Math.max(range, 0.1)
      const effectiveMax = maxMultiplier
      const effectiveMin = effectiveMax - effectiveRange
      
      for (let i = 0; i < 5; i++) {
        const value = effectiveMax - (effectiveRange * i / 4)
        
        if (value >= 100) {
          labels.push('x100+')
        } else if (value >= 10) {
          labels.push(`x${value.toFixed(0)}`)
        } else if (value >= 1) {
          labels.push(`x${value.toFixed(2)}`)
        } else if (value > 0) {
          labels.push(`x${value.toFixed(3)}`)
        } else {
          labels.push('x0.000')
        }
      }
      
      return labels
    },
    
    getPeakBarStyle(peakPerformance, allContracts) {
      if (peakPerformance === null || peakPerformance === undefined || peakPerformance === 'N/A') {
        return {
          height: '8px',
          background: 'linear-gradient(to top, #9CA3AF, #6B7280)',
          minHeight: '8px'
        }
      }
      
      // Find the max and min peak performance to scale dynamically
      const maxPeak = this.getMaxPeakPerformance(allContracts)
      const minPeak = this.getMinPeakPerformance(allContracts)
      const peakRange = maxPeak - minPeak
      
      // Calculate height based on relative peak performance within the data range
      let peakHeightPercent = 0
      
      if (peakRange === 0) {
        // All values are the same
        peakHeightPercent = 50
      } else {
        // Scale peak performance to 10-90% range for better visibility
        const normalizedPeak = (peakPerformance - minPeak) / peakRange
        peakHeightPercent = 10 + (normalizedPeak * 80) // Maps to 10-90% range
      }
      
      // Ensure minimum visibility
      peakHeightPercent = Math.max(8, Math.min(95, peakHeightPercent))
      const peakHeight = `${peakHeightPercent}%`
      
      // Color based on peak performance
      let peakBackground
      if (peakPerformance > 0) {
        // Green gradient for profitable peak
        peakBackground = `linear-gradient(to top, #10B981, #059669)`
      } else if (peakPerformance < 0) {
        // Red gradient for loss peak
        peakBackground = `linear-gradient(to top, #EF4444, #DC2626)`
      } else {
        // Gray for no change peak
        peakBackground = 'linear-gradient(to top, #9CA3AF, #6B7280)'
      }
      
      return {
        height: peakHeight,
        background: peakBackground,
        minHeight: '8px'
      }
    },
    
    getMaxPeakPerformance(contracts) {
      if (!contracts || contracts.length === 0) return 100
      const peakPerformances = contracts
        .map(c => c.max_performance)
        .filter(p => p !== 'N/A' && p !== null && p !== undefined)
        .map(p => Number(p) || 0)
      return peakPerformances.length > 0 ? Math.max(...peakPerformances) : 100
    },
    
    getMinPeakPerformance(contracts) {
      if (!contracts || contracts.length === 0) return -50
      const peakPerformances = contracts
        .map(c => c.max_performance)
        .filter(p => p !== 'N/A' && p !== null && p !== undefined)
        .map(p => Number(p) || 0)
      return peakPerformances.length > 0 ? Math.min(...peakPerformances) : -50
    },
    
    getPeakMultiplierText(peakPerformance) {
      if (peakPerformance === null || peakPerformance === undefined || peakPerformance === 'N/A') return 'N/A'
      
      const peakMultiplier = (100 + peakPerformance) / 100
      if (peakMultiplier >= 100) return 'x100+'
      if (peakMultiplier >= 10) return `x${peakMultiplier.toFixed(0)}`
      if (peakMultiplier >= 1) return `x${peakMultiplier.toFixed(2)}`
      return `x${peakMultiplier.toFixed(3)}`
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