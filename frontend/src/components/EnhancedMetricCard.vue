<template>
	<div
		class="card bg-base-100 shadow-xl hover:shadow-2xl transition-all duration-300 group"
	>
		<!-- Header -->
		<div class="card-body">
			<div class="flex justify-between items-start">
				<div class="flex-1">
					<h2
						class="card-title text-base group-hover:text-primary transition-colors"
					>
						{{ title }}
					</h2>
					<div class="flex items-center gap-2 mt-1">
						<span class="text-xs opacity-70"
							>Weight: {{ (weight * 100).toFixed(1) }}%</span
						>
						<progress
							class="progress progress-primary w-12 h-1"
							:value="weight * 100"
							max="100"
						></progress>
					</div>
				</div>

				<div class="flex items-center gap-2">
					<!-- Status Indicator -->
					<div v-if="loading" class="loading loading-spinner loading-sm"></div>
					<div v-else-if="error" class="badge badge-error badge-xs"></div>
					<div v-else class="badge badge-success badge-xs"></div>

					<!-- Refresh Button -->
					<button
						@click="$emit('refresh')"
						:disabled="loading"
						class="btn btn-ghost btn-xs btn-circle"
					>
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
							></path>
						</svg>
					</button>
				</div>
			</div>

			<!-- Content -->
			<div v-if="loading" class="flex items-center justify-center h-24">
				<div class="text-center">
					<div class="skeleton h-8 w-20 mx-auto mb-2"></div>
					<div class="skeleton h-4 w-16 mx-auto"></div>
				</div>
			</div>

			<div
				v-else-if="error"
				class="text-center h-24 flex items-center justify-center"
			>
				<div class="alert alert-error">
					<svg
						class="w-6 h-6"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
						></path>
					</svg>
					<div>
						<div class="font-bold">Data Error</div>
						<div class="text-xs">Unable to fetch data</div>
					</div>
				</div>
			</div>

			<div v-else class="space-y-4">
				<!-- Value Display -->
				<div class="text-center">
					<div class="stat-value text-2xl">
						{{ formatValue(value) }}
					</div>
					<div class="flex items-center justify-center gap-2 mt-2">
						<div class="badge" :class="indicatorBadgeClass">
							{{ indicator }}
						</div>
						<div v-if="lastUpdated" class="text-xs opacity-70">
							{{ formatTime(lastUpdated) }}
						</div>
					</div>
				</div>

				<!-- Score Visualization -->
				<div class="bg-base-200 rounded-lg p-3">
					<div class="flex justify-between items-center mb-2">
						<span class="text-xs opacity-70">Score</span>
						<span class="text-xs font-mono font-bold" :class="scoreColorClass">
							{{ score.toFixed(3) }}
						</span>
					</div>

					<!-- Score Bar -->
					<div class="relative">
						<div class="w-full bg-base-300 rounded-full h-2">
							<div class="absolute inset-0 flex items-center justify-center">
								<div class="w-px h-3 bg-base-content opacity-30"></div>
							</div>
							<div
								class="h-2 rounded-full transition-all duration-500"
								:class="scoreBarClass"
								:style="scoreBarStyle"
							></div>
						</div>
						<div class="flex justify-between text-xs opacity-60 mt-1">
							<span>-1.0</span>
							<span>0</span>
							<span>+1.0</span>
						</div>
					</div>
				</div>

				<!-- Mini Chart -->
				<div class="h-16">
					<Line
						v-if="chartData && chartData.length > 0"
						:data="chartConfig"
						:options="chartOptions"
						:height="64"
					/>
					<div
						v-else
						class="h-16 bg-base-200 rounded flex items-center justify-center"
					>
						<span class="text-xs opacity-60">No chart data</span>
					</div>
				</div>

				<!-- Contribution Display -->
				<div class="text-center pt-2 border-t border-base-300">
					<div class="text-xs opacity-70">Contribution to Signal</div>
					<div class="text-sm font-bold" :class="contributionColorClass">
						{{ (score * weight).toFixed(4) }}
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { Line } from 'vue-chartjs'

export default {
	name: 'EnhancedMetricCard',
	components: { Line },
	props: {
		title: String,
		value: [String, Number, null],
		indicator: String,
		score: { type: Number, default: 0 },
		weight: { type: Number, default: 0 },
		chartData: Array,
		chartLabels: Array,
		error: Boolean,
		loading: Boolean,
		lastUpdated: String,
	},
	emits: ['refresh'],
	computed: {
		indicatorBadgeClass() {
			if (this.error) return 'badge-error';
			switch (this.indicator) {
				case 'Buy': return 'badge-success';
				case 'Sell': return 'badge-error';
				case 'Hold': return 'badge-warning';
				default: return 'badge-neutral';
			}
		},
		scoreColorClass() {
			if (this.score > 0.3) return 'text-success'
			if (this.score < -0.3) return 'text-error'
			return 'text-base-content'
		},
		contributionColorClass() {
			const contribution = this.score * this.weight
			if (contribution > 0.05) return 'text-success'
			if (contribution < -0.05) return 'text-error'
			return 'text-base-content'
		},
		scoreBarClass() {
			if (this.score > 0) return 'bg-success'
			if (this.score < 0) return 'bg-error'
			return 'bg-base-content'
		},
		scoreBarStyle() {
			const normalizedScore = Math.max(-1, Math.min(1, this.score))
			const percentage = ((normalizedScore + 1) / 2) * 100

			if (normalizedScore >= 0) {
				return {
					width: `${percentage - 50}%`,
					marginLeft: '50%',
				}
			} else {
				return {
					width: `${50 - percentage}%`,
					marginLeft: `${percentage}%`,
				}
			}
		},
		chartConfig() {
			return {
				labels: this.chartLabels || [],
				datasets: [
					{
						label: this.title,
						data: this.chartData || [],
						borderColor: this.getChartColor(),
						backgroundColor: this.getChartBackgroundColor(),
						tension: 0.4,
						fill: true,
						pointRadius: 0,
						borderWidth: 2,
					},
				],
			}
		},
		chartOptions() {
			return {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					title: { display: false },
					tooltip: { enabled: false },
				},
				scales: {
					x: { display: false },
					y: { display: false },
				},
				elements: {
					point: { radius: 0 },
				},
			}
		},
	},
	methods: {
		formatValue(value) {
			if (value === null || value === undefined) return 'N/A'
			if (typeof value === 'number') {
				if (value > 1000000) return (value / 1000000).toFixed(1) + 'M'
				if (value > 1000) return (value / 1000).toFixed(1) + 'K'
				return value.toFixed(2)
			}
			return String(value)
		},
		formatTime(timestamp) {
			if (!timestamp) return ''
			const date = new Date(timestamp)
			const now = new Date()
			const diffMs = now - date
			const diffMins = Math.floor(diffMs / 60000)

			if (diffMins < 1) return 'Just now'
			if (diffMins < 60) return `${diffMins}m ago`
			if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
			return date.toLocaleDateString()
		},
		getChartColor() {
			if (this.score > 0.3) return '#10b981' // green
			if (this.score < -0.3) return '#ef4444' // red
			return '#6b7280' // gray
		},
		getChartBackgroundColor() {
			if (this.score > 0.3) return 'rgba(16, 185, 129, 0.1)'
			if (this.score < -0.3) return 'rgba(239, 68, 68, 0.1)'
			return 'rgba(107, 114, 128, 0.1)'
		},
	},
}
</script>
