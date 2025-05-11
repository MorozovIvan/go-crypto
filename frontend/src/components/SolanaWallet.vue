<template>
  <div class="flex items-center space-x-4">
    <div v-if="connected" class="flex items-center space-x-2">
      <span class="text-sm text-gray-600">Balance:</span>
      <span class="font-medium">{{ formatBalance(balance) }} SOL</span>
      <button
        @click="disconnect"
        class="px-3 py-1 text-sm text-red-600 hover:text-red-700 transition-colors"
      >
        Disconnect
      </button>
    </div>
    <button
      v-else
      @click="connect"
      class="px-4 py-2 bg-[#0088cc] text-white rounded-lg hover:bg-[#0077b3] transition-colors flex items-center space-x-2"
      :disabled="connecting"
    >
      <svg v-if="connecting" class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
      <span>{{ connecting ? 'Connecting...' : 'Connect Wallet' }}</span>
    </button>
  </div>
</template>

<script>
import { Connection, PublicKey, LAMPORTS_PER_SOL } from '@solana/web3.js'

export default {
  name: 'SolanaWallet',
  data() {
    return {
      connected: false,
      connecting: false,
      balance: 0,
      publicKey: null,
      connection: null
    }
  },
  methods: {
    async connect() {
      if (this.connected) return

      this.connecting = true
      try {
        // Check if Phantom is installed
        if (!window.solana || !window.solana.isPhantom) {
          window.open('https://phantom.app/', '_blank')
          throw new Error('Phantom wallet is not installed')
        }

        // Connect to Phantom
        const resp = await window.solana.connect()
        this.publicKey = new PublicKey(resp.publicKey.toString())
        
        // Get API key from Vite environment variable
        const apiKey = import.meta.env.VITE_SOLANA_RPC_KEY
        if (!apiKey) {
          throw new Error('Solana RPC API key is not set in .env')
        }
        // Initialize connection to Solana network with your Extrnode API key
        this.connection = new Connection(`https://solana-mainnet.rpc.extrnode.com/${apiKey}`, 'confirmed')
        
        // Get balance
        const balance = await this.connection.getBalance(this.publicKey)
        this.balance = balance / LAMPORTS_PER_SOL

        // Set up balance update interval
        this.balanceInterval = setInterval(this.updateBalance, 10000) // Update every 10 seconds

        this.connected = true
        this.$emit('connected', this.publicKey.toString())
      } catch (err) {
        console.error('Failed to connect wallet:', err)
        let errorMessage = 'Failed to connect wallet'
        if (err.message) {
          errorMessage = err.message
        } else if (err.toString) {
          errorMessage = err.toString()
        }
        alert(errorMessage)
      } finally {
        this.connecting = false
      }
    },
    async disconnect() {
      if (!this.connected) return;

      try {
        // Listen for Phantom's disconnect event to ensure state is cleared
        if (window.solana && window.solana.isPhantom) {
          window.solana.on('disconnect', () => {
            this.connected = false;
            this.publicKey = null;
            this.balance = 0;
            if (this.balanceInterval) {
              clearInterval(this.balanceInterval);
              this.balanceInterval = null;
            }
            this.$emit('disconnected');
          });
          await window.solana.disconnect();
        } else {
          // Fallback: just clear state
          this.connected = false;
          this.publicKey = null;
          this.balance = 0;
          if (this.balanceInterval) {
            clearInterval(this.balanceInterval);
            this.balanceInterval = null;
          }
          this.$emit('disconnected');
        }
      } catch (err) {
        console.error('Failed to disconnect wallet:', err);
        alert('Failed to disconnect wallet');
        // Always clear state even if error
        this.connected = false;
        this.publicKey = null;
        this.balance = 0;
        if (this.balanceInterval) {
          clearInterval(this.balanceInterval);
          this.balanceInterval = null;
        }
        this.$emit('disconnected');
      }
    },
    async updateBalance() {
      if (!this.connected || !this.publicKey || !this.connection) return

      try {
        const balance = await this.connection.getBalance(this.publicKey)
        this.balance = balance / LAMPORTS_PER_SOL
      } catch (err) {
        console.error('Failed to update balance:', err)
      }
    },
    formatBalance(balance) {
      return balance.toFixed(4)
    },
    async handleAutoConnect() {
      try {
        this.publicKey = new PublicKey(window.solana.publicKey.toString());
        const apiKey = import.meta.env.VITE_SOLANA_RPC_KEY;
        this.connection = new Connection(`https://solana-mainnet.rpc.extrnode.com/${apiKey}`, 'confirmed');
        const balance = await this.connection.getBalance(this.publicKey);
        this.balance = balance / LAMPORTS_PER_SOL;
        this.connected = true;
        this.balanceInterval = setInterval(this.updateBalance, 10000);
        this.$emit('connected', this.publicKey.toString());
      } catch (err) {
        console.error('Failed to auto-connect wallet:', err);
      }
    }
  },
  beforeUnmount() {
    if (this.balanceInterval) {
      clearInterval(this.balanceInterval)
    }
  },
  mounted() {
    // Only treat as connected if Phantom is connected in this session
    if (
      window.solana &&
      window.solana.isPhantom &&
      window.solana.isConnected &&
      window.solana.publicKey
    ) {
      this.handleAutoConnect();
    }
  }
}
</script> 