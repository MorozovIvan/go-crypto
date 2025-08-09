<template>
	<!-- Compact Wallet Button for Navbar -->
	<div class="dropdown dropdown-end" v-if="!isConnected">
		<div tabindex="0" role="button" class="btn btn-primary btn-sm">
			<svg
				class="w-4 h-4"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z"
				/>
			</svg>
			Connect Wallet
		</div>
		<ul
			tabindex="0"
			class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-200 rounded-box w-64 border border-base-300"
		>
			<li class="menu-title">
				<span>Choose Wallet</span>
			</li>
			<li>
				<a @click="connectPhantom" :class="{ disabled: !isPhantomAvailable }">
					<div class="avatar">
						<div
							class="w-8 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500"
						>
							<div
								class="flex items-center justify-center w-full h-full text-white text-xs font-bold"
							>
								P
							</div>
						</div>
					</div>
					<div>
						<div class="font-medium">Phantom</div>
						<div class="text-xs opacity-70">
							{{ isPhantomAvailable ? 'Available' : 'Not installed' }}
						</div>
					</div>
				</a>
			</li>
			<li>
				<a @click="connectSolflare" :class="{ disabled: !isSolflareAvailable }">
					<div class="avatar">
						<div
							class="w-8 rounded-lg bg-gradient-to-br from-orange-500 to-red-500"
						>
							<div
								class="flex items-center justify-center w-full h-full text-white text-xs font-bold"
							>
								S
							</div>
						</div>
					</div>
					<div>
						<div class="font-medium">Solflare</div>
						<div class="text-xs opacity-70">
							{{ isSolflareAvailable ? 'Available' : 'Not installed' }}
						</div>
					</div>
				</a>
			</li>
			<div class="divider my-1"></div>
			<li class="p-2">
				<div class="text-xs opacity-70 text-center">
					Connect your wallet to access advanced trading features
				</div>
			</li>
		</ul>
	</div>

	<!-- Connected Wallet Display -->
	<div v-else class="dropdown dropdown-end">
		<div tabindex="0" role="button" class="btn btn-success btn-sm">
			<div class="w-2 h-2 bg-success rounded-full animate-pulse mr-2"></div>
			<span class="hidden sm:inline">{{ formatAddress(publicKey) }}</span>
			<span class="sm:hidden">Wallet</span>
			<div class="badge badge-outline badge-xs ml-2">{{ balance }} SOL</div>
		</div>
		<ul
			tabindex="0"
			class="dropdown-content z-[1] menu p-2 shadow-lg bg-base-200 rounded-box w-80 border border-base-300"
		>
			<li class="menu-title">
				<span>Connected Wallet</span>
			</li>
			<li class="p-4">
				<div class="space-y-3">
					<div class="flex items-center gap-3">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-success/20 to-primary/20 border-2 border-success/30"
							>
								<div
									class="flex items-center justify-center w-full h-full text-success text-lg font-bold"
								>
									✓
								</div>
							</div>
						</div>
						<div>
							<div class="font-semibold text-success">Connected</div>
							<div class="text-xs opacity-70">{{ walletType }} Wallet</div>
						</div>
					</div>

					<div class="space-y-2">
						<div class="flex justify-between items-center">
							<span class="text-sm opacity-70">Address:</span>
							<div class="flex items-center gap-2">
								<code class="text-xs bg-base-300 px-2 py-1 rounded font-mono">
									{{ formatAddress(publicKey) }}
								</code>
								<button @click="copyAddress" class="btn btn-ghost btn-xs">
									<svg
										class="w-3 h-3"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
										></path>
									</svg>
								</button>
							</div>
						</div>

						<div class="flex justify-between items-center">
							<span class="text-sm opacity-70">Balance:</span>
							<div class="flex items-center gap-2">
								<span class="font-mono font-bold text-success"
									>{{ balance }} SOL</span
								>
								<button
									@click="refreshBalance"
									:disabled="refreshing"
									class="btn btn-ghost btn-xs"
								>
									<svg
										class="w-3 h-3"
										:class="{ 'animate-spin': refreshing }"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
										></path>
									</svg>
								</button>
							</div>
						</div>

						<div class="flex justify-between items-center">
							<span class="text-sm opacity-70">Network:</span>
							<div class="badge badge-primary badge-sm">Mainnet</div>
						</div>
					</div>

					<div class="flex gap-2 pt-2">
						<button
							@click="viewOnExplorer"
							class="btn btn-outline btn-sm flex-1"
						>
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
								></path>
							</svg>
							Explorer
						</button>
						<button @click="disconnect" class="btn btn-error btn-sm flex-1">
							<svg
								class="w-4 h-4"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"
								></path>
							</svg>
							Disconnect
						</button>
					</div>
				</div>
			</li>
		</ul>
	</div>
</template>

<script>
import { Connection, PublicKey, LAMPORTS_PER_SOL } from '@solana/web3.js'

export default {
	name: 'WalletConnection',
	emits: ['wallet-connected', 'wallet-disconnected'],
	data() {
		return {
			isConnected: false,
			publicKey: null,
			balance: 0,
			connection: null,
			walletType: 'Unknown',
			refreshing: false,
			rpcEndpoints: [],
		}
	},
	computed: {
		isPhantomAvailable() {
			return (
				typeof window !== 'undefined' &&
				window.solana &&
				window.solana.isPhantom
			)
		},
		isSolflareAvailable() {
			return (
				typeof window !== 'undefined' &&
				window.solflare &&
				window.solflare.isSolflare
			)
		},
	},
	async mounted() {
		// Use multiple RPC endpoints with fallback
		await this.initializeConnection()
		this.checkWalletConnection()
	},
	methods: {
		async initializeConnection() {
			// Use working RPC endpoints with your Helius API key
			this.rpcEndpoints = [
				'https://mainnet.helius-rpc.com/?api-key=40189f56-ecbc-45e6-a123-4a39830d27d8',
				'https://rpc.ankr.com/solana',
				'https://api.mainnet-beta.solana.com',
			]

			// Use the first endpoint by default
			this.connection = new Connection(this.rpcEndpoints[0], 'confirmed')
		},

		async checkWalletConnection() {
			if (this.isPhantomAvailable && window.solana.isConnected) {
				try {
					const publicKey = window.solana.publicKey.toString()
					await this.handleWalletConnected(publicKey)
				} catch (error) {
					console.error('Error checking wallet connection:', error)
				}
			}
		},
		async connectPhantom() {
			if (!this.isPhantomAvailable) {
				window.open('https://phantom.app/', '_blank')
				return
			}

			try {
				const response = await window.solana.connect()
				const publicKey = response.publicKey.toString()
				await this.handleWalletConnected(publicKey, 'Phantom')
			} catch (error) {
				console.error('Error connecting to Phantom:', error)
				this.showError('Failed to connect to Phantom wallet')
			}
		},
		async connectSolflare() {
			if (!this.isSolflareAvailable) {
				window.open('https://solflare.com/', '_blank')
				return
			}

			try {
				const response = await window.solflare.connect()
				const publicKey = response.publicKey.toString()
				await this.handleWalletConnected(publicKey, 'Solflare')
			} catch (error) {
				console.error('Error connecting to Solflare:', error)
				this.showError('Failed to connect to Solflare wallet')
			}
		},
		async handleWalletConnected(publicKeyString, type = 'Unknown') {
			this.isConnected = true
			this.publicKey = publicKeyString
			this.walletType = type

			// Get wallet balance with fallback RPC endpoints
			await this.fetchBalanceWithFallback(publicKeyString)
			this.$emit('wallet-connected', publicKeyString)
		},

		async fetchBalanceWithFallback(publicKeyString) {
			const endpoints =
				this.rpcEndpoints.length > 0
					? this.rpcEndpoints
					: [
							'https://rpc.ankr.com/solana',
							'https://api.mainnet-beta.solana.com',
					  ]

			for (const endpoint of endpoints) {
				try {
					const connection = new Connection(endpoint, 'confirmed')
					const publicKey = new PublicKey(publicKeyString)
					const balance = await connection.getBalance(publicKey)
					this.balance = (balance / LAMPORTS_PER_SOL).toFixed(4)
					this.connection = connection // Update to working connection
					return
				} catch (error) {
					console.warn(`RPC endpoint ${endpoint} failed:`, error.message)
					continue
				}
			}

			// If all endpoints fail, set balance to N/A
			console.warn('All RPC endpoints failed, unable to fetch balance')
			this.balance = 'N/A'
		},
		async disconnect() {
			try {
				if (this.isPhantomAvailable && window.solana.isConnected) {
					await window.solana.disconnect()
				}
				if (this.isSolflareAvailable && window.solflare.isConnected) {
					await window.solflare.disconnect()
				}
			} catch (error) {
				console.error('Error disconnecting wallet:', error)
			}

			this.isConnected = false
			this.publicKey = null
			this.balance = 0
			this.$emit('wallet-disconnected')
		},
		formatAddress(address) {
			if (!address) return ''
			return `${address.slice(0, 4)}...${address.slice(-4)}`
		},

		async copyAddress() {
			if (this.publicKey) {
				try {
					await navigator.clipboard.writeText(this.publicKey)
					// You can add a toast notification here
					console.log('Address copied to clipboard')
				} catch (error) {
					console.error('Failed to copy address:', error)
				}
			}
		},

		async refreshBalance() {
			if (!this.publicKey) return

			this.refreshing = true
			try {
				await this.fetchBalanceWithFallback(this.publicKey)
			} catch (error) {
				console.warn('Unable to refresh balance:', error.message)
				this.showError('Unable to fetch balance. Please try again.')
			} finally {
				this.refreshing = false
			}
		},

		viewOnExplorer() {
			if (this.publicKey) {
				window.open(`https://solscan.io/account/${this.publicKey}`, '_blank')
			}
		},

		showError(message) {
			// You can implement a toast notification system here
			alert(message)
		},
	},
}
</script>
