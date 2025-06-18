<template>
  <aside class="w-64 h-screen bg-white border-r border-gray-200">
    <div class="p-4">
      <h2 class="mb-4 text-xl font-bold text-gray-800">Menu</h2>
      <nav class="space-y-2">
        <button
          @click="$emit('change-view', 'telegram')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'telegram' }"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.07-.2-.08-.06-.19-.04-.27-.02-.12.02-1.96 1.25-5.54 3.69-.52.36-1 .53-1.42.52-.47-.01-1.37-.26-2.03-.48-.82-.27-1.47-.42-1.42-.88.03-.24.29-.49.8-.75 3.12-1.36 5.2-2.26 6.24-2.7 2.98-1.24 3.6-1.46 4.01-1.46.09 0 .28.02.4.09.11.06.19.14.22.25.02.1.02.21.01.3z"/>
          </svg>
          <span>Telegram Analysis</span>
        </button>
        <button
          @click="$emit('change-view', 'market')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'market' }"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M3.5 18.49l6-6.01 4 4L22 6.92l-1.41-1.41-7.09 7.97-4-4L2 16.99z"/>
          </svg>
          <span>Market Analysis</span>
        </button>
        <button
          @click="$emit('change-view', 'solana-wallets')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'solana-wallets' }"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <rect x="2" y="7" width="20" height="10" rx="2" stroke-width="2" stroke="currentColor"/>
            <path d="M16 12h.01" stroke-width="2" stroke-linecap="round" stroke="currentColor"/>
          </svg>
          <span>Find Solana wallets</span>
        </button>
        <button
          @click="$emit('change-view', 'portfolio')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'portfolio' }"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <span>Portfolio</span>
        </button>
        <button
          @click="$emit('change-view', 'arbitration')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'arbitration' }"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          <span>Arbitration</span>
        </button>
        <button
          @click="$emit('change-view', 'sniping')"
          class="flex items-center w-full px-4 py-2 space-x-2 text-gray-700 transition-colors rounded-lg hover:bg-gray-100"
          :class="{ 'bg-gray-100': currentView === 'sniping' }"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
          <span>Sniping New Tokens</span>
        </button>
      </nav>
    </div>
  </aside>
</template>

<script>
export default {
  name: 'AppSidebar',
  props: {
    currentView: {
      type: String,
      required: true
    }
  },
  emits: ['change-view']
}
</script> 