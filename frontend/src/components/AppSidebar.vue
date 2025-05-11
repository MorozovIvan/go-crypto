<template>
  <aside class="w-64 bg-white border-r border-gray-200 h-screen">
    <div class="p-4">
      <h2 class="text-xl font-bold text-gray-800 mb-4">Menu</h2>
      <nav class="space-y-2">
        <button
          @click="$emit('change-view', 'telegram')"
          class="w-full flex items-center space-x-2 px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          :class="{ 'bg-gray-100': currentView === 'telegram' }"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.07-.2-.08-.06-.19-.04-.27-.02-.12.02-1.96 1.25-5.54 3.69-.52.36-1 .53-1.42.52-.47-.01-1.37-.26-2.03-.48-.82-.27-1.47-.42-1.42-.88.03-.24.29-.49.8-.75 3.12-1.36 5.2-2.26 6.24-2.7 2.98-1.24 3.6-1.46 4.01-1.46.09 0 .28.02.4.09.11.06.19.14.22.25.02.1.02.21.01.3z"/>
          </svg>
          <span>Telegram Analysis</span>
        </button>
        <button
          @click="$emit('change-view', 'market')"
          class="w-full flex items-center space-x-2 px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
          :class="{ 'bg-gray-100': currentView === 'market' }"
        >
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
            <path d="M3.5 18.49l6-6.01 4 4L22 6.92l-1.41-1.41-7.09 7.97-4-4L2 16.99z"/>
          </svg>
          <span>Market Analysis</span>
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