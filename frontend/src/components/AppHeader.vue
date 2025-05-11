<template>
  <header class="bg-[#0a78b9]">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center h-16">
        <div class="flex items-center">
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