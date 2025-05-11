<template>
  <div class="p-6">
    <h1 class="text-2xl font-bold mb-6">Market Analysis</h1>
    <div class="bg-white rounded-lg shadow-md p-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <!-- Market Overview Card -->
        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <h3 class="text-lg font-semibold mb-2">Market Overview</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-gray-600">Total Volume</span>
              <span class="font-medium">$0.00</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">Active Traders</span>
              <span class="font-medium">0</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">Market Cap</span>
              <span class="font-medium">$0.00</span>
            </div>
          </div>
        </div>

        <!-- Price Analysis Card -->
        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <h3 class="text-lg font-semibold mb-2">Price Analysis</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-gray-600">Current Price</span>
              <span class="font-medium">$0.00</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">24h Change</span>
              <span class="font-medium text-green-600">+0.00%</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">7d Change</span>
              <span class="font-medium text-red-600">-0.00%</span>
            </div>
          </div>
        </div>

        <!-- Trading Activity Card -->
        <div class="bg-white rounded-lg border border-gray-200 p-4">
          <h3 class="text-lg font-semibold mb-2">Trading Activity</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span class="text-gray-600">24h Volume</span>
              <span class="font-medium">$0.00</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">Total Trades</span>
              <span class="font-medium">0</span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600">Avg. Trade Size</span>
              <span class="font-medium">$0.00</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Placeholder for Charts -->
      <div class="mt-8">
        <div class="bg-gray-50 rounded-lg p-4 text-center">
          <p class="text-gray-500">Charts and detailed analysis will be available soon</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MarketAnalysis'
}
</script> 