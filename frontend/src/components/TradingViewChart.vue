<template>
  <div>
    <div v-if="!chartData || chartData.length === 0" class="text-gray-500 p-4 text-center">
      <p>No chart data available</p>
      <p class="text-sm mt-1">This token may not have recent trading activity</p>
    </div>
    <div v-else-if="chartData.length === 1" class="text-gray-500 p-4 text-center">
      <p>Limited data available ({{ chartData.length }} data point)</p>
      <p class="text-sm mt-1">Current price: ${{ chartData[0]?.price?.toFixed(8) || 'N/A' }}</p>
      <p class="text-sm">Market cap: ${{ chartData[0]?.market_cap?.toLocaleString() || 'N/A' }}</p>
    </div>
    <div v-else-if="chartData.length < 5" class="text-gray-500 p-4 text-center">
      <p>Limited chart data ({{ chartData.length }} data points)</p>
      <p class="text-sm mt-1">Current price: ${{ chartData[chartData.length - 1]?.price?.toFixed(8) || 'N/A' }}</p>
      <p class="text-sm">Market cap: ${{ chartData[chartData.length - 1]?.market_cap?.toLocaleString() || 'N/A' }}</p>
    </div>
    <div v-else class="relative">
      <canvas ref="chartContainer" :style="{ height: chartData.length < 10 ? '300px' : '400px' }"></canvas>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, onBeforeUnmount } from 'vue'
import { Chart, registerables } from 'chart.js'
import 'chartjs-adapter-date-fns'

// Register all Chart.js components
Chart.register(...registerables)

const props = defineProps({
  chartData: {
    type: Array,
    required: true,
    default: () => [],
  },
})

const chartContainer = ref(null)
let chart = null

function formatData(data) {
  if (!data || !Array.isArray(data)) return []
  
  return data.map(d => ({
    x: new Date(d.timestamp), // Date object for time scale
    y: d.market_cap ?? d.price ?? 0,
  }))
}

function createChartInstance() {
  if (!chartContainer.value || !props.chartData || props.chartData.length < 2) return
  
  try {
    console.log('Creating Chart.js chart with data points:', props.chartData.length)
    
    // Clean up existing chart if any
    if (chart) {
      chart.destroy()
      chart = null
    }
    
    const ctx = chartContainer.value.getContext('2d')
    const formattedData = formatData(props.chartData)
    
    console.log('Formatted data:', formattedData.slice(0, 3))
    
    chart = new Chart(ctx, {
      type: 'line',
      data: {
        datasets: [{
          label: 'Market Cap',
          data: formattedData,
          borderColor: '#3b82f6',
          backgroundColor: 'rgba(59, 130, 246, 0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0.4,
          pointRadius: props.chartData.length < 10 ? 4 : 2,
          pointHoverRadius: 6,
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          intersect: false,
          mode: 'index',
        },
        plugins: {
          legend: {
            display: false,
          },
          tooltip: {
            callbacks: {
              label: function(context) {
                const value = context.parsed.y
                return `Market Cap: $${value.toLocaleString()}`
              }
            }
          }
        },
        scales: {
          x: {
            type: 'time',
            time: {
              unit: props.chartData.length < 10 ? 'hour' : 'day',
              displayFormats: {
                hour: 'MMM dd, HH:mm',
                day: 'MMM dd'
              }
            },
            grid: {
              color: '#f3f4f6',
            },
            ticks: {
              color: '#6b7280',
            }
          },
          y: {
            beginAtZero: false,
            grid: {
              color: '#f3f4f6',
            },
            ticks: {
              color: '#6b7280',
              callback: function(value) {
                return '$' + value.toLocaleString()
              }
            }
          }
        }
      }
    })
    
    console.log('Chart.js chart created successfully')
  } catch (error) {
    console.error('Error creating Chart.js chart:', error)
  }
}

// Watch for changes in chart data
watch(() => props.chartData, () => {
  if (props.chartData && props.chartData.length > 0) {
    createChartInstance()
  }
}, { deep: true })

onMounted(() => {
  if (props.chartData && props.chartData.length > 0) {
    createChartInstance()
  }
})

onBeforeUnmount(() => {
  if (chart) {
    chart.destroy()
    chart = null
  }
})
</script>

<style scoped>
/* Additional styling if needed */
</style> 