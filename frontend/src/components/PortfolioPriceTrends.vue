<template>
	<div>
		<!-- Price Trends Button -->
		<button
			@click="handleButtonClick"
			class="price-trends-btn"
			:disabled="loading"
		>
			<svg
				class="btn-icon"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<path d="M3 3v18h18"></path>
				<path d="M7 12l4-4 4 4 6-6"></path>
			</svg>
			View Portfolio Price Trends
		</button>

		<!-- Price Trends Modal -->
		<div
			v-if="showModal"
			class="modal show"
			role="dialog"
			aria-labelledby="modalTitle"
			aria-hidden="false"
			@click="closeModal"
			style="display: flex !important"
		>
			<div class="modal-content" @click.stop>
				<!-- Modal Header -->
				<div class="modal-header">
					<h2 id="modalTitle">Portfolio Price Trends</h2>
					<button
						@click="closeModal"
						class="close-btn"
						aria-label="Close modal"
					>
						<svg
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<line x1="18" y1="6" x2="6" y2="18"></line>
							<line x1="6" y1="6" x2="18" y2="18"></line>
						</svg>
					</button>
				</div>

				<!-- Time Period Selector -->
				<div class="time-selector">
					<label for="timePeriod">Time Period:</label>
					<select
						v-model="currentPeriod"
						@change="handleTimePeriodChange"
						class="time-dropdown"
					>
						<option value="1d">1 Day</option>
						<option value="7d">7 Days</option>
						<option value="30d">30 Days</option>
						<option value="90d">90 Days</option>
						<option value="custom">Custom Range</option>
					</select>

					<!-- Custom Date Range -->
					<div v-if="currentPeriod === 'custom'" class="custom-range">
						<input v-model="startDate" type="date" class="date-input" />
						<span>to</span>
						<input v-model="endDate" type="date" class="date-input" />
						<button @click="applyCustomRange" class="apply-btn">Apply</button>
					</div>
				</div>

				<!-- Chart Container -->
				<div class="chart-container">
					<!-- Loading Indicator -->
					<div v-if="chartLoading" class="loading-indicator">
						<div class="spinner"></div>
						<p>Loading price data...</p>
					</div>

					<!-- Error Message -->
					<div v-else-if="error" class="error-message">
						<svg
							class="error-icon"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
						>
							<circle cx="12" cy="12" r="10"></circle>
							<line x1="15" y1="9" x2="9" y2="15"></line>
							<line x1="9" y1="9" x2="15" y2="15"></line>
						</svg>
						<p>{{ error }}</p>
						<button @click="loadChartData" class="retry-btn">Retry</button>
					</div>

					<!-- Chart Canvas -->
					<canvas v-else ref="priceChart" class="chart-canvas"></canvas>
				</div>

				<!-- Chart Stats -->
				<div v-if="!chartLoading && !error && stats" class="chart-stats">
					<div class="stats-grid">
						<div class="stat-item">
							<span class="stat-label">Best Performer:</span>
							<span
								:class="[
									'stat-value',
									stats.bestPerformer.change >= 0 ? 'positive' : 'negative',
								]"
							>
								{{ stats.bestPerformer.token }}:
								{{ stats.bestPerformer.change >= 0 ? '+' : ''
								}}{{ stats.bestPerformer.change.toFixed(2) }}%
							</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Worst Performer:</span>
							<span
								:class="[
									'stat-value',
									stats.worstPerformer.change >= 0 ? 'positive' : 'negative',
								]"
							>
								{{ stats.worstPerformer.token }}:
								{{ stats.worstPerformer.change >= 0 ? '+' : ''
								}}{{ stats.worstPerformer.change.toFixed(2) }}%
							</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Average Change:</span>
							<span
								:class="[
									'stat-value',
									stats.averageChange >= 0 ? 'positive' : 'negative',
								]"
							>
								{{ stats.averageChange >= 0 ? '+' : ''
								}}{{ stats.averageChange.toFixed(2) }}%
							</span>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import Chart from 'chart.js/auto'
// Note: We're not using time scale to avoid the chartjs-adapter-date-fns dependency
import 'chartjs-adapter-date-fns'

export default {
	name: 'PortfolioPriceTrends',
	data() {
		return {
			showModal: false,
			loading: false,
			chartLoading: false,
			error: null,
			chart: null,
			currentPeriod: '7d',
			startDate: '',
			endDate: '',
			portfolioTokens: [],
			cache: new Map(),
			cacheExpiry: 5 * 60 * 1000, // 5 minutes
			stats: null,
			clickCount: 0,
			colors: [
				'#FF6384',
				'#36A2EB',
				'#FFCE56',
				'#4BC0C0',
				'#9966FF',
				'#FF9F40',
				'#FF6384',
				'#C9CBCF',
				'#4BC0C0',
				'#FF6384',
				'#36A2EB',
				'#FFCE56',
				'#4BC0C0',
				'#9966FF',
				'#FF9F40',
			],
		}
	},
	async mounted() {
		console.log('PortfolioPriceTrends: component mounted')
		await this.loadPortfolioTokens()
		this.setDefaultCustomRange()
		console.log(
			'PortfolioPriceTrends: initialization complete, portfolioTokens:',
			this.portfolioTokens.length
		)
	},
	beforeUnmount() {
		if (this.chart) {
			this.chart.destroy()
		}
	},
	methods: {
		handleButtonClick() {
			console.log('PortfolioPriceTrends: Button clicked!')
			this.showModal = true
		},

		async loadPortfolioTokens() {
			try {
				const response = await fetch('/api/portfolio')
				const data = await response.json()

				if (data.assets) {
					// Filter out problematic tokens for now and limit to first 5 for demo
					const validTokens = data.assets
						.filter(
							asset => !asset.symbol.startsWith('LD') && asset.symbol !== 'USDT'
						)
						.slice(0, 5)

					this.portfolioTokens = validTokens.map(asset => ({
						id: asset.symbol,
						name: asset.symbol,
						amount: asset.amount,
					}))
				}
			} catch (error) {
				console.error('Failed to load portfolio tokens:', error)
				this.portfolioTokens = []
			}
		},

		async openModal() {
			console.log('PortfolioPriceTrends: openModal called')
			console.log('PortfolioPriceTrends: Before - showModal:', this.showModal)
			this.showModal = true
			console.log('PortfolioPriceTrends: After - showModal:', this.showModal)

			// Force Vue to update the DOM
			await this.$nextTick()
			console.log(
				'PortfolioPriceTrends: After nextTick - showModal:',
				this.showModal
			)

			// Wait longer for the canvas to be available and DOM to settle
			setTimeout(async () => {
				console.log('PortfolioPriceTrends: Loading chart data...')
				await this.loadChartData()
			}, 300)
		},

		closeModal() {
			this.showModal = false
			if (this.chart) {
				this.chart.destroy()
				this.chart = null
			}
		},

		handleTimePeriodChange() {
			if (this.currentPeriod !== 'custom') {
				this.loadChartData()
			}
		},

		setDefaultCustomRange() {
			const endDate = new Date()
			const startDate = new Date()
			startDate.setDate(startDate.getDate() - 30)

			this.endDate = endDate.toISOString().split('T')[0]
			this.startDate = startDate.toISOString().split('T')[0]
		},

		applyCustomRange() {
			if (!this.startDate || !this.endDate) {
				alert('Please select both start and end dates')
				return
			}

			if (new Date(this.startDate) >= new Date(this.endDate)) {
				alert('Start date must be before end date')
				return
			}

			this.loadChartData()
		},

		async loadChartData() {
			this.chartLoading = true
			this.error = null

			try {
				const priceData = await this.fetchAllTokenPrices()

				// Clear loading state before rendering chart so canvas becomes available
				this.chartLoading = false
				await this.$nextTick()

				await this.renderChart(priceData)
				this.updateStats(priceData)
			} catch (error) {
				console.error('Failed to load chart data:', error)
				this.error = 'Failed to load price data. Please try again.'
				this.chartLoading = false
			}
		},

		async fetchAllTokenPrices() {
			const promises = this.portfolioTokens.map(token =>
				this.fetchTokenPrices(token.id)
			)

			const results = await Promise.allSettled(promises)
			const priceData = {}

			results.forEach((result, index) => {
				const tokenId = this.portfolioTokens[index].id
				if (result.status === 'fulfilled') {
					priceData[tokenId] = result.value
				} else {
					console.warn(`Failed to fetch data for ${tokenId}:`, result.reason)
				}
			})

			return priceData
		},

		async fetchTokenPrices(tokenId) {
			const cacheKey = `${tokenId}_${this.currentPeriod}`

			// Check cache first
			if (this.cache.has(cacheKey)) {
				const cached = this.cache.get(cacheKey)
				if (Date.now() - cached.timestamp < this.cacheExpiry) {
					return cached.data
				}
			}

			// Determine API endpoint
			let apiUrl
			if (this.currentPeriod === 'custom') {
				apiUrl = `/api/portfolio/token-prices/${tokenId}?start=${this.startDate}&end=${this.endDate}`
			} else {
				apiUrl = `/api/portfolio/token-prices/${tokenId}?period=${this.currentPeriod}`
			}

			const response = await fetch(apiUrl)
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`)
			}

			const data = await response.json()

			// Cache the result
			this.cache.set(cacheKey, {
				data: data,
				timestamp: Date.now(),
			})

			return data
		},

		async renderChart(priceData) {
			if (this.chart) {
				this.chart.destroy()
			}

			// Wait for the DOM to update
			await this.$nextTick()

			// Retry mechanism for canvas reference
			let retries = 0
			const maxRetries = 5

			while (!this.$refs.priceChart && retries < maxRetries) {
				console.log(`Waiting for canvas reference, attempt ${retries + 1}`)
				await new Promise(resolve => setTimeout(resolve, 100))
				await this.$nextTick()
				retries++
			}

			// Check if the canvas ref exists after retries
			if (!this.$refs.priceChart) {
				console.error('Canvas reference not found after retries')
				this.error = 'Chart canvas not available. Please try again.'
				return
			}

			try {
				const ctx = this.$refs.priceChart.getContext('2d')
				const { datasets, labels } = this.createChartDatasets(priceData)

				this.chart = new Chart(ctx, {
					type: 'line',
					data: {
						labels: labels,
						datasets: datasets,
					},
					options: {
						responsive: true,
						maintainAspectRatio: false,
						interaction: {
							mode: 'index',
							intersect: false,
						},
						plugins: {
							title: {
								display: true,
								text: `Portfolio Price Changes - ${this.getPeriodLabel()}`,
								font: { size: 16, weight: 'bold' },
							},
							legend: {
								display: true,
								position: 'top',
								labels: { usePointStyle: true, padding: 20 },
							},
							tooltip: {
								backgroundColor: 'rgba(0, 0, 0, 0.8)',
								titleColor: 'white',
								bodyColor: 'white',
								borderColor: 'rgba(255, 255, 255, 0.2)',
								borderWidth: 1,
								callbacks: {
									label: context => {
										const value = context.parsed.y
										const sign = value >= 0 ? '+' : ''
										return `${context.dataset.label}: ${sign}${value.toFixed(
											2
										)}%`
									},
								},
							},
						},
						scales: {
							x: {
								type: 'time',
								time: {
									displayFormats: {
										day: 'MMM dd',
										hour: 'HH:mm',
									},
								},
								title: { display: true, text: 'Time' },
							},
							y: {
								title: { display: true, text: 'Price Change (%)' },
								ticks: {
									callback: value => value.toFixed(1) + '%',
								},
							},
						},
						elements: {
							line: { tension: 0.4 },
							point: { radius: 3, hoverRadius: 6 },
						},
					},
				})
			} catch (error) {
				console.error('Error creating chart:', error)
				this.error = 'Failed to create chart. Please try again.'
			}
		},

		createChartDatasets(priceData) {
			const datasets = []
			let colorIndex = 0
			let labels = []

			Object.keys(priceData).forEach(tokenId => {
				const tokenData = priceData[tokenId]
				if (!tokenData.prices || tokenData.prices.length === 0) return

				const prices = tokenData.prices
				const basePrice = prices[0].price

				const data = prices.map(point => ({
					x: new Date(point.timestamp),
					y: ((point.price - basePrice) / basePrice) * 100,
				}))

				const color = this.colors[colorIndex % this.colors.length]

				datasets.push({
					label: tokenId,
					data: data,
					borderColor: color,
					backgroundColor: color + '20',
					fill: false,
					tension: 0.4,
				})

				colorIndex++
			})

			return datasets
		},

		updateStats(priceData) {
			const changes = []

			Object.keys(priceData).forEach(tokenId => {
				const tokenData = priceData[tokenId]
				if (!tokenData.prices || tokenData.prices.length < 2) return

				const prices = tokenData.prices
				const firstPrice = prices[0].price
				const lastPrice = prices[prices.length - 1].price
				const change = ((lastPrice - firstPrice) / firstPrice) * 100

				changes.push({ token: tokenId, change })
			})

			if (changes.length === 0) return

			changes.sort((a, b) => b.change - a.change)

			this.stats = {
				bestPerformer: changes[0],
				worstPerformer: changes[changes.length - 1],
				averageChange:
					changes.reduce((sum, item) => sum + item.change, 0) / changes.length,
			}
		},

		getPeriodLabel() {
			switch (this.currentPeriod) {
				case '1d':
					return 'Last 24 Hours'
				case '7d':
					return 'Last 7 Days'
				case '30d':
					return 'Last 30 Days'
				case '90d':
					return 'Last 90 Days'
				case 'custom':
					return `${this.startDate} to ${this.endDate}`
				default:
					return 'Custom Period'
			}
		},
	},
}
</script>

<style scoped>
/* Price Trends Button */
.price-trends-btn {
	display: inline-flex;
	align-items: center;
	gap: 8px;
	background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	color: white;
	border: none;
	padding: 12px 24px;
	border-radius: 8px;
	font-size: 16px;
	font-weight: 600;
	cursor: pointer;
	transition: all 0.3s ease;
	box-shadow: 0 2px 4px rgba(102, 126, 234, 0.3);
	position: relative;
	z-index: 10;
	pointer-events: auto;
}

.price-trends-btn:hover:not(:disabled) {
	transform: translateY(-2px);
	box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.price-trends-btn:disabled {
	opacity: 0.6;
	cursor: not-allowed;
}

.btn-icon {
	width: 20px;
	height: 20px;
}

/* Modal Styles */
.modal {
	position: fixed;
	z-index: 1000;
	left: 0;
	top: 0;
	width: 100%;
	height: 100%;
	background-color: rgba(0, 0, 0, 0.5);
	backdrop-filter: blur(4px);
	display: flex;
	align-items: center;
	justify-content: center;
	animation: fadeIn 0.3s ease;
}

.modal-content {
	background: white;
	border-radius: 16px;
	width: 90%;
	max-width: 1000px;
	max-height: 90vh;
	overflow-y: auto;
	box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
	animation: slideIn 0.3s ease;
}

.modal-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 24px 24px 0;
	border-bottom: 1px solid #e9ecef;
	margin-bottom: 24px;
}

.modal-header h2 {
	margin: 0;
	font-size: 24px;
	font-weight: 700;
	color: #2c3e50;
}

.close-btn {
	background: none;
	border: none;
	cursor: pointer;
	padding: 8px;
	border-radius: 50%;
	transition: background-color 0.2s ease;
	color: #6c757d;
}

.close-btn:hover {
	background-color: #f8f9fa;
	color: #495057;
}

.close-btn svg {
	width: 24px;
	height: 24px;
}

/* Time Selector */
.time-selector {
	padding: 0 24px 24px;
	border-bottom: 1px solid #e9ecef;
	margin-bottom: 24px;
}

.time-selector label {
	display: block;
	margin-bottom: 8px;
	font-weight: 600;
	color: #495057;
}

.time-dropdown {
	padding: 10px 16px;
	border: 2px solid #e9ecef;
	border-radius: 8px;
	font-size: 16px;
	background: white;
	cursor: pointer;
	transition: border-color 0.2s ease;
	min-width: 150px;
}

.time-dropdown:focus {
	outline: none;
	border-color: #667eea;
	box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.custom-range {
	margin-top: 16px;
	display: flex;
	align-items: center;
	gap: 12px;
	flex-wrap: wrap;
}

.date-input {
	padding: 8px 12px;
	border: 2px solid #e9ecef;
	border-radius: 6px;
	font-size: 14px;
}

.date-input:focus {
	outline: none;
	border-color: #667eea;
}

.apply-btn {
	background: #28a745;
	color: white;
	border: none;
	padding: 8px 16px;
	border-radius: 6px;
	cursor: pointer;
	font-weight: 600;
	transition: background-color 0.2s ease;
}

.apply-btn:hover {
	background: #218838;
}

/* Chart Container */
.chart-container {
	position: relative;
	padding: 0 24px;
	min-height: 400px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.chart-canvas {
	max-height: 400px;
	width: 100% !important;
	height: 400px !important;
}

/* Loading Indicator */
.loading-indicator {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16px;
	color: #6c757d;
}

.spinner {
	width: 40px;
	height: 40px;
	border: 4px solid #f3f3f3;
	border-top: 4px solid #667eea;
	border-radius: 50%;
	animation: spin 1s linear infinite;
}

/* Error Message */
.error-message {
	display: flex;
	flex-direction: column;
	align-items: center;
	gap: 16px;
	color: #dc3545;
	text-align: center;
}

.error-icon {
	width: 48px;
	height: 48px;
	color: #dc3545;
}

.retry-btn {
	background: #dc3545;
	color: white;
	border: none;
	padding: 10px 20px;
	border-radius: 6px;
	cursor: pointer;
	font-weight: 600;
	transition: background-color 0.2s ease;
}

.retry-btn:hover {
	background: #c82333;
}

/* Chart Stats */
.chart-stats {
	padding: 24px;
	background: #f8f9fa;
	border-top: 1px solid #e9ecef;
	margin-top: 24px;
}

.stats-grid {
	display: grid;
	grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
	gap: 16px;
}

.stat-item {
	display: flex;
	justify-content: space-between;
	align-items: center;
	padding: 12px 16px;
	background: white;
	border-radius: 8px;
	box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.stat-label {
	font-weight: 600;
	color: #495057;
}

.stat-value {
	font-weight: 700;
	font-size: 16px;
}

.stat-value.positive {
	color: #28a745;
}

.stat-value.negative {
	color: #dc3545;
}

/* Animations */
@keyframes fadeIn {
	from {
		opacity: 0;
	}
	to {
		opacity: 1;
	}
}

@keyframes slideIn {
	from {
		opacity: 0;
		transform: translateY(-50px) scale(0.95);
	}
	to {
		opacity: 1;
		transform: translateY(0) scale(1);
	}
}

@keyframes spin {
	0% {
		transform: rotate(0deg);
	}
	100% {
		transform: rotate(360deg);
	}
}

/* Responsive Design */
@media (max-width: 768px) {
	.modal-content {
		width: 95%;
		margin: 10px;
		max-height: 95vh;
	}

	.modal-header {
		padding: 16px 16px 0;
	}

	.modal-header h2 {
		font-size: 20px;
	}

	.time-selector {
		padding: 0 16px 16px;
	}

	.chart-container {
		padding: 0 16px;
		min-height: 300px;
	}

	.chart-canvas {
		height: 300px !important;
	}

	.custom-range {
		flex-direction: column;
		align-items: stretch;
	}

	.stats-grid {
		grid-template-columns: 1fr;
	}

	.chart-stats {
		padding: 16px;
	}
}

@media (max-width: 480px) {
	.price-trends-btn {
		width: 100%;
		justify-content: center;
	}

	.modal-content {
		width: 100%;
		height: 100%;
		border-radius: 0;
		max-height: 100vh;
	}
}
</style>
