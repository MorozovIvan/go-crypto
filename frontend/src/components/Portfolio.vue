<template>
	<div class="min-h-screen bg-gray-50 p-4">
		<!-- Header -->
		<div class="mb-6">
			<h1 class="text-2xl font-semibold text-gray-900 mb-1">Portfolio</h1>
			<p class="text-gray-600 text-sm">
				Track your cryptocurrency holdings and performance
			</p>
		</div>

		<!-- Exchange Tabs -->
		<div class="mb-6">
			<div class="border-b border-gray-200">
				<nav class="-mb-px flex space-x-8">
					<button
						@click="activeTab = 'binance'"
						:class="[
							'py-2 px-1 border-b-2 font-medium text-sm',
							activeTab === 'binance'
								? 'border-orange-500 text-orange-600'
								: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
						]"
					>
						Binance
					</button>
					<button
						@click="activeTab = 'okx'"
						:class="[
							'py-2 px-1 border-b-2 font-medium text-sm',
							activeTab === 'okx'
								? 'border-blue-500 text-blue-600'
								: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
						]"
					>
						OKX
					</button>
				</nav>
			</div>
		</div>

		<!-- Tab Content -->
		<div v-if="activeTab === 'binance'">
			<!-- Binance Portfolio -->
			<div v-if="loading" class="flex justify-center items-center py-12">
				<div
					class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"
				></div>
			</div>

			<div
				v-else-if="
					binancePortfolio &&
					binancePortfolio.assets &&
					binancePortfolio.assets.length > 0
				"
			>
				<!-- Portfolio Overview Chart -->
				<div class="bg-white rounded-lg border border-gray-200 p-6 mb-6">
					<div class="flex justify-between items-center mb-4">
						<h2 class="text-lg font-medium text-gray-900">
							Portfolio Performance
						</h2>
						<div class="flex items-center space-x-3">
							<!-- Portfolio Price Trends Button -->
							<PortfolioPriceTrends />

							<div class="flex space-x-1">
								<button
									v-for="period in ['7d', '30d', '90d', '1y']"
									:key="period"
									@click="changePeriod(period)"
									:class="[
										'px-3 py-1 rounded text-sm font-medium',
										selectedPeriod === period
											? 'bg-gray-900 text-white'
											: 'text-gray-500 hover:text-gray-700',
									]"
								>
									{{ period }}
								</button>
							</div>
						</div>
					</div>
					<div class="h-64">
						<canvas id="portfolioChart"></canvas>
					</div>
				</div>

				<!-- Portfolio Summary -->
				<div class="grid grid-cols-1 md:grid-cols-5 gap-4 mb-6">
					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-500 text-sm">Total Balance</p>
								<p class="text-xl font-semibold text-gray-900">
									${{ formatNumber(binancePortfolio.totalValue || 0) }}
								</p>
							</div>
						</div>
					</div>

					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-500 text-sm">24h Change</p>
								<p
									:class="[
										'text-xl font-semibold',
										(binancePortfolio.portfolioChange || 0) >= 0
											? 'text-green-600'
											: 'text-red-600',
									]"
								>
									{{ formatPercentage(binancePortfolio.portfolioChange || 0) }}%
								</p>
							</div>
						</div>
					</div>

					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-500 text-sm">Assets Count</p>
								<p class="text-xl font-semibold text-gray-900">
									{{ binancePortfolio.assets?.length || 0 }}
								</p>
							</div>
						</div>
					</div>

					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-500 text-sm">24h Volume</p>
								<p class="text-xl font-semibold text-gray-900">
									${{ formatNumber(binancePortfolio.volume24h || 0) }}
								</p>
							</div>
						</div>
					</div>

					<div class="bg-white rounded-lg border border-gray-200 p-4">
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-500 text-sm">Total P&L</p>
								<p
									:class="[
										'text-xl font-semibold',
										(binancePortfolio.totalProfitLoss || 0) >= 0
											? 'text-green-600'
											: 'text-red-600',
									]"
								>
									{{
										(binancePortfolio.totalProfitLoss || 0) >= 0 ? '+' : ''
									}}${{
										formatNumber(
											Math.abs(binancePortfolio.totalProfitLoss || 0)
										)
									}}
								</p>
							</div>
						</div>
					</div>

					<div
						class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700/50"
					>
						<div class="flex items-center justify-between">
							<div>
								<p class="text-gray-400 text-sm">Initial Investment</p>
								<p class="text-2xl font-bold text-white">
									${{ formatNumber(binancePortfolio.initialInvestment || 0) }}
								</p>
							</div>
							<div class="p-3 bg-indigo-500/20 rounded-lg">
								<svg
									class="w-6 h-6 text-indigo-400"
									fill="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"
									/>
								</svg>
							</div>
						</div>
					</div>
				</div>

				<!-- Assets Grid -->
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					<div
						v-for="asset in binancePortfolio.assets"
						:key="asset.symbol"
						@click="fetchAssetDetails(asset.symbol)"
						class="bg-white rounded-lg border border-gray-200 p-4 hover:border-gray-300 transition-colors cursor-pointer"
					>
						<!-- Asset Header -->
						<div class="flex items-center justify-between mb-3">
							<div class="flex items-center space-x-3">
								<div
									class="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center"
								>
									<span class="text-gray-700 font-medium text-xs">{{
										asset.symbol?.slice(0, 2) || 'N/A'
									}}</span>
								</div>
								<div>
									<h3 class="text-gray-900 font-medium">
										{{ asset.symbol || 'Unknown' }}
									</h3>
								</div>
							</div>
							<div class="text-right">
								<p class="text-gray-900 font-medium">
									${{ formatNumber(asset.value || 0) }}
								</p>
								<p
									:class="[
										'text-sm',
										(asset.change || 0) >= 0
											? 'text-green-600'
											: 'text-red-600',
									]"
								>
									{{ formatPercentage(asset.change || 0) }}%
								</p>
							</div>
						</div>

						<!-- Asset Details -->
						<div class="space-y-2">
							<div class="flex justify-between text-sm">
								<span class="text-gray-500">Amount</span>
								<span class="text-gray-900">{{
									formatNumber(asset.amount || 0)
								}}</span>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-500">Current Price</span>
								<span class="text-gray-900"
									>${{ formatNumber(asset.currentPrice || 0) }}</span
								>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-500">Purchase Price</span>
								<span class="text-gray-900"
									>${{ formatNumber(asset.purchasePrice || 0) }}</span
								>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-500">P&L</span>
								<span
									:class="[
										'font-medium',
										(asset.profitLoss || 0) >= 0
											? 'text-green-600'
											: 'text-red-600',
									]"
								>
									{{ (asset.profitLoss || 0) >= 0 ? '+' : '' }}${{
										formatNumber(Math.abs(asset.profitLoss || 0))
									}}
								</span>
							</div>
							<div class="flex justify-between text-sm">
								<span class="text-gray-500">P&L %</span>
								<span
									:class="[
										'font-medium',
										(asset.profitLossPerc || 0) >= 0
											? 'text-green-600'
											: 'text-red-600',
									]"
								>
									{{ formatPercentage(asset.profitLossPerc || 0) }}%
								</span>
							</div>
						</div>

						<!-- Click to view chart hint -->
						<div
							class="mt-4 h-16 bg-gray-50 rounded border border-dashed border-gray-300 flex items-center justify-center"
						>
							<div class="text-center">
								<svg
									class="w-5 h-5 text-gray-400 mx-auto mb-1"
									fill="none"
									stroke="currentColor"
									viewBox="0 0 24 24"
								>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
									/>
								</svg>
								<span class="text-gray-400 text-xs">Click to view chart</span>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- No Data State -->
			<div
				v-else-if="
					!binancePortfolio ||
					!binancePortfolio.assets ||
					binancePortfolio.assets.length === 0
				"
				class="text-center py-12"
			>
				<div
					class="bg-white rounded-lg border border-gray-200 p-8 max-w-md mx-auto"
				>
					<svg
						class="w-12 h-12 text-gray-400 mx-auto mb-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z"
						/>
					</svg>
					<h3 class="text-gray-900 font-medium mb-2">
						No Portfolio Data Available
					</h3>
					<p class="text-gray-500 text-sm mb-4">
						Configure your Binance API credentials to view portfolio data.
					</p>
					<button
						@click="fetchBinancePortfolio"
						class="bg-gray-900 hover:bg-gray-800 text-white px-4 py-2 rounded text-sm"
					>
						Retry
					</button>
				</div>
			</div>

			<!-- Error State -->
			<div v-else-if="error" class="text-center py-12">
				<div
					class="bg-white rounded-lg border border-red-200 p-8 max-w-md mx-auto"
				>
					<svg
						class="w-12 h-12 text-red-500 mx-auto mb-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<h3 class="text-gray-900 font-medium mb-2">
						Failed to Load Portfolio
					</h3>
					<p class="text-gray-500 text-sm mb-4">{{ error }}</p>
					<button
						@click="fetchBinancePortfolio"
						class="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded text-sm"
					>
						Retry
					</button>
				</div>
			</div>
		</div>

		<!-- OKX Tab Content -->
		<div v-else-if="activeTab === 'okx'">
			<div class="text-center py-12">
				<div
					class="bg-white rounded-lg border border-gray-200 p-8 max-w-md mx-auto"
				>
					<svg
						class="w-12 h-12 text-gray-400 mx-auto mb-4"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 100 4m0-4v2m0-6V4"
						/>
					</svg>
					<h3 class="text-gray-900 font-medium text-lg mb-2">
						OKX Integration
					</h3>
					<p class="text-gray-500 mb-6">
						OKX portfolio integration is coming soon.
					</p>
					<button
						class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded text-sm"
					>
						Connect OKX Account
					</button>
				</div>
			</div>
		</div>

		<!-- Asset Details Modal -->
		<div
			v-if="showAssetModal"
			class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
			@click="closeAssetModal"
		>
			<div
				class="bg-gray-800 rounded-xl p-6 max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto"
				@click.stop
			>
				<div class="flex justify-between items-center mb-6">
					<h2 class="text-2xl font-bold text-white">
						{{ selectedAsset }} Details
					</h2>
					<button
						@click="closeAssetModal"
						class="text-gray-400 hover:text-white"
					>
						<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
							<path
								d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
							/>
						</svg>
					</button>
				</div>

				<div v-if="loadingAsset" class="flex justify-center items-center py-20">
					<div
						class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"
					></div>
				</div>

				<div v-else-if="assetDetails" class="space-y-6">
					<!-- Asset Summary -->
					<div class="grid grid-cols-1 md:grid-cols-4 gap-6">
						<div class="bg-gray-700/50 rounded-lg p-4">
							<p class="text-gray-400 text-sm">Current Value</p>
							<p class="text-xl font-bold text-white">
								${{ formatNumber(assetDetails.value || 0) }}
							</p>
						</div>
						<div class="bg-gray-700/50 rounded-lg p-4">
							<p class="text-gray-400 text-sm">Total P&L</p>
							<p
								:class="[
									'text-xl font-bold',
									(assetDetails.profitLoss || 0) >= 0
										? 'text-green-400'
										: 'text-red-400',
								]"
							>
								{{ (assetDetails.profitLoss || 0) >= 0 ? '+' : '' }}${{
									formatNumber(Math.abs(assetDetails.profitLoss || 0))
								}}
							</p>
						</div>
						<div class="bg-gray-700/50 rounded-lg p-4">
							<p class="text-gray-400 text-sm">P&L Percentage</p>
							<p
								:class="[
									'text-xl font-bold',
									(assetDetails.profitLossPerc || 0) >= 0
										? 'text-green-400'
										: 'text-red-400',
								]"
							>
								{{ formatPercentage(assetDetails.profitLossPerc || 0) }}%
							</p>
						</div>
						<div class="bg-gray-700/50 rounded-lg p-4">
							<p class="text-gray-400 text-sm">Purchase Date</p>
							<p class="text-xl font-bold text-white">
								{{ new Date(assetDetails.purchaseDate).toLocaleDateString() }}
							</p>
						</div>
					</div>

					<!-- Price Chart -->
					<div class="bg-gray-700/50 rounded-lg p-6">
						<h3 class="text-lg font-semibold text-white mb-4">
							Price Since Purchase
						</h3>
						<div class="h-80">
							<canvas id="assetChart"></canvas>
						</div>
					</div>

					<!-- Detailed Information -->
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div class="bg-gray-700/50 rounded-lg p-4">
							<h4 class="text-white font-semibold mb-3">
								Purchase Information
							</h4>
							<div class="space-y-2">
								<div class="flex justify-between">
									<span class="text-gray-400">Purchase Price</span>
									<span class="text-white"
										>${{ formatNumber(assetDetails.purchasePrice || 0) }}</span
									>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-400">Amount Purchased</span>
									<span class="text-white">{{
										formatNumber(assetDetails.amount || 0)
									}}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-400">Initial Investment</span>
									<span class="text-white"
										>${{
											formatNumber(
												(assetDetails.amount || 0) *
													(assetDetails.purchasePrice || 0)
											)
										}}</span
									>
								</div>
							</div>
						</div>
						<div class="bg-gray-700/50 rounded-lg p-4">
							<h4 class="text-white font-semibold mb-3">Current Information</h4>
							<div class="space-y-2">
								<div class="flex justify-between">
									<span class="text-gray-400">Current Price</span>
									<span class="text-white"
										>${{ formatNumber(assetDetails.currentPrice || 0) }}</span
									>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-400">24h High</span>
									<span class="text-white"
										>${{ formatNumber(assetDetails.high24h || 0) }}</span
									>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-400">24h Low</span>
									<span class="text-white"
										>${{ formatNumber(assetDetails.low24h || 0) }}</span
									>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { Chart, registerables } from 'chart.js'
import PortfolioPriceTrends from './PortfolioPriceTrends.vue'

Chart.register(...registerables)

export default {
	name: 'Portfolio',
	components: {
		PortfolioPriceTrends,
	},
	data() {
		return {
			activeTab: 'binance',
			loading: false,
			error: null,
			binancePortfolio: null,
			refreshInterval: null,
			selectedPeriod: '30d',
			portfolioHistory: [],
			selectedAsset: null,
			showAssetModal: false,
			assetDetails: null,
			loadingAsset: false,
		}
	},
	methods: {
		async fetchBinancePortfolio() {
			this.loading = true
			this.error = null

			try {
				console.log('Fetching portfolio from /api/portfolio...')
				const response = await fetch('/api/portfolio')
				console.log('Portfolio API response status:', response.status)

				if (!response.ok) {
					const errorText = await response.text()
					console.error('Portfolio API error response:', errorText)
					throw new Error(
						`HTTP error! status: ${response.status} - ${errorText}`
					)
				}

				const data = await response.json()
				console.log('Portfolio API success response:', data)

				// Only set data if it's valid and not mock data
				if (data && typeof data === 'object' && !data.error) {
					this.binancePortfolio = data
					console.log('Portfolio data set successfully')

					// Try to render chart now that portfolio data is loaded
					this.tryRenderChart()
				} else {
					throw new Error('Invalid portfolio data received')
				}
			} catch (error) {
				console.error('Failed to fetch Binance portfolio:', error)
				this.error = `Portfolio data unavailable: ${error.message}`
				this.binancePortfolio = null
			} finally {
				this.loading = false
			}
		},

		formatNumber(value) {
			if (value === null || value === undefined) return '0.00'
			return new Intl.NumberFormat('en-US', {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2,
			}).format(value)
		},

		formatPercentage(value) {
			if (value === null || value === undefined) return '0.00'
			return new Intl.NumberFormat('en-US', {
				minimumFractionDigits: 2,
				maximumFractionDigits: 2,
				signDisplay: 'always',
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
		},

		async fetchPortfolioHistory(period = '30d') {
			try {
				const response = await fetch(`/api/portfolio/history/${period}`)
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`)
				}

				const data = await response.json()
				console.log('Portfolio history response:', data)
				// Handle both formats: direct array or object with history property
				this.portfolioHistory = data.history || data
				// Try to render chart now that history data is loaded
				this.tryRenderChart()
			} catch (error) {
				console.error('Failed to fetch portfolio history:', error)
			}
		},

		async fetchAssetDetails(symbol) {
			this.loadingAsset = true
			try {
				const response = await fetch(`/api/portfolio/asset/${symbol}`)
				if (!response.ok) {
					throw new Error(`HTTP error! status: ${response.status}`)
				}

				const data = await response.json()
				this.assetDetails = data
				this.selectedAsset = symbol
				this.showAssetModal = true
				this.$nextTick(() => {
					this.renderAssetChart()
				})
			} catch (error) {
				console.error('Failed to fetch asset details:', error)
			} finally {
				this.loadingAsset = false
			}
		},

		changePeriod(period) {
			this.selectedPeriod = period
			this.fetchPortfolioHistory(period)
		},

		closeAssetModal() {
			this.showAssetModal = false
			this.selectedAsset = null
			this.assetDetails = null
			if (this.assetChart) {
				this.assetChart.destroy()
				this.assetChart = null
			}
		},

		tryRenderChart() {
			// Helper method to safely attempt chart rendering
			if (
				this.portfolioHistory &&
				this.portfolioHistory.length > 0 &&
				this.binancePortfolio &&
				this.binancePortfolio.assets &&
				this.binancePortfolio.assets.length > 0
			) {
				this.$nextTick(() => {
					setTimeout(() => {
						this.renderPortfolioChart()
					}, 100)
				})
			}
		},

		renderPortfolioChart() {
			console.log('Rendering portfolio chart with data:', this.portfolioHistory)
			if (!this.portfolioHistory || this.portfolioHistory.length === 0) {
				console.log('No portfolio history data available')
				return
			}

			// Check if portfolio data is loaded (required for canvas to exist)
			if (
				!this.binancePortfolio ||
				!this.binancePortfolio.assets ||
				this.binancePortfolio.assets.length === 0
			) {
				console.log('Portfolio data not loaded yet, deferring chart rendering')
				return
			}

			const ctx = document.getElementById('portfolioChart')
			if (!ctx) {
				console.log('Chart canvas element not found - DOM may not be ready')
				// Retry after a longer delay
				setTimeout(() => {
					this.renderPortfolioChart()
				}, 500)
				return
			}

			// Destroy existing chart if it exists
			if (this.portfolioChart) {
				this.portfolioChart.destroy()
			}

			console.log(
				'Creating new chart with data points:',
				this.portfolioHistory.length
			)
			try {
				this.portfolioChart = new Chart(ctx, {
					type: 'line',
					data: {
						labels: this.portfolioHistory.map(point =>
							new Date(point.timestamp).toLocaleDateString()
						),
						datasets: [
							{
								label: 'Portfolio Value',
								data: this.portfolioHistory.map(point => point.value),
								borderColor: '#374151',
								backgroundColor: 'rgba(55, 65, 81, 0.1)',
								borderWidth: 2,
								fill: true,
								tension: 0.4,
							},
						],
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						plugins: {
							legend: {
								display: false,
							},
						},
						scales: {
							x: {
								display: true,
								grid: {
									color: 'rgba(255, 255, 255, 0.1)',
								},
								ticks: {
									color: '#9CA3AF',
								},
							},
							y: {
								display: true,
								grid: {
									color: 'rgba(255, 255, 255, 0.1)',
								},
								ticks: {
									color: '#6B7280',
									callback: function (value) {
										return '$' + value.toLocaleString()
									},
								},
							},
						},
					},
				})
				console.log('Chart created successfully')
			} catch (error) {
				console.error('Error creating portfolio chart:', error)
			}
		},

		renderAssetChart() {
			if (!this.assetDetails || !this.assetDetails.priceHistory) return

			const ctx = document.getElementById('assetChart')
			if (!ctx) return

			// Destroy existing chart if it exists
			if (this.assetChart) {
				this.assetChart.destroy()
			}

			const priceHistory = this.assetDetails.priceHistory
			const isProfit = this.assetDetails.profitLoss >= 0

			this.assetChart = new Chart(ctx, {
				type: 'line',
				data: {
					labels: priceHistory.map(point =>
						new Date(point.timestamp).toLocaleDateString()
					),
					datasets: [
						{
							label: `${this.assetDetails.symbol} Price`,
							data: priceHistory.map(point => point.price),
							borderColor: isProfit ? '#059669' : '#DC2626',
							backgroundColor: isProfit
								? 'rgba(5, 150, 105, 0.1)'
								: 'rgba(220, 38, 38, 0.1)',
							borderWidth: 2,
							fill: true,
							tension: 0.4,
						},
					],
				},
				options: {
					responsive: true,
					maintainAspectRatio: false,
					plugins: {
						legend: {
							display: false,
						},
					},
					scales: {
						x: {
							display: true,
							grid: {
								color: 'rgba(255, 255, 255, 0.1)',
							},
							ticks: {
								color: '#9CA3AF',
							},
						},
						y: {
							display: true,
							grid: {
								color: 'rgba(255, 255, 255, 0.1)',
							},
							ticks: {
								color: '#9CA3AF',
								callback: function (value) {
									return '$' + value.toFixed(2)
								},
							},
						},
					},
				},
			})
		},
	},

	watch: {
		activeTab(newTab) {
			if (newTab === 'binance' && !this.binancePortfolio) {
				this.fetchBinancePortfolio()
			}
		},
	},

	mounted() {
		// Load Binance portfolio on component mount
		this.fetchBinancePortfolio()
		this.fetchPortfolioHistory(this.selectedPeriod)
		this.startAutoRefresh()
	},

	beforeUnmount() {
		this.stopAutoRefresh()
		if (this.portfolioChart) {
			this.portfolioChart.destroy()
		}
		if (this.assetChart) {
			this.assetChart.destroy()
		}
	},
}
</script>

<style scoped>
/* Additional custom styles if needed */
.backdrop-blur-sm {
	backdrop-filter: blur(4px);
}
</style>
