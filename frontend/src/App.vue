<template>
  <div class="min-h-screen bg-gray-50">
    <AppHeader 
      @wallet-connected="handleWalletConnected"
      @wallet-disconnected="handleWalletDisconnected"
    />
    <div class="flex">
      <AppSidebar :current-view="currentView" @change-view="handleViewChange" />
      <main class="flex-1">
        <div v-if="currentView === 'telegram'">
          <MainContent
            :is-connected="isConnected"
            :user-id="userId"
            @connect="handleConnect"
            @2fa-required="show2FAModal = true"
          />
        </div>
        <div v-else-if="currentView === 'market'">
          <MarketAnalysis />
        </div>
        <div v-else-if="currentView === 'solana-wallets'">
          <ProfitableSolanaWallets />
        </div>
      </main>
    </div>
    <AppFooter />

    <!-- 2FA Modal -->
    <div v-if="show2FAModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-semibold mb-4">Two-Factor Authentication Required</h3>
        <p class="text-gray-600 mb-4">Please enter your 2FA code from the Telegram app.</p>
        <div class="space-y-4">
          <input
            type="text"
            v-model="twoFACode"
            placeholder="Enter 2FA code"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0088cc] focus:border-transparent"
          />
          <div class="flex justify-end space-x-2">
            <button
              @click="show2FAModal = false"
              class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-md transition-colors"
            >
              Cancel
            </button>
            <button
              @click="handle2FASubmit"
              class="px-4 py-2 bg-[#0088cc] text-white rounded-md hover:bg-[#0077b3] transition-colors"
            >
              Submit
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import AppHeader from './components/AppHeader.vue'
import AppSidebar from './components/AppSidebar.vue'
import AppFooter from './components/AppFooter.vue'
import MainContent from './components/MainContent.vue'
import MarketAnalysis from './components/MarketAnalysis.vue'
import ProfitableSolanaWallets from './components/ProfitableSolanaWallets.vue'

const API_BASE_URL = 'http://localhost:8080'

export default {
  name: 'App',
  components: {
    AppHeader,
    AppSidebar,
    AppFooter,
    MainContent,
    MarketAnalysis,
    ProfitableSolanaWallets
  },
  data() {
    return {
      isConnected: false,
      userId: null,
      show2FAModal: false,
      twoFACode: '',
      currentView: 'telegram',
      walletAddress: null
    }
  },
  methods: {
    handleConnect(userId) {
      this.isConnected = true
      this.userId = userId
    },
    handle2FASubmit() {
      // Handle 2FA submission
      this.show2FAModal = false
      this.twoFACode = ''
    },
    handleViewChange(view) {
      this.currentView = view
    },
    handleWalletConnected(publicKey) {
      this.walletAddress = publicKey
      // You can add additional logic here, such as fetching user data or initializing services
    },
    handleWalletDisconnected() {
      this.walletAddress = null
      // You can add cleanup logic here
    }
  }
}
</script>

<style>
/* Remove all existing styles as we're using Tailwind now */
</style> 