<template>
  <header class="bg-[#0a78b9]">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center h-16">
        <div class="flex items-center gap-2">
          <!-- Solana SVG Icon -->
          <svg class="w-7 h-7" viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <linearGradient id="solana-gradient-header" x1="8" y1="8" x2="40" y2="40" gradientUnits="userSpaceOnUse">
              <stop stop-color="#00FFA3"/>
              <stop offset="1" stop-color="#DC1FFF"/>
            </linearGradient>
            <path d="M10.74 33.46c.18-.23.45-.36.74-.36h27.1c.55 0 .82.66.44 1.05l-4.1 4.09c-.18.18-.42.29-.67.29H7.25c-.55 0-.82-.66-.44-1.05l3.93-4.02zm0-18.92c.18-.18.42-.29.67-.29h26.9c.55 0 .82.66.44 1.05l-4.1 4.09c-.18.18-.42.29-.67.29H7.25c-.55 0-.82-.66-.44-1.05l3.93-4.09zm4.67 9.46c.18-.18.42-.29.67-.29h18.1c.55 0 .82.66.44 1.05l-4.1 4.09c-.18.18-.42.29-.67.29H12.25c-.55 0-.82-.66-.44-1.05l3.6-4.09z" fill="url(#solana-gradient-header)"/>
          </svg>
          <h1 class="text-xl font-bold text-white">Make me reach</h1>
        </div>
        <div class="flex items-center space-x-4 text-white">
          <SolanaWallet @connected="handleWalletConnected" @disconnected="handleWalletDisconnected" />
        </div>
      </div>
    </div>
  </header>
</template>

<script>
import SolanaWallet from './SolanaWallet.vue'

export default {
  name: 'AppHeader',
  components: {
    SolanaWallet
  },
  methods: {
    handleWalletConnected(publicKey) {
      this.$emit('wallet-connected', publicKey)
    },
    handleWalletDisconnected() {
      this.$emit('wallet-disconnected')
    }
  },
  emits: ['wallet-connected', 'wallet-disconnected']
}
</script> 