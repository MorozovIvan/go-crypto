<template>
	<div class="card bg-base-100 shadow-xl">
		<div class="card-body">
			<div class="flex justify-between items-center">
				<h2 class="card-title text-base">{{ title }}</h2>
				<div class="badge" :class="indicatorBadgeClass">
					{{ error ? 'Error' : indicator }}
				</div>
			</div>

			<div class="stat-value text-2xl">
				<span v-if="!error">{{ value }}</span>
				<span v-else class="text-error">Error</span>
			</div>

			<div class="h-24 mt-2">
				<Line
					v-if="!error && chartData?.length"
					:data="chartConfig"
					:options="chartOptions"
					height="96"
				/>
				<div v-else-if="error" class="flex items-center justify-center h-full">
					<div class="alert alert-error alert-sm">
						<span class="text-xs">Data unavailable</span>
					</div>
				</div>
				<div v-else class="flex items-center justify-center h-full opacity-50">
					<span class="text-xs">No chart data</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
import { Line } from 'vue-chartjs'

export default {
	name: 'MetricCard',
	components: { Line },
	props: {
		title: String,
		value: [String, Number, null],
		indicator: String,
		chartData: Array,
		chartLabels: Array,
		error: Boolean,
	},
	computed: {
		indicatorBadgeClass() {
			if (this.error) return 'badge-error'
			switch (this.indicator) {
				case 'Buy':
					return 'badge-success'
				case 'Sell':
					return 'badge-error'
				case 'Hold':
					return 'badge-warning'
				default:
					return 'badge-neutral'
			}
		},
		chartConfig() {
			return {
				labels: this.chartLabels,
				datasets: [
					{
						label: this.title,
						data: this.chartData,
						borderColor: '#0a78b9',
						backgroundColor: 'rgba(10,120,185,0.1)',
						tension: 0.3,
						fill: true,
					},
				],
			}
		},
		chartOptions() {
			return {
				responsive: true,
				plugins: {
					legend: { display: false },
					title: { display: false },
				},
				scales: {
					x: { display: false },
					y: { display: false },
				},
			}
		},
	},
}
</script>
