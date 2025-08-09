/**
 * Chart.js Configuration for Portfolio Price Trends
 * This is the core configuration used for rendering the price change graph
 */

const portfolioPriceChartConfig = {
	type: 'line',
	data: {
		datasets: [
			// Dynamic datasets will be generated based on portfolio tokens
			// Example dataset structure:
			{
				label: 'BTC',
				data: [
					{ x: '2024-01-01T00:00:00Z', y: 0 },
					{ x: '2024-01-02T00:00:00Z', y: 2.5 },
					{ x: '2024-01-03T00:00:00Z', y: -1.2 },
					{ x: '2024-01-04T00:00:00Z', y: 4.8 },
					{ x: '2024-01-05T00:00:00Z', y: 3.2 },
				],
				borderColor: '#FF6384',
				backgroundColor: '#FF638420',
				fill: false,
				tension: 0.4,
			},
		],
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
				text: 'Portfolio Price Changes - Last 7 Days',
				font: {
					size: 16,
					weight: 'bold',
				},
				color: '#2c3e50',
			},
			legend: {
				display: true,
				position: 'top',
				labels: {
					usePointStyle: true,
					padding: 20,
					font: {
						size: 12,
					},
					color: '#495057',
				},
			},
			tooltip: {
				backgroundColor: 'rgba(0, 0, 0, 0.8)',
				titleColor: 'white',
				bodyColor: 'white',
				borderColor: 'rgba(255, 255, 255, 0.2)',
				borderWidth: 1,
				cornerRadius: 8,
				displayColors: true,
				callbacks: {
					title: function (tooltipItems) {
						// Format the date/time for tooltip title
						const date = new Date(tooltipItems[0].parsed.x)
						return date.toLocaleDateString() + ' ' + date.toLocaleTimeString()
					},
					label: function (context) {
						const value = context.parsed.y
						const sign = value >= 0 ? '+' : ''
						return `${context.dataset.label}: ${sign}${value.toFixed(2)}%`
					},
					afterBody: function (tooltipItems) {
						// Add additional context if needed
						return tooltipItems.length > 1
							? `Showing ${tooltipItems.length} tokens`
							: ''
					},
				},
			},
		},
		scales: {
			x: {
				type: 'time',
				time: {
					displayFormats: {
						minute: 'HH:mm',
						hour: 'HH:mm',
						day: 'MMM dd',
						week: 'MMM dd',
						month: 'MMM yyyy',
					},
					tooltipFormat: 'MMM dd, yyyy HH:mm',
				},
				title: {
					display: true,
					text: 'Time',
					font: {
						size: 14,
						weight: 'bold',
					},
					color: '#495057',
				},
				grid: {
					color: 'rgba(0, 0, 0, 0.1)',
					drawBorder: false,
				},
				ticks: {
					color: '#6c757d',
					maxTicksLimit: 8,
				},
			},
			y: {
				title: {
					display: true,
					text: 'Price Change (%)',
					font: {
						size: 14,
						weight: 'bold',
					},
					color: '#495057',
				},
				grid: {
					color: 'rgba(0, 0, 0, 0.1)',
					drawBorder: false,
				},
				ticks: {
					color: '#6c757d',
					callback: function (value) {
						return value.toFixed(1) + '%'
					},
				},
				// Add zero line for reference
				afterBuildTicks: function (scale) {
					if (!scale.ticks.some(tick => tick.value === 0)) {
						scale.ticks.push({ value: 0, major: false })
						scale.ticks.sort((a, b) => a.value - b.value)
					}
				},
			},
		},
		elements: {
			line: {
				tension: 0.4,
				borderWidth: 2,
			},
			point: {
				radius: 3,
				hoverRadius: 6,
				borderWidth: 2,
				hoverBorderWidth: 3,
			},
		},
		// Animation configuration
		animation: {
			duration: 1000,
			easing: 'easeInOutQuart',
		},
		// Responsive breakpoints
		onResize: function (chart, size) {
			// Adjust legend position on mobile
			if (size.width < 768) {
				chart.options.plugins.legend.position = 'bottom'
				chart.options.plugins.legend.labels.padding = 10
			} else {
				chart.options.plugins.legend.position = 'top'
				chart.options.plugins.legend.labels.padding = 20
			}
		},
	},
}

// Color palette for different tokens
const tokenColors = [
	'#FF6384', // Red
	'#36A2EB', // Blue
	'#FFCE56', // Yellow
	'#4BC0C0', // Teal
	'#9966FF', // Purple
	'#FF9F40', // Orange
	'#FF6384', // Pink
	'#C9CBCF', // Grey
	'#4BC0C0', // Cyan
	'#FF6384', // Rose
	'#36A2EB', // Sky Blue
	'#FFCE56', // Gold
	'#4BC0C0', // Aqua
	'#9966FF', // Violet
	'#FF9F40', // Amber
]

// Helper function to generate dataset for a token
function createTokenDataset(tokenId, priceData, colorIndex) {
	const color = tokenColors[colorIndex % tokenColors.length]

	// Calculate percentage changes from first price
	const basePrice = priceData[0].price
	const data = priceData.map(point => ({
		x: new Date(point.timestamp),
		y: ((point.price - basePrice) / basePrice) * 100,
	}))

	return {
		label: tokenId,
		data: data,
		borderColor: color,
		backgroundColor: color + '20', // 20% opacity for fill
		fill: false,
		tension: 0.4,
		pointBackgroundColor: color,
		pointBorderColor: '#ffffff',
		pointHoverBackgroundColor: '#ffffff',
		pointHoverBorderColor: color,
	}
}

// Export configuration for use in main application
if (typeof module !== 'undefined' && module.exports) {
	module.exports = {
		portfolioPriceChartConfig,
		tokenColors,
		createTokenDataset,
	}
}
