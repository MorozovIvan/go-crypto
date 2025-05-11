<template>
  <div class="flex min-h-screen bg-gray-50">
    <!-- Sidebar -->
    <aside class="w-64 bg-white border-r flex flex-col">
      <div class="p-6 text-xl font-bold border-b">Crypto Analytics</div>
      <nav class="flex-1 p-4 space-y-2">
        <button class="w-full text-left px-4 py-2 rounded hover:bg-blue-100 flex items-center gap-2" :class="{ 'bg-blue-200 font-semibold': currentView === 'telegram' }">
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-9.6-3.2a1 1 0 00-1.3 1.3l3.2 9.6a1 1 0 001.3.6l9.6-3.2a1 1 0 00.6-1.3l-3.2-9.6a1 1 0 00-1.3-.6z" /></svg>
          Telegram Analysis
        </button>
        <button class="w-full text-left px-4 py-2 rounded hover:bg-blue-100 flex items-center gap-2" :class="{ 'bg-blue-200 font-semibold': currentView === 'market' }">
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3v18h18" /></svg>
          Market Analysis
        </button>
        <button class="w-full text-left px-4 py-2 rounded hover:bg-blue-100 flex items-center gap-2" :class="{ 'bg-blue-200 font-semibold': currentView === 'wallets' }">
          <!-- Wallet SVG Icon -->
          <svg class="w-5 h-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <rect x="2" y="7" width="20" height="10" rx="2" stroke-width="2" stroke="currentColor"/>
            <path d="M16 12h.01" stroke-width="2" stroke-linecap="round" stroke="currentColor"/>
          </svg>
          Profitable Wallets
        </button>
      </nav>
      <div class="p-4 border-t text-xs text-gray-500">Footer</div>
    </aside>
    <!-- ... existing code ... -->
  </div>
</template>

<script>
// ... existing code ...
</script>

<style scoped>
.font-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}
</style> 