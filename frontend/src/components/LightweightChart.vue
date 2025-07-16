<template>
	<div class="relative w-full h-96">
		<!-- Time Range Controls -->
		<div class="flex absolute top-2 left-2 z-20 space-x-1">
			<button
				v-for="period in timePeriods"
				:key="period.value"
				@click="setRange(period.value)"
				:class="[
					'px-3 py-1 text-xs font-medium rounded transition-colors',
					selectedPeriod === period.value
						? 'bg-blue-500 text-white'
						: 'bg-white text-gray-700 hover:bg-gray-100 border border-gray-200',
				]"
			>
				{{ period.label }}
			</button>
		</div>

		<!-- Zoom Controls -->
		<div class="flex absolute top-2 right-2 z-20 space-x-1">
			<button
				@click="resetZoom"
				class="px-3 py-1 text-xs font-medium text-gray-700 bg-white rounded border border-gray-200 transition-colors hover:bg-gray-100"
				title="Reset Zoom"
			>
				<svg
					class="w-3 h-3"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4"
					/>
				</svg>
			</button>
			<button
				@click="zoomIn"
				class="px-3 py-1 text-xs font-medium text-gray-700 bg-white rounded border border-gray-200 transition-colors hover:bg-gray-100"
				title="Zoom In"
			>
				<svg
					class="w-3 h-3"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"
					/>
				</svg>
			</button>
			<button
				@click="zoomOut"
				class="px-3 py-1 text-xs font-medium text-gray-700 bg-white rounded border border-gray-200 transition-colors hover:bg-gray-100"
				title="Zoom Out"
			>
				<svg
					class="w-3 h-3"
					fill="none"
					stroke="currentColor"
					viewBox="0 0 24 24"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7"
					/>
				</svg>
			</button>
		</div>

		<div ref="chartContainer" class="w-full h-full"></div>
		<div
			ref="tooltip"
			class="hidden absolute z-10 px-2 py-1 text-xs text-white bg-black rounded pointer-events-none"
			style="transform: translate(-50%, -100%); margin-top: -8px"
		></div>
	</div>
</template>

<script>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import {
	createChart,
	AreaSeries,
	CandlestickSeries,
	LineSeries,
	HistogramSeries,
	createSeriesMarkers,
	createTextWatermark,
} from 'lightweight-charts'

export default {
	name: 'LightweightChart',
	props: {
		chartData: {
			type: Array,
			required: true,
			default: () => [],
		},
		volumeData: {
			type: Array,
			default: () => [],
		},
		chartType: {
			type: String,
			default: 'line',
			validator: value => ['line', 'candlestick', 'area'].includes(value),
		},
		color: {
			type: String,
			default: 'rgb(59, 130, 246)',
		},
		height: {
			type: Number,
			default: 360,
		},
		showVolume: {
			type: Boolean,
			default: false,
		},
		markers: {
			type: Array,
			default: () => [],
		},
		watermark: {
			type: String,
			default: '',
		},
	},
	setup(props) {
		const chartContainer = ref(null)
		const tooltip = ref(null)
		const selectedPeriod = ref('1h')
		let chart = null
		let series = null
		let volumeSeries = null
		let seriesMarkers = null
		let watermark = null

		const timePeriods = [
			{ label: '1s', value: '1s' },
			{ label: '1m', value: '1m' },
			{ label: '5m', value: '5m' },
			{ label: '1h', value: '1h' },
			{ label: '1d', value: '1d' },
		]

		const prepareData = raw => {
			return [...raw]
				.map(d => ({
					...d,
					time:
						typeof d.time === 'string'
							? Math.floor(new Date(d.time).getTime() / 1000)
							: d.time,
				}))
				.sort((a, b) => a.time - b.time)
				.filter((d, i, arr) => i === 0 || d.time !== arr[i - 1].time)
		}

		const prepareVolumeData = raw => {
			return [...raw]
				.map(d => ({
					...d,
					time:
						typeof d.time === 'string'
							? Math.floor(new Date(d.time).getTime() / 1000)
							: d.time,
					value: d.volume || d.value || 0,
				}))
				.sort((a, b) => a.time - b.time)
				.filter((d, i, arr) => i === 0 || d.time !== arr[i - 1].time)
		}

		const formatNumber = value => {
			if (value === null || value === undefined) return '0'

			if (value >= 1e9) {
				return (value / 1e9).toFixed(2) + 'B'
			} else if (value >= 1e6) {
				return (value / 1e6).toFixed(2) + 'M'
			} else if (value >= 1e3) {
				return (value / 1e3).toFixed(2) + 'K'
			} else {
				return value.toFixed(2)
			}
		}

		const setRange = period => {
			if (!chart) return

			selectedPeriod.value = period

			// Get the full time range of the data
			const timeScale = chart.timeScale()
			const visibleRange = timeScale.getVisibleRange()

			if (!visibleRange) {
				// If no visible range, fit all content
				timeScale.fitContent()
				return
			}

			const now = Math.floor(Date.now() / 1000)
			let fromTime

			// Calculate the time range based on the selected period
			switch (period) {
				case '1s':
					fromTime = now - 5 * 60 // Last 5 minutes
					break
				case '1m':
					fromTime = now - 4 * 60 * 60 // Last 4 hours
					break
				case '5m':
					fromTime = now - 24 * 60 * 60 // Last 24 hours
					break
				case '1h':
					fromTime = now - 7 * 24 * 60 * 60 // Last 7 days
					break
				case '1d':
					fromTime = now - 90 * 24 * 60 * 60 // Last 90 days
					break
				default:
					// Show all data
					timeScale.fitContent()
					return
			}

			// Set the visible range
			timeScale.setVisibleRange({
				from: Math.max(fromTime, visibleRange.from),
				to: now,
			})
		}

		const resetZoom = () => {
			if (!chart) return
			chart.timeScale().fitContent()
		}

		const zoomIn = () => {
			if (!chart) return
			const currentRange = chart.timeScale().getVisibleRange()
			if (currentRange) {
				const timeDiff = currentRange.to - currentRange.from
				const newTimeDiff = timeDiff * 0.7 // Zoom in by 30%
				const center = (currentRange.from + currentRange.to) / 2
				const from = center - newTimeDiff / 2
				const to = center + newTimeDiff / 2
				chart.timeScale().setVisibleRange({ from, to })
			}
		}

		const zoomOut = () => {
			if (!chart) return
			const currentRange = chart.timeScale().getVisibleRange()
			if (currentRange) {
				const timeDiff = currentRange.to - currentRange.from
				const newTimeDiff = timeDiff * 1.3 // Zoom out by 30%
				const center = (currentRange.from + currentRange.to) / 2
				const from = center - newTimeDiff / 2
				const to = center + newTimeDiff / 2
				chart.timeScale().setVisibleRange({ from, to })
			}
		}

		const createChartInstance = () => {
			if (!chartContainer.value) return

			const container = chartContainer.value
			const width = container.clientWidth

			chart = createChart(container, {
				width,
				height: props.height,
				layout: {
					background: { color: '#ffffff' },
					textColor: '#333333',
				},
				grid: {
					vertLines: { color: '#f0f0f0' },
					horzLines: { color: '#f0f0f0' },
				},
				crosshair: {
					mode: 1,
				},
				rightPriceScale: {
					borderColor: '#f0f0f0',
				},
				timeScale: {
					borderColor: '#f0f0f0',
					timeVisible: true,
					secondsVisible: false,
				},
				handleScroll: {
					mouseWheel: true,
					pressedMouseMove: true,
				},
				handleScale: {
					axisPressedMouseMove: true,
					mouseWheel: true,
					pinch: true,
				},
			})

			// Add series based on chart type
			switch (props.chartType) {
				case 'candlestick':
					series = chart.addSeries(CandlestickSeries, {
						upColor: '#00C851', // Bright green for bullish candles
						downColor: '#FF4444', // Bright red for bearish candles
						borderVisible: true, // Show borders for better definition
						borderUpColor: '#00A83A', // Darker green border for bullish
						borderDownColor: '#CC0000', // Darker red border for bearish
						wickUpColor: '#00A83A', // Darker green wicks for bullish
						wickDownColor: '#CC0000', // Darker red wicks for bearish
						wickVisible: true, // Ensure wicks are visible
						priceLineVisible: false, // Hide price line for cleaner look
						// Additional styling for better visibility
						priceFormat: {
							type: 'price',
							precision: 6,
							minMove: 0.000001,
						},
					})
					break
				case 'area':
					series = chart.addSeries(AreaSeries, {
						lineColor: props.color,
						lineWidth: 2,
						topColor: props.color,
						bottomColor: 'rgba(59, 130, 246, 0.1)',
					})
					break
				default: // line
					series = chart.addSeries(LineSeries, {
						color: props.color,
						lineWidth: 2,
					})
			}

			// Add volume histogram if enabled and data available
			if (props.showVolume && props.volumeData.length > 0) {
				volumeSeries = chart.addSeries(HistogramSeries, {
					priceFormat: { type: 'volume' },
					priceScaleId: '',
					scaleMargins: { top: 0.7, bottom: 0 },
					color: '#26a69a', // Default green color
				})

				// Prepare volume data with proper color coding
				const volumeDataWithColors = prepareVolumeData(props.volumeData).map(
					(item, index, array) => {
						// Color volume bars based on price movement
						// Green if price went up, red if price went down
						let color = '#26a69a' // Default green

						if (
							index > 0 &&
							props.chartData[index] &&
							props.chartData[index - 1]
						) {
							const currentPrice =
								props.chartData[index].close || props.chartData[index].price
							const previousPrice =
								props.chartData[index - 1].close ||
								props.chartData[index - 1].price

							if (currentPrice < previousPrice) {
								color = '#FF4444' // Red for down movement
							} else {
								color = '#00C851' // Green for up movement
							}
						}

						return {
							...item,
							color: color,
						}
					}
				)

				volumeSeries.setData(volumeDataWithColors)
			}

			// Set initial data
			if (props.chartData.length > 0) {
				series.setData(prepareData(props.chartData))
			}

			// Add markers if provided
			if (props.markers && props.markers.length > 0) {
				seriesMarkers = createSeriesMarkers(series)
				seriesMarkers.setMarkers(props.markers)
			}

			// Add watermark if provided
			if (props.watermark) {
				const pane = chart.panes()[0]
				watermark = createTextWatermark(pane, {
					horzAlign: 'center',
					vertAlign: 'center',
					lines: [
						{
							text: props.watermark,
							color: 'rgba(0,0,0,0.05)',
							fontSize: 48,
							fontFamily: 'Arial, sans-serif',
						},
					],
				})
			}

			// Setup tooltip
			chart.subscribeCrosshairMove(param => {
				if (!param.time || param.point.x < 0) {
					tooltip.value.style.display = 'none'
					return
				}

				const data = param.seriesData.get(series)
				if (data) {
					const price = data.value ?? data.close ?? data.high ?? 0

					const tooltipContent = `<div class="font-semibold">$${price.toFixed(
						6
					)}</div>`

					tooltip.value.innerHTML = tooltipContent
					tooltip.value.style.left = param.point.x + 'px'
					tooltip.value.style.top = param.point.y + 'px'
					tooltip.value.style.display = 'block'
				}
			})

			// Handle window resize
			const handleResize = () => {
				if (chart && chartContainer.value) {
					chart.applyOptions({ width: chartContainer.value.clientWidth })
				}
			}

			window.addEventListener('resize', handleResize)

			// Cleanup function
			return () => {
				window.removeEventListener('resize', handleResize)
				if (chart) {
					chart.remove()
				}
			}
		}

		onMounted(() => {
			createChartInstance()
		})

		watch(
			() => props.chartData,
			newData => {
				if (series && newData.length > 0) {
					const clean = prepareData(newData)
					series.setData(clean)
				}
			},
			{ deep: true }
		)

		watch(
			() => props.volumeData,
			newData => {
				if (volumeSeries && newData.length > 0) {
					const clean = prepareVolumeData(newData)
					volumeSeries.setData(clean)
				}
			},
			{ deep: true }
		)

		watch(
			() => props.markers,
			newMarkers => {
				if (seriesMarkers) {
					seriesMarkers.setMarkers(newMarkers)
				} else if (series && newMarkers && newMarkers.length > 0) {
					seriesMarkers = createSeriesMarkers(series)
					seriesMarkers.setMarkers(newMarkers)
				}
			},
			{ deep: true }
		)

		watch(
			() => props.chartType,
			() => {
				if (chart) {
					chart.remove()
					createChartInstance()
				}
			}
		)

		onUnmounted(() => {
			if (chart) {
				chart.remove()
			}
		})

		return {
			chartContainer,
			tooltip,
			selectedPeriod,
			timePeriods,
			setRange,
			resetZoom,
			zoomIn,
			zoomOut,
		}
	},
}
</script>
