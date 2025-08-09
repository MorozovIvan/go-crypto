<template>
	<div
		class="min-h-screen bg-gradient-to-br from-base-100 via-base-200 to-base-300 p-6"
	>
		<!-- Enhanced Controls -->
		<div class="mb-8">
			<div
				class="flex flex-col lg:flex-row justify-between items-start lg:items-center gap-4 mb-6"
			>

				<!-- Enhanced Controls -->
				<div class="flex flex-col sm:flex-row gap-4">
					<!-- Data Source Switcher -->
					<div
						class="flex items-center space-x-3 bg-base-100 p-3 rounded-xl shadow-lg border border-base-300"
					>
						<span class="text-sm font-medium text-base-content">Data:</span>
						<div class="flex items-center space-x-2">
							<label class="relative inline-flex items-center cursor-pointer">
								<input
									type="checkbox"
									v-model="useMockData"
									@change="toggleDataSource"
									class="sr-only peer"
								/>
								<div
									class="w-11 h-6 bg-base-300 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-base-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"
								></div>
							</label>
							<div class="flex flex-col text-xs">
								<span
									:class="
										useMockData
											? 'text-primary font-semibold'
											: 'text-base-content/50'
									"
									>Mock</span
								>
								<span
									:class="
										!useMockData
											? 'text-success font-semibold'
											: 'text-base-content/50'
									"
									>Live</span
								>
							</div>
						</div>
						<div
							:class="
								useMockData ? 'badge badge-primary' : 'badge badge-success'
							"
						>
							{{ useMockData ? 'Enhanced Mock' : 'Real APIs' }}
						</div>
					</div>

					<!-- Auto-refresh Toggle -->
					<div
						class="flex items-center space-x-3 bg-base-100 p-3 rounded-xl shadow-lg border border-base-300"
					>
						<span class="text-sm font-medium text-base-content"
							>Auto-refresh:</span
						>
						<label class="relative inline-flex items-center cursor-pointer">
							<input
								type="checkbox"
								v-model="autoRefresh"
								@change="toggleAutoRefresh"
								class="sr-only peer"
							/>
							<div
								class="w-11 h-6 bg-base-300 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-success/30 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-base-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-success"
							></div>
						</label>
						<span
							class="text-xs"
							:class="autoRefresh ? 'text-success' : 'text-base-content/50'"
						>
							{{ autoRefresh ? 'ON' : 'OFF' }}
						</span>
					</div>
				</div>
			</div>

			<!-- Status Bar -->
			<div
				class="flex flex-wrap items-center gap-4 p-4 bg-base-100 rounded-xl shadow-lg border border-base-300"
			>
				<div class="flex items-center space-x-2">
					<div
						:class="{
							'bg-success': connectionStatus === 'connected',
							'bg-warning': connectionStatus === 'partial',
							'bg-error': connectionStatus === 'disconnected',
							'bg-info': connectionStatus === 'connecting',
						}"
						class="w-3 h-3 rounded-full animate-pulse"
					></div>
					<span class="text-sm font-medium text-base-content">
						{{
							connectionStatus === 'connected'
								? 'Connected'
								: connectionStatus === 'partial'
								? 'Partial Connection'
								: connectionStatus === 'connecting'
								? 'Connecting...'
								: 'Disconnected'
						}}
					</span>
				</div>
				<div class="text-sm text-base-content/70">
					Coverage: <span class="font-semibold">{{ dataCoverage }}%</span>
				</div>
				<div class="text-sm text-base-content/70">
					Last Update: <span class="font-mono">{{ lastUpdateTime }}</span>
				</div>
				<div class="text-sm text-base-content/70">
					Next Refresh: <span class="font-mono">{{ nextRefreshIn }}s</span>
				</div>
			</div>
		</div>

		<!-- Alert Banner -->
		<div v-if="error" class="alert alert-error mb-6">
			<svg class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
				<path
					fill-rule="evenodd"
					d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
					clip-rule="evenodd"
				/>
			</svg>
			<div>
				<div class="font-bold">Warning</div>
				<div class="text-xs">{{ error }}</div>
			</div>
		</div>

		<!-- Main Content Layout -->
		<div class="grid grid-cols-1 xl:grid-cols-4 gap-8">
			<!-- Left Column: Signal Summary & Key Metrics -->
			<div class="xl:col-span-1 space-y-6">
				<!-- AI Signal Card -->
				<div class="card bg-base-100 shadow-xl">
					<div class="card-body">
						<h3 class="card-title">🤖 AI Trading Signal</h3>
						<div class="text-center">
							<div class="text-3xl font-bold mb-2" :class="signalColorClass">
								{{ signal }}
							</div>
							<div v-if="asset" class="text-base-content/70 text-sm mb-3">
								Recommended Asset:
								<span class="font-semibold">{{ asset }}</span>
							</div>

							<!-- Confidence Meter -->
							<div class="mb-4">
								<div class="flex justify-between text-sm mb-1">
									<span>Confidence</span>
									<span class="font-semibold" :class="confidenceColorClass">{{
										confidence
									}}</span>
								</div>
								<progress
									class="progress progress-primary w-full"
									:value="confidencePercentage"
									max="100"
								></progress>
								<div class="text-xs text-base-content/50 mt-1">
									{{ confidencePercentage }}%
								</div>
							</div>

							<!-- Score Display -->
							<div class="bg-base-200 rounded-lg p-3">
								<div class="text-sm text-base-content/70 mb-1">
									Weighted Score
								</div>
								<div
									class="text-2xl font-mono font-bold"
									:class="scoreColorClass"
								>
									{{ totalScore.toFixed(3) }}
								</div>
								<div class="text-xs text-base-content/50">
									Range: -1.0 to +1.0
								</div>
							</div>
						</div>

						<!-- Quick Stats -->
						<div class="grid grid-cols-3 gap-3 pt-4 border-t border-base-300">
							<div class="text-center">
								<div class="text-lg font-bold text-success">
									{{ bullishCount }}
								</div>
								<div class="text-xs text-base-content/50">Bullish</div>
							</div>
							<div class="text-center">
								<div class="text-lg font-bold text-base-content">
									{{ neutralCount }}
								</div>
								<div class="text-xs text-base-content/50">Neutral</div>
							</div>
							<div class="text-center">
								<div class="text-lg font-bold text-error">
									{{ bearishCount }}
								</div>
								<div class="text-xs text-base-content/50">Bearish</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Market Health Overview -->
				<div class="card bg-base-100 shadow-xl">
					<div class="card-body">
						<h4 class="card-title">📊 Market Health</h4>

						<div class="space-y-3">
							<div class="flex justify-between items-center">
								<span class="text-sm text-base-content/70">Data Quality</span>
								<div class="flex items-center space-x-2">
									<progress
										class="progress progress-success w-16"
										:value="dataQuality"
										max="100"
									></progress>
									<span class="text-sm font-semibold">{{ dataQuality }}%</span>
								</div>
							</div>

							<div class="flex justify-between items-center">
								<span class="text-sm text-base-content/70"
									>Signal Strength</span
								>
								<div class="flex items-center space-x-2">
									<progress
										class="progress progress-info w-16"
										:value="signalStrength * 100"
										max="100"
									></progress>
									<span class="text-sm font-semibold"
										>{{ (signalStrength * 100).toFixed(0) }}%</span
									>
								</div>
							</div>

							<div class="flex justify-between items-center">
								<span class="text-sm text-base-content/70">Consensus</span>
								<div class="flex items-center space-x-2">
									<progress
										class="progress progress-secondary w-16"
										:value="consensusRatio * 100"
										max="100"
									></progress>
									<span class="text-sm font-semibold"
										>{{ (consensusRatio * 100).toFixed(0) }}%</span
									>
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
							:class="
								selectedCategory === category ? 'btn-primary' : 'btn-outline'
							"
							class="btn btn-sm"
						>
							{{ category }}
						</button>
					</div>

					<div class="flex items-center space-x-3">
						<select v-model="sortBy" class="select select-bordered select-sm">
							<option value="weight">Sort by Weight</option>
							<option value="score">Sort by Score</option>
							<option value="alphabetical">Sort A-Z</option>
						</select>

						<button
							@click="refreshAllMetrics"
							:disabled="isRefreshing"
							class="btn btn-primary btn-sm"
						>
							<span
								v-if="isRefreshing"
								class="loading loading-spinner loading-xs"
							></span>
							{{ isRefreshing ? 'Refreshing...' : 'Refresh' }}
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
	</div>
</template>

<script>
import { defineComponent } from 'vue'
import {
	Chart as ChartJS,
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	Title,
	Tooltip,
	Legend,
	Filler,
} from 'chart.js'
import EnhancedMetricCard from './EnhancedMetricCard.vue'

// Register Chart.js components
ChartJS.register(
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	Title,
	Tooltip,
	Legend,
	Filler
)

export default defineComponent({
	name: 'MarketAnalysis',
	components: { EnhancedMetricCard },
	data() {
		return {
			metrics: [
				// Core Market Metrics
				{
					key: 'fear-greed',
					title: 'Fear & Greed Index',
					endpoint: '/api/market/fear-greed',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.15,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Sentiment',
				},
				{
					key: 'btc-dominance',
					title: 'BTC Dominance',
					endpoint: '/api/market/btc-dominance',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.12,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'altcoin-season',
					title: 'Altcoin Season Index',
					endpoint: '/api/market/altcoin-season',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.1,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'rsi',
					title: 'RSI',
					endpoint: '/api/market/rsi',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.12,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'market-cap',
					title: 'Market Cap Change',
					endpoint: '/api/market/market-cap',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.1,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Fundamental',
				},
				{
					key: 'ssr',
					title: 'Stablecoin Supply Ratio',
					endpoint: '/api/market/ssr',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.08,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Fundamental',
				},
				{
					key: 'google-trends',
					title: 'Google Trends',
					endpoint: '/api/market/google-trends',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.06,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Sentiment',
				},
				{
					key: 'moving-averages',
					title: 'Moving Averages',
					endpoint: '/api/market/moving-averages',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.1,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'volume-trend',
					title: 'Volume Trend',
					endpoint: '/api/market/volume-trend',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.08,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'exchange-flows',
					title: 'Exchange Flows',
					endpoint: '/api/market/exchange-flows',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.07,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'On-Chain',
				},
				{
					key: 'active-addresses',
					title: 'Active Addresses',
					endpoint: '/api/market/active-addresses',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.06,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'On-Chain',
				},
				{
					key: 'whale-transactions',
					title: 'Whale Transactions',
					endpoint: '/api/market/whale-transactions',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.06,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'On-Chain',
				},
				{
					key: 'bollinger-bands',
					title: 'Bollinger Bands',
					endpoint: '/api/market/bollinger-bands',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.05,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'funding-rate',
					title: 'Funding Rate',
					endpoint: '/api/market/funding-rate',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.05,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'open-interest',
					title: 'Open Interest',
					endpoint: '/api/market/open-interest',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.05,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'eth-btc-ratio',
					title: 'ETH/BTC Ratio',
					endpoint: '/api/market/eth-btc-ratio',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.04,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				// Advanced DeFi & Crypto Metrics
				{
					key: 'defi-tvl',
					title: 'DeFi TVL',
					endpoint: '/api/market/defi-tvl',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.08,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Fundamental',
				},
				{
					key: 'social-sentiment',
					title: 'Social Sentiment',
					endpoint: '/api/market/social-sentiment',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.07,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Sentiment',
				},
				{
					key: 'options-flow',
					title: 'Options Flow',
					endpoint: '/api/market/options-flow',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.06,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'stablecoin-flows',
					title: 'Stablecoin Flows',
					endpoint: '/api/market/stablecoin-flows',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.07,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'On-Chain',
				},
				{
					key: 'network-health',
					title: 'Network Health',
					endpoint: '/api/market/network-health',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.05,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'On-Chain',
				},
				{
					key: 'institutional-flows',
					title: 'Institutional Flows',
					endpoint: '/api/market/institutional-flows',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.09,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Fundamental',
				},
				{
					key: 'yield-curves',
					title: 'DeFi Yield Premium',
					endpoint: '/api/market/yield-curves',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.04,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Fundamental',
				},
				{
					key: 'correlation-matrix',
					title: 'BTC-Stock Correlation',
					endpoint: '/api/market/correlation-matrix',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.05,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'volatility-surface',
					title: 'Implied Volatility',
					endpoint: '/api/market/volatility-surface',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.04,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
				{
					key: 'liquidation-heatmap',
					title: 'Liquidation Risk',
					endpoint: '/api/market/liquidation-heatmap',
					value: null,
					indicator: 'Hold',
					score: 0,
					weight: 0.06,
					chartData: [],
					chartLabels: [],
					error: false,
					loading: true,
					category: 'Technical',
				},
			],
			totalScore: 0,
			signal: 'Hold',
			asset: null,
			error: null,
			confidence: 'Low',
			lastUpdateTime: null,
			useMockData: false,
			autoRefresh: false,
			connectionStatus: 'disconnected',
			dataCoverage: 0,
			nextRefreshIn: 30,
			selectedCategory: 'All',
			sortBy: 'weight',
			isRefreshing: false,
			signalStrength: 0,
			consensusRatio: 0,
			dataQuality: 0,
			bullishCount: 0,
			neutralCount: 0,
			bearishCount: 0,
			signalColorClass: 'text-warning',
			confidenceColorClass: 'text-info',
			scoreColorClass: 'text-info',
			confidencePercentage: 0,
			categories: ['All', 'Technical', 'Fundamental', 'Sentiment', 'On-Chain'],
			refreshTimer: null,
		}
	},
	mounted() {
		this.lastUpdateTime = new Date().toLocaleTimeString()
		this.connectionStatus = 'connected'
		this.fetchAllMetrics()
		this.startRefreshTimer()
	},

	beforeUnmount() {
		if (this.refreshTimer) {
			clearInterval(this.refreshTimer)
		}
	},

	computed: {
		filteredAndSortedMetrics() {
			let filtered = this.metrics

			// Filter by category
			if (this.selectedCategory !== 'All') {
				filtered = filtered.filter(
					metric => metric.category === this.selectedCategory
				)
			}

			// Sort metrics
			return filtered.sort((a, b) => {
				switch (this.sortBy) {
					case 'weight':
						return b.weight - a.weight
					case 'score':
						return Math.abs(b.score) - Math.abs(a.score)
					case 'alphabetical':
						return a.title.localeCompare(b.title)
					default:
						return 0
				}
			})
		},
	},
	methods: {
		async fetchAllMetrics() {
			this.isRefreshing = true
			this.connectionStatus = 'connecting'

			const promises = this.metrics.map(metric => this.fetchMetric(metric.key))

			try {
				await Promise.allSettled(promises)
				this.calculateOverallSignal()
				this.updateConnectionStatus()
				this.lastUpdateTime = new Date().toLocaleTimeString()
			} catch (error) {
				console.error('Error fetching metrics:', error)
				this.error = 'Failed to fetch market data'
			} finally {
				this.isRefreshing = false
			}
		},

		async fetchMetric(key) {
			const metric = this.metrics.find(m => m.key === key)
			if (!metric) return

			metric.loading = true
			metric.error = false

			try {
				console.log(`Fetching ${key} from ${metric.endpoint}`)
				const response = await fetch(`http://localhost:8080${metric.endpoint}`)

				if (!response.ok) {
					const errorText = await response.text()
					throw new Error(
						`HTTP ${response.status}: ${response.statusText} - ${errorText}`
					)
				}

				const data = await response.json()
				console.log(`Received data for ${key}:`, data)

				// Update metric with real data
				metric.value = data.value || data.current || data.index || 'N/A'
				metric.indicator = this.determineIndicator(
					data.value,
					data.signal,
					data.trend,
					data.indicator
				)
				metric.score = this.calculateScore(
					data.value,
					data.signal,
					data.trend,
					key,
					data.score
				)
				metric.lastUpdated = new Date().toISOString()

				// Update chart data if available - prioritize chart_data from API
				if (data.chart_data && Array.isArray(data.chart_data)) {
					metric.chartData = data.chart_data
					metric.chartLabels =
						data.chart_labels ||
						data.chart_data.map((_, index) => `${index + 1}`)
				} else if (data.historical && Array.isArray(data.historical)) {
					metric.chartData = data.historical.map(
						item => item.value || item.price || item
					)
					metric.chartLabels = data.historical.map((_, index) => `${index + 1}`)
				}
			} catch (error) {
				console.warn(`Failed to fetch ${key}:`, error.message)
				metric.error = true
				metric.value = 'Error'
				metric.indicator = 'Hold'
				metric.score = 0
			} finally {
				metric.loading = false
			}
		},

		determineIndicator(value, signal, trend, apiIndicator) {
			// Prioritize indicator from API response
			if (apiIndicator) {
				return apiIndicator
			}

			// Determine indicator based on API response
			if (signal) {
				return signal.charAt(0).toUpperCase() + signal.slice(1)
			}

			if (trend) {
				switch (trend.toLowerCase()) {
					case 'bullish':
					case 'up':
					case 'positive':
						return 'Buy'
					case 'bearish':
					case 'down':
					case 'negative':
						return 'Sell'
					default:
						return 'Hold'
				}
			}

			// Default logic based on value
			if (typeof value === 'number') {
				if (value > 70) return 'Sell'
				if (value < 30) return 'Buy'
			}

			return 'Hold'
		},

		calculateScore(value, signal, trend, key, apiScore) {
			// Prioritize score from API response
			if (typeof apiScore === 'number') {
				return apiScore
			}
			
			// Calculate normalized score (-1 to 1) based on metric type and value
			if (typeof value !== 'number') return 0

			switch (key) {
				case 'fear-greed':
					// Fear & Greed: 0-100, where 0 is extreme fear (bullish), 100 is extreme greed (bearish)
					return (50 - value) / 50
				case 'rsi':
					// RSI: 0-100, where <30 is oversold (bullish), >70 is overbought (bearish)
					if (value < 30) return (30 - value) / 30
					if (value > 70) return (70 - value) / 30
					return 0
				case 'btc-dominance':
					// BTC Dominance: normalize around 50%
					return (value - 50) / 50
				default:
					// Generic normalization
					if (signal === 'bullish' || trend === 'up') return 0.5
					if (signal === 'bearish' || trend === 'down') return -0.5
					return 0
			}
		},

		calculateOverallSignal() {
			const validMetrics = this.metrics.filter(m => !m.error && !m.loading)

			if (validMetrics.length === 0) {
				this.totalScore = 0
				this.signal = 'Hold'
				this.confidence = 'Low'
				return
			}

			// Calculate weighted score
			let totalWeightedScore = 0
			let totalWeight = 0

			validMetrics.forEach(metric => {
				totalWeightedScore += metric.score * metric.weight
				totalWeight += metric.weight
			})

			this.totalScore = totalWeight > 0 ? totalWeightedScore / totalWeight : 0

			// Determine signal
			if (this.totalScore > 0.3) {
				this.signal = 'Strong Buy'
				this.signalColorClass = 'text-success'
			} else if (this.totalScore > 0.1) {
				this.signal = 'Buy'
				this.signalColorClass = 'text-success'
			} else if (this.totalScore < -0.3) {
				this.signal = 'Strong Sell'
				this.signalColorClass = 'text-error'
			} else if (this.totalScore < -0.1) {
				this.signal = 'Sell'
				this.signalColorClass = 'text-error'
			} else {
				this.signal = 'Hold'
				this.signalColorClass = 'text-warning'
			}

			// Calculate confidence
			const avgAbsScore =
				validMetrics.reduce((sum, m) => sum + Math.abs(m.score), 0) /
				validMetrics.length
			if (avgAbsScore > 0.5) {
				this.confidence = 'High'
				this.confidenceColorClass = 'text-success'
				this.confidencePercentage = Math.min(90, 60 + avgAbsScore * 60)
			} else if (avgAbsScore > 0.2) {
				this.confidence = 'Medium'
				this.confidenceColorClass = 'text-warning'
				this.confidencePercentage = Math.min(70, 40 + avgAbsScore * 60)
			} else {
				this.confidence = 'Low'
				this.confidenceColorClass = 'text-error'
				this.confidencePercentage = Math.max(20, avgAbsScore * 60)
			}

			// Update counts
			this.bullishCount = validMetrics.filter(
				m => m.indicator === 'Buy' || m.indicator === 'Strong Buy'
			).length
			this.bearishCount = validMetrics.filter(
				m => m.indicator === 'Sell' || m.indicator === 'Strong Sell'
			).length
			this.neutralCount = validMetrics.filter(
				m => m.indicator === 'Hold'
			).length

			// Update other metrics
			this.signalStrength = Math.abs(this.totalScore)
			this.consensusRatio =
				Math.max(this.bullishCount, this.bearishCount) / validMetrics.length
		},

		updateConnectionStatus() {
			const validMetrics = this.metrics.filter(m => !m.error && !m.loading)
			const totalMetrics = this.metrics.length

			this.dataCoverage = Math.round((validMetrics.length / totalMetrics) * 100)
			this.dataQuality = this.dataCoverage

			if (this.dataCoverage > 80) {
				this.connectionStatus = 'connected'
			} else if (this.dataCoverage > 50) {
				this.connectionStatus = 'partial'
			} else {
				this.connectionStatus = 'disconnected'
			}
		},

		toggleDataSource() {
			this.useMockData = !this.useMockData
			console.log(`Switched to ${this.useMockData ? 'Mock' : 'Real'} data`)
			if (!this.useMockData) {
				this.fetchAllMetrics()
			}
		},

		toggleAutoRefresh() {
			this.autoRefresh = !this.autoRefresh
			if (this.autoRefresh) {
				this.startRefreshTimer()
			} else {
				this.stopRefreshTimer()
			}
		},

		startRefreshTimer() {
			if (this.refreshTimer) {
				clearInterval(this.refreshTimer)
			}

			if (this.autoRefresh) {
				this.refreshTimer = setInterval(() => {
					this.fetchAllMetrics()
				}, 30000) // Refresh every 30 seconds
			}
		},

		stopRefreshTimer() {
			if (this.refreshTimer) {
				clearInterval(this.refreshTimer)
				this.refreshTimer = null
			}
		},

		refreshAllMetrics() {
			this.fetchAllMetrics()
		},
	},
})
</script>

<style scoped>
.metric-indicator {
	font-weight: 600;
}

.metric-indicator.bullish {
	color: #22c55e;
}

.metric-indicator.bearish {
	color: #ef4444;
}

.metric-indicator.neutral {
	color: #3b82f6;
}
</style>
