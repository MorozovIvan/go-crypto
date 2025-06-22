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
            @2fa-required="handle2FARequired"
            @logout="logout"
          />
        </div>
        <div v-else-if="currentView === 'market'">
          <MarketAnalysis />
        </div>
        <div v-else-if="currentView === 'solana-wallets'">
          <ProfitableSolanaWallets />
        </div>
        <div v-else-if="currentView === 'portfolio'">
          <Portfolio />
        </div>
        <div v-else-if="currentView === 'arbitration'">
          <ArbitrationDashboard />
        </div>
        <div v-else-if="currentView === 'sniping'">
          <SnipingDashboard />
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
import Portfolio from './components/Portfolio.vue'
import ArbitrationDashboard from './components/ArbitrationDashboard.vue'
import SnipingDashboard from './components/SnipingDashboard.vue'


const API_BASE_URL = 'http://localhost:8080'

export default {
  name: 'App',
  components: {
    AppHeader,
    AppSidebar,
    AppFooter,
    MainContent,
    MarketAnalysis,
    ProfitableSolanaWallets,
    Portfolio,
    ArbitrationDashboard,
    SnipingDashboard
  },
  data() {
    return {
      isConnected: false,
      userId: null,
      show2FAModal: false,
      twoFACode: '',
      currentView: 'telegram',
      walletAddress: null,
      currentPhone: '',
    }
  },
  async mounted() {
    // Check authentication status on app startup
    await this.checkAuthStatus()
  },
  methods: {
    async checkAuthStatus() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/telegram/status`)
        const data = await response.json()
        
        if (data.status && data.status.authenticated && data.status.user_id) {
          this.isConnected = true
          this.userId = data.status.user_id
          console.log('Restored authenticated session:', data.status)
        } else {
          this.isConnected = false
          this.userId = null
        }
      } catch (error) {
        console.error('Failed to check auth status:', error)
        this.isConnected = false
        this.userId = null
      }
    },
    async logout() {
      try {
        const response = await fetch(`${API_BASE_URL}/api/telegram/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        })
        
        const data = await response.json()
        if (data.success) {
          this.isConnected = false
          this.userId = null
          console.log('Logged out successfully')
          // Optionally refresh the page to reset all state
          window.location.reload()
        }
      } catch (error) {
        console.error('Logout failed:', error)
      }
    },
    handleConnect(userId) {
      this.isConnected = true
      this.userId = userId
    },
    async handle2FASubmit() {
      if (!this.twoFACode) {
        return;
      }

      try {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout

        const response = await fetch(`${API_BASE_URL}/api/telegram/verify-2fa`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            phone: this.currentPhone,
            password: this.twoFACode,
          }),
          signal: controller.signal
        });

        clearTimeout(timeoutId);
        const data = await response.json();

        if (data.success) {
          this.isConnected = true;
          this.userId = data.user_id;
          this.show2FAModal = false;
          this.twoFACode = '';
          console.log('2FA verification successful, refreshing page...');
          // Refresh the page to ensure clean state
          window.location.reload();
        } else {
          alert(data.message || '2FA verification failed');
        }
      } catch (error) {
        if (error.name === 'AbortError') {
          alert('Request timed out. Please try again.');
        } else {
          console.error('2FA verification error:', error);
          alert('Failed to verify 2FA code. Please try again.');
        }
      }
    },
    handle2FARequired(phone) {
      this.currentPhone = phone;
      this.show2FAModal = true;
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