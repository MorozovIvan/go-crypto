<template>
  <header class="bg-gray-800 shadow-lg">
    <div class="px-4 mx-auto max-w-7xl sm:px-6 lg:px-8">
      <div class="flex items-center justify-between py-6">
        <!-- Logo and Title -->
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <div class="flex items-center justify-center w-10 h-10 rounded-lg bg-gradient-to-r from-blue-500 to-purple-600">
              <svg class="w-6 h-6 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M3 3a1 1 0 000 2v8a2 2 0 002 2h2.586l-1.293 1.293a1 1 0 101.414 1.414L10 15.414l2.293 2.293a1 1 0 001.414-1.414L12.414 15H15a2 2 0 002-2V5a1 1 0 100-2H3zm11.707 4.707a1 1 0 00-1.414-1.414L10 9.586 8.707 8.293a1 1 0 00-1.414 1.414L9.586 12l-2.293 2.293a1 1 0 101.414 1.414L10 13.414l2.293 2.293a1 1 0 001.414-1.414L11.414 12l2.293-2.293z" clip-rule="evenodd"/>
              </svg>
            </div>
          </div>
          <div class="ml-4">
            <h1 class="text-2xl font-bold text-white">Make Me Rich</h1>
            <p class="text-sm text-gray-300">Professional Trading Analytics</p>
          </div>
        </div>

        <!-- Navigation and Wallet -->
        <div class="flex items-center space-x-8">


          <!-- Connect Wallet Button -->
          <div class="relative">
            <button
              @click="isConnected ? toggleWalletMenu() : connectWallet()"
              :disabled="isConnecting"
              class="flex items-center px-6 py-2 space-x-2 font-medium text-white transition-all duration-200 transform rounded-lg shadow-lg bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 disabled:from-gray-600 disabled:to-gray-700 hover:shadow-xl hover:scale-105"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z"/>
              </svg>
              <span v-if="!isConnecting && !isConnected">Connect Wallet</span>
              <span v-else-if="isConnecting">Connecting...</span>
              <span v-else class="flex items-center space-x-3">
                <div class="flex items-center space-x-2">
                  <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                  <span>{{ formatAddress(walletAddress) }}</span>
                </div>
                <div class="pl-3 border-l border-gray-400">
                  <span class="text-sm font-bold text-green-400">{{ walletBalance }} SOL</span>
                </div>
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
                </svg>
              </span>
            </button>
            
            <!-- Wallet Menu Dropdown -->
            <div v-if="showWalletMenu && isConnected" class="absolute right-0 z-50 w-64 mt-2 bg-gray-800 border border-gray-700 rounded-lg shadow-xl wallet-menu">
              <div class="p-4">
                <div class="mb-2 text-sm text-gray-300">Wallet Address:</div>
                <div class="mb-3 font-mono text-xs text-white break-all">{{ walletAddress }}</div>
                <div class="mb-2 text-sm text-gray-300">Balance:</div>
                <div class="mb-2 font-bold text-green-400">{{ walletBalance }} SOL</div>
                <div class="mb-4 text-xs text-gray-400">
                  {{ balanceType === 'real' ? '✅ Real balance' : '🎭 Demo balance' }}
                </div>
                <button
                  @click="refreshBalance"
                  :disabled="isRefreshing"
                  class="w-full px-4 py-2 mb-2 text-sm text-white transition-colors duration-200 bg-blue-600 rounded-lg hover:bg-blue-700"
                >
                  {{ isRefreshing ? 'Refreshing...' : 'Refresh Balance' }}
                </button>
                <button
                  @click="disconnectWallet"
                  class="w-full px-4 py-2 text-sm text-white transition-colors duration-200 bg-red-600 rounded-lg hover:bg-red-700"
                >
                  Disconnect Wallet
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script>
export default {
  name: 'AppHeader',
  data() {
    return {
      isConnecting: false,
      isConnected: false,
      walletAddress: null,
      walletBalance: '0.00',
      balanceType: 'demo', // 'real' or 'demo'
      showWalletMenu: false,
      isRefreshing: false
    }
  },
  mounted() {
    // Add click outside listener to close wallet menu
    document.addEventListener('click', this.handleClickOutside);
    
    // Check if wallet is already connected on page load
    this.checkWalletConnection();
  },
  beforeUnmount() {
    // Remove click outside listener
    document.removeEventListener('click', this.handleClickOutside);
  },
  methods: {
    async connectWallet() {
      if (this.isConnecting || this.isConnected) return;
      
      this.isConnecting = true;
      
      try {
        // Check if Phantom wallet is available
        if (window.solana && window.solana.isPhantom) {
          console.log('Phantom wallet detected, connecting...');
          
          // Check if wallet is already connected
          if (window.solana.isConnected) {
            console.log('Wallet already connected');
            this.walletAddress = window.solana.publicKey.toString();
            await this.fetchWalletBalance(this.walletAddress);
            this.isConnected = true;
            this.$emit('wallet-connected', this.walletAddress);
            return;
          }
          
          // Request connection to Phantom with proper error handling
          const response = await window.solana.connect({ onlyIfTrusted: false });
          
          if (response && response.publicKey) {
            this.walletAddress = response.publicKey.toString();
            console.log('🎉 Successfully connected to Phantom wallet:', this.walletAddress);
            
            // Fetch wallet balance (with demo fallback due to RPC restrictions)
            await this.fetchWalletBalance(this.walletAddress);
            
            this.isConnected = true;
            this.$emit('wallet-connected', this.walletAddress);
          } else {
            throw new Error('Failed to get wallet address from connection response');
          }
          
        } else {
          console.log('Phantom wallet not detected, using demo mode');
          alert('Phantom wallet not detected. Please install Phantom wallet extension or use demo mode.');
          // Demo mode - simulate wallet connection
          this.walletAddress = 'Demo: 7xKXt...9d4Qc';
          this.walletBalance = '12.45';
          this.isConnected = true;
          this.$emit('wallet-connected', this.walletAddress);
        }
        
      } catch (error) {
        console.error('Wallet connection failed:', error);
        
        // Handle specific error cases
        if (error.code === 4001 || error.message?.includes('User rejected')) {
          console.log('User rejected wallet connection');
          alert('Wallet connection was rejected. Please try again and approve the connection.');
        } else if (error.code === 403 || error.message?.includes('Access forbidden')) {
          console.log('Access forbidden - wallet may be locked or permissions denied');
          alert('Access forbidden. Please unlock your Phantom wallet and try again.\n\nNote: Balance display may be limited due to RPC restrictions, but wallet connection will work.');
        } else if (error.code === -32603 || error.message?.includes('Internal error')) {
          console.log('Internal wallet error');
          alert('Wallet internal error. Please refresh the page and try again.');
        } else {
          console.log('Connection failed, offering demo mode');
          const useDemo = confirm('Wallet connection failed. Would you like to use demo mode instead?');
          if (useDemo) {
            // Fall back to demo mode
            this.walletAddress = 'Demo: 7xKXt...9d4Qc';
            this.walletBalance = '12.45';
            this.isConnected = true;
            this.$emit('wallet-connected', this.walletAddress);
          }
        }
      } finally {
        this.isConnecting = false;
      }
    },

    async fetchWalletBalance(walletAddress) {
      try {
        // Skip balance fetching for demo addresses
        if (walletAddress.startsWith('Demo:')) {
          this.walletBalance = '12.45';
          this.balanceType = 'demo';
          return;
        }

        console.log('Fetching balance for:', walletAddress);
        
        // Use backend proxy to avoid CORS issues
        if (window.solana && window.solana.isPhantom && window.solana.isConnected) {
          try {
            console.log('Attempting to fetch real balance via backend proxy...');
            
            // Create a timeout promise
            const timeoutPromise = new Promise((_, reject) => 
              setTimeout(() => reject(new Error('Request timeout')), 5000)
            );
            
            // Fetch balance via backend proxy
            const fetchPromise = fetch(`http://localhost:8080/api/solana/balance/${walletAddress}`, {
              method: 'GET',
              headers: {
                'Content-Type': 'application/json',
              }
            });
            
            const response = await Promise.race([fetchPromise, timeoutPromise]);
            
            if (response.ok) {
              const data = await response.json();
              if (data.balance !== undefined && data.balance !== null) {
                this.walletBalance = parseFloat(data.balance).toFixed(4);
                this.balanceType = 'real';
                console.log('✅ Real balance fetched successfully:', this.walletBalance, 'SOL');
                console.log('🔗 RPC endpoint used:', data.endpoint);
                return;
              }
            } else {
              const errorData = await response.json();
              throw new Error(errorData.error || 'Backend proxy failed');
            }
            
          } catch (error) {
            console.log('Backend proxy balance fetch failed:', error.message);
            console.log('Falling back to demo balance...');
          }
        }
        
        // Fallback to demo balance
        console.log('📝 Using demo balance');
        const baseBalance = 12.45;
        const variation = (Math.random() - 0.5) * 0.4;
        this.walletBalance = Math.max(0.1, baseBalance + variation).toFixed(4);
        this.balanceType = 'demo';
        
        console.log('✅ Wallet connected successfully!');
        console.log('🔗 Connected wallet:', walletAddress);
        console.log('💰 Demo balance:', this.walletBalance, 'SOL');
        
      } catch (error) {
        console.error('Error in wallet balance handling:', error);
        this.walletBalance = '12.45';
        this.balanceType = 'demo';
      }
    },


    
    formatAddress(address) {
      if (!address) return '';
      if (address.length < 8) return address;
      return `${address.slice(0, 4)}...${address.slice(-4)}`;
    },
    
    toggleWalletMenu() {
      this.showWalletMenu = !this.showWalletMenu;
    },
    
    refreshBalance() {
      this.isRefreshing = true;
      this.fetchWalletBalance(this.walletAddress).finally(() => {
        this.isRefreshing = false;
      });
    },
    
    disconnectWallet() {
      this.isConnected = false;
      this.walletAddress = null;
      this.walletBalance = '0.00';
      this.balanceType = 'demo';
      this.showWalletMenu = false;
    },

    handleClickOutside(event) {
      if (this.showWalletMenu && !event.target.closest('.wallet-menu')) {
        this.showWalletMenu = false;
      }
    },

    async checkWalletConnection() {
      if (window.solana && window.solana.isPhantom) {
        try {
          if (window.solana.isConnected) {
            console.log('Wallet already connected');
            this.walletAddress = window.solana.publicKey.toString();
            await this.fetchWalletBalance(this.walletAddress);
            this.isConnected = true;
            this.$emit('wallet-connected', this.walletAddress);
          }
        } catch (error) {
          console.error('Error checking wallet connection:', error);
        }
      }
    }
  }
}
</script>

<style scoped>
/* Simple, clean styles */
</style> 