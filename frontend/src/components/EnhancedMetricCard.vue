<template>
  <div class="bg-white rounded-2xl shadow-lg border border-gray-200 overflow-hidden hover:shadow-xl transition-all duration-300 group">
    <!-- Header -->
    <div class="p-4 border-b border-gray-100">
      <div class="flex justify-between items-start">
        <div class="flex-1">
          <h4 class="font-semibold text-gray-900 group-hover:text-blue-600 transition-colors">{{ title }}</h4>
          <div class="flex items-center space-x-2 mt-1">
            <span class="text-xs text-gray-500">Weight: {{ (weight * 100).toFixed(1) }}%</span>
            <div class="w-12 bg-gray-200 rounded-full h-1">
              <div class="bg-blue-500 h-1 rounded-full transition-all duration-300" 
                   :style="{ width: (weight * 100) + '%' }"></div>
            </div>
          </div>
        </div>
        
        <div class="flex items-center space-x-2">
          <!-- Status Indicator -->
          <div v-if="loading" class="animate-spin w-4 h-4 border-2 border-blue-500 border-t-transparent rounded-full"></div>
          <div v-else-if="error" class="w-4 h-4 bg-red-500 rounded-full"></div>
          <div v-else class="w-4 h-4 bg-green-500 rounded-full"></div>
          
          <!-- Refresh Button -->
          <button 
            @click="$emit('refresh')"
            :disabled="loading"
            class="p-1 rounded-full hover:bg-gray-100 transition-colors disabled:opacity-50"
          >
            <svg class="w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-4">
      <div v-if="loading" class="flex items-center justify-center h-24">
        <div class="text-center">
          <div class="animate-pulse bg-gray-200 rounded h-8 w-20 mx-auto mb-2"></div>
          <div class="animate-pulse bg-gray-200 rounded h-4 w-16 mx-auto"></div>
        </div>
      </div>
      
      <div v-else-if="error" class="text-center h-24 flex items-center justify-center">
        <div>
          <div class="text-red-500 text-2xl mb-2">⚠️</div>
          <div class="text-red-600 text-sm font-medium">Data Error</div>
          <div class="text-xs text-gray-500 mt-1">Unable to fetch data</div>
        </div>
      </div>
      
      <div v-else class="space-y-3">
        <!-- Value Display -->
        <div class="text-center">
          <div class="text-2xl font-bold text-gray-900">
            {{ formatValue(value) }}
          </div>
          <div class="flex items-center justify-center space-x-2 mt-1">
            <span class="px-2 py-1 rounded-full text-xs font-medium" :class="indicatorBadgeClass">
              {{ indicator }}
            </span>
            <span class="text-xs text-gray-500" v-if="lastUpdated">
              {{ formatTime(lastUpdated) }}
            </span>
          </div>
        </div>

        <!-- Score Visualization -->
        <div class="bg-gray-50 rounded-lg p-3">
          <div class="flex justify-between items-center mb-2">
            <span class="text-xs text-gray-600">Score</span>
            <span class="text-xs font-mono font-bold" :class="scoreColorClass">
              {{ score.toFixed(3) }}
            </span>
          </div>
          
          <!-- Score Bar -->
          <div class="relative">
            <div class="w-full bg-gray-200 rounded-full h-2">
              <div class="absolute inset-0 flex items-center justify-center">
                <div class="w-px h-3 bg-gray-400"></div>
              </div>
              <div 
                class="h-2 rounded-full transition-all duration-500"
                :class="scoreBarClass"
                :style="scoreBarStyle"
              ></div>
            </div>
            <div class="flex justify-between text-xs text-gray-500 mt-1">
              <span>-1.0</span>
              <span>0</span>
              <span>+1.0</span>
            </div>
          </div>
        </div>

        <!-- Mini Chart -->
        <div class="h-16">
          <Line v-if="chartData && chartData.length > 0" 
                :data="chartConfig" 
                :options="chartOptions" 
                :height="64" />
          <div v-else class="h-16 bg-gray-100 rounded flex items-center justify-center">
            <span class="text-xs text-gray-500">No chart data</span>
          </div>
        </div>

        <!-- Contribution Display -->
        <div class="text-center pt-2 border-t border-gray-100">
          <div class="text-xs text-gray-600">Contribution to Signal</div>
          <div class="text-sm font-bold" :class="contributionColorClass">
            {{ (score * weight).toFixed(4) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { Line } from 'vue-chartjs';

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
    lastUpdated: String
  },
  emits: ['refresh'],
  computed: {
    indicatorBadgeClass() {
      return {
        'bg-green-100 text-green-800': this.indicator === 'Buy' && !this.error,
        'bg-red-100 text-red-800': this.indicator === 'Sell' && !this.error,
        'bg-yellow-100 text-yellow-800': this.indicator === 'Hold' && !this.error,
        'bg-gray-100 text-gray-800': this.error
      };
    },
    scoreColorClass() {
      if (this.score > 0.3) return 'text-green-600';
      if (this.score < -0.3) return 'text-red-600';
      return 'text-gray-600';
    },
    contributionColorClass() {
      const contribution = this.score * this.weight;
      if (contribution > 0.05) return 'text-green-600';
      if (contribution < -0.05) return 'text-red-600';
      return 'text-gray-600';
    },
    scoreBarClass() {
      if (this.score > 0) return 'bg-green-500';
      if (this.score < 0) return 'bg-red-500';
      return 'bg-gray-400';
    },
    scoreBarStyle() {
      const normalizedScore = Math.max(-1, Math.min(1, this.score));
      const percentage = ((normalizedScore + 1) / 2) * 100;
      
      if (normalizedScore >= 0) {
        return {
          width: `${(percentage - 50)}%`,
          marginLeft: '50%'
        };
      } else {
        return {
          width: `${50 - (percentage)}%`,
          marginLeft: `${percentage}%`
        };
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
            borderWidth: 2
          }
        ]
      };
    },
    chartOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          title: { display: false },
          tooltip: { enabled: false }
        },
        scales: {
          x: { display: false },
          y: { display: false }
        },
        elements: {
          point: { radius: 0 }
        }
      };
    }
  },
  methods: {
    formatValue(value) {
      if (value === null || value === undefined) return 'N/A';
      if (typeof value === 'number') {
        if (value > 1000000) return (value / 1000000).toFixed(1) + 'M';
        if (value > 1000) return (value / 1000).toFixed(1) + 'K';
        return value.toFixed(2);
      }
      return String(value);
    },
    formatTime(timestamp) {
      if (!timestamp) return '';
      const date = new Date(timestamp);
      const now = new Date();
      const diffMs = now - date;
      const diffMins = Math.floor(diffMs / 60000);
      
      if (diffMins < 1) return 'Just now';
      if (diffMins < 60) return `${diffMins}m ago`;
      if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`;
      return date.toLocaleDateString();
    },
    getChartColor() {
      if (this.score > 0.3) return '#10b981'; // green
      if (this.score < -0.3) return '#ef4444'; // red
      return '#6b7280'; // gray
    },
    getChartBackgroundColor() {
      if (this.score > 0.3) return 'rgba(16, 185, 129, 0.1)';
      if (this.score < -0.3) return 'rgba(239, 68, 68, 0.1)';
      return 'rgba(107, 114, 128, 0.1)';
    }
  }
};
</script> 