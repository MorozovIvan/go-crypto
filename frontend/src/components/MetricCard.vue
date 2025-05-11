<template>
  <div class="bg-white rounded-lg shadow p-4 flex flex-col">
    <div class="flex justify-between items-center mb-2">
      <h4 class="font-semibold">{{ title }}</h4>
      <span :class="indicatorClass">{{ error ? 'Error' : indicator }}</span>
    </div>
    <div class="text-2xl font-bold mb-2">
      <span v-if="!error">{{ value }}</span>
      <span v-else class="text-red-600">Error</span>
    </div>
    <Line v-if="!error" :data="chartConfig" :options="chartOptions" height="120" />
    <div v-else class="text-xs text-red-500">Data unavailable</div>
  </div>
</template>

<script>
import { Line } from 'vue-chartjs';

export default {
  name: 'MetricCard',
  components: { Line },
  props: {
    title: String,
    value: [String, Number, null],
    indicator: String,
    chartData: Array,
    chartLabels: Array,
    error: Boolean
  },
  computed: {
    indicatorClass() {
      return {
        'text-green-600': this.indicator === 'Buy' && !this.error,
        'text-red-600': this.indicator === 'Sell' && !this.error,
        'text-gray-500': this.indicator === 'Hold' && !this.error,
        'text-red-600': this.error,
        'font-bold': true
      };
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
            fill: true
          }
        ]
      };
    },
    chartOptions() {
      return {
        responsive: true,
        plugins: {
          legend: { display: false },
          title: { display: false }
        },
        scales: {
          x: { display: false },
          y: { display: false }
        }
      };
    }
  }
};
</script> 