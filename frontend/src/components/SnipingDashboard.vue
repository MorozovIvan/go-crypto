<template>
	<div class="max-w-7xl mx-auto space-y-6">
		<!-- Header with Monitoring Control -->
		<div class="bg-base-100 rounded-xl shadow-sm border border-base-300 p-6">
			<div class="flex items-center justify-between">
				<div>
					<h1 class="text-2xl font-bold text-base-content">
						Solana Token Sniping
					</h1>
					<p class="text-base-content/70 mt-1">
						Real-time detection of new tokens on Pump.fun and Raydium
					</p>
				</div>

				<div class="flex items-center gap-4">
					<!-- Status Indicator -->
					<div class="flex items-center gap-2">
						<div
							:class="[
								'w-3 h-3 rounded-full',
								monitorStatus.isRunning
									? 'bg-success animate-pulse'
									: 'bg-error',
							]"
						></div>
						<span class="text-sm font-medium">
							{{
								monitorStatus.isRunning
									? 'Monitoring Active'
									: 'Monitoring Stopped'
							}}
						</span>
					</div>

					<!-- Start/Stop Button -->
					<button
						@click="toggleMonitoring"
						:disabled="isToggling"
						:class="[
							'btn',
							monitorStatus.isRunning ? 'btn-error' : 'btn-success',
						]"
					>
						<span
							v-if="isToggling"
							class="loading loading-spinner loading-sm"
						></span>
						{{
							isToggling
								? 'Processing...'
								: monitorStatus.isRunning
								? 'Stop Monitoring'
								: 'Start Monitoring'
						}}
					</button>
				</div>
			</div>
		</div>

		<!-- Stats Overview -->
		<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
			<div class="bg-base-100 rounded-xl shadow-sm border border-base-300 p-6">
				<div class="flex items-center gap-3">
					<div
						class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center"
					>
						<svg
							class="w-6 h-6 text-primary"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
							></path>
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-base-content">
							{{ monitorStatus.tokensFound || 0 }}
						</p>
						<p class="text-sm text-base-content/70">Tokens Found</p>
					</div>
				</div>
			</div>

			<div class="bg-base-100 rounded-xl shadow-sm border border-base-300 p-6">
				<div class="flex items-center gap-3">
					<div
						class="w-12 h-12 bg-secondary/10 rounded-lg flex items-center justify-center"
					>
						<svg
							class="w-6 h-6 text-secondary"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
							></path>
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-base-content">
							{{ formatUptime() }}
						</p>
						<p class="text-sm text-base-content/70">Uptime</p>
					</div>
				</div>
			</div>

			<div class="bg-base-100 rounded-xl shadow-sm border border-base-300 p-6">
				<div class="flex items-center gap-3">
					<div
						class="w-12 h-12 bg-accent/10 rounded-lg flex items-center justify-center"
					>
						<svg
							class="w-6 h-6 text-accent"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 10V3L4 14h7v7l9-11h-7z"
							></path>
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-base-content">
							{{ formatLastToken() }}
						</p>
						<p class="text-sm text-base-content/70">Last Token</p>
					</div>
				</div>
			</div>

			<div class="bg-base-100 rounded-xl shadow-sm border border-base-300 p-6">
				<div class="flex items-center gap-3">
					<div
						class="w-12 h-12 bg-info/10 rounded-lg flex items-center justify-center"
					>
						<svg
							class="w-6 h-6 text-info"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"
							></path>
						</svg>
					</div>
					<div>
						<p class="text-2xl font-bold text-base-content">
							<span :class="wsConnected ? 'text-success' : 'text-error'">
								{{ wsConnected ? 'Connected' : 'Disconnected' }}
							</span>
						</p>
						<p class="text-sm text-base-content/70">WebSocket</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Tokens List -->
		<div class="bg-base-100 rounded-xl shadow-sm border border-base-300">
			<div class="p-6 border-b border-base-300">
				<div class="flex items-center justify-between">
					<h2 class="text-xl font-semibold text-base-content">
						Detected Tokens
					</h2>
					<div class="flex items-center gap-3">
						<!-- Platform Filter -->
						<div class="form-control">
							<select
								v-model="selectedPlatform"
								@change="onPlatformChange"
								class="select select-sm select-bordered"
							>
								<option value="">All Platforms</option>
								<option
									v-for="platform in availablePlatforms"
									:key="platform"
									:value="platform"
								>
									{{ platform }}
								</option>
							</select>
						</div>
						<button
							@click="refreshTokens"
							:disabled="isLoading"
							class="btn btn-sm btn-ghost"
						>
							<svg
								:class="['w-4 h-4', isLoading && 'animate-spin']"
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
							Refresh
						</button>
					</div>
				</div>
			</div>

			<!-- Loading State -->
			<div v-if="isLoading && tokens.length === 0" class="p-12 text-center">
				<span class="loading loading-spinner loading-lg text-primary"></span>
				<p class="mt-4 text-base-content/70">Loading tokens...</p>
			</div>

			<!-- Empty State -->
			<div
				v-else-if="tokens.length === 0"
				class="p-12 text-center text-base-content/70"
			>
				<div
					class="w-16 h-16 bg-base-200 rounded-full flex items-center justify-center mx-auto mb-4"
				>
					<svg
						class="w-8 h-8"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"
						></path>
					</svg>
				</div>
				<h3 class="text-lg font-medium text-base-content mb-2">
					No tokens detected yet
				</h3>
				<p class="text-sm">
					{{
						monitorStatus.isRunning
							? 'Monitoring Solana blockchain for new token launches...'
							: 'Start monitoring to detect new tokens on Pump.fun and Raydium'
					}}
				</p>
			</div>

			<!-- Tokens Grid -->
			<div v-else class="p-6">
				<div class="grid gap-4">
					<div
						v-for="token in filteredTokens"
						:key="token.id"
						:class="[
							'bg-base-200 rounded-lg p-4 border border-base-300 hover:shadow-md transition-all',
							isNewToken(token) && 'ring-2 ring-primary animate-pulse',
						]"
					>
						<div class="flex items-start justify-between">
							<div class="flex-1">
								<!-- Token Header -->
								<div class="flex items-center gap-3 mb-3">
									<div
										class="w-10 h-10 bg-gradient-to-br from-primary to-secondary rounded-lg flex items-center justify-center text-white font-bold text-xs"
									>
										{{ getTokenIndex(token) }}
									</div>
									<div>
										<h3 class="font-semibold text-base-content">
											{{ getTokenDisplayName(token) }}
										</h3>
										<p class="text-sm text-base-content/70">
											{{ formatTime(token.timestamp) }}
										</p>
									</div>
									<div class="ml-auto flex items-center gap-2">
										<span :class="['badge', getPlatformColor(token.platform)]">
											{{ token.platform || 'Unknown' }}
										</span>
										<span :class="['badge', getAgeColor(token.timestamp)]">
											{{ getAgeStatus(token.timestamp) }}
										</span>
									</div>
								</div>

								<!-- Token Details -->
								<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
									<div>
										<label
											class="text-xs font-medium text-base-content/70 uppercase tracking-wide"
											>Token Address</label
										>
										<div
											class="font-mono text-sm bg-base-100 p-2 rounded border mt-1"
										>
											{{ formatAddress(token.baseInfo.baseAddress) }}
										</div>
									</div>
									<div>
										<label
											class="text-xs font-medium text-base-content/70 uppercase tracking-wide"
											>Creator</label
										>
										<div
											class="font-mono text-sm bg-base-100 p-2 rounded border mt-1"
										>
											{{ formatAddress(token.creator) }}
										</div>
									</div>
								</div>

								<!-- Transaction Info -->
								<div class="mb-4">
									<label
										class="text-xs font-medium text-base-content/70 uppercase tracking-wide"
										>Transaction Signature</label
									>
									<div
										class="font-mono text-sm bg-base-100 p-2 rounded border mt-1"
									>
										{{ formatAddress(token.lpSignature) }}
									</div>
								</div>
							</div>
						</div>

						<!-- Actions -->
						<div class="flex items-center gap-2 pt-4 border-t border-base-300">
							<button @click="viewToken(token)" class="btn btn-sm btn-primary">
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
										d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
									></path>
									<path
										stroke-linecap="round"
										stroke-linejoin="round"
										stroke-width="2"
										d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
									></path>
								</svg>
								View Details
							</button>
							<button
								@click="copyAddress(token.baseInfo.baseAddress)"
								class="btn btn-sm btn-ghost"
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
										d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
									></path>
								</svg>
								Copy Address
							</button>
							<a
								:href="`https://solscan.io/tx/${token.lpSignature}`"
								target="_blank"
								class="btn btn-sm btn-ghost"
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
								Solscan
							</a>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Token Details Modal -->
		<dialog :class="{ 'modal modal-open': selectedToken }" class="modal">
			<div class="modal-box max-w-4xl">
				<div v-if="selectedToken" class="space-y-6">
					<div class="flex items-center justify-between">
						<h3 class="font-bold text-xl">Token Details</h3>
						<button
							@click="selectedToken = null"
							class="btn btn-sm btn-circle btn-ghost"
						>
							✕
						</button>
					</div>

					<!-- Token Overview -->
					<div class="bg-base-200 rounded-lg p-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="text-sm font-medium text-base-content/70"
									>Token Address</label
								>
								<div
									class="font-mono text-sm bg-base-100 p-3 rounded border mt-1 break-all"
								>
									{{ selectedToken.baseInfo.baseAddress }}
								</div>
							</div>
							<div>
								<label class="text-sm font-medium text-base-content/70"
									>Creator</label
								>
								<div
									class="font-mono text-sm bg-base-100 p-3 rounded border mt-1 break-all"
								>
									{{ selectedToken.creator }}
								</div>
							</div>
							<div>
								<label class="text-sm font-medium text-base-content/70"
									>Platform</label
								>
								<div class="text-sm bg-base-100 p-3 rounded border mt-1">
									<span
										:class="['badge', getPlatformColor(selectedToken.platform)]"
									>
										{{ selectedToken.platform || 'Unknown' }}
									</span>
								</div>
							</div>
							<div>
								<label class="text-sm font-medium text-base-content/70"
									>Transaction Signature</label
								>
								<div
									class="font-mono text-sm bg-base-100 p-3 rounded border mt-1 break-all"
								>
									{{ selectedToken.lpSignature }}
								</div>
							</div>
							<div>
								<label class="text-sm font-medium text-base-content/70"
									>Created At</label
								>
								<div class="text-sm bg-base-100 p-3 rounded border mt-1">
									{{ formatFullTime(selectedToken.timestamp) }}
								</div>
							</div>
						</div>
					</div>

					<!-- Transaction Logs -->
					<div>
						<label class="text-sm font-medium text-base-content/70 mb-2 block"
							>Transaction Logs</label
						>
						<div class="bg-base-200 p-4 rounded-lg max-h-60 overflow-y-auto">
							<div
								v-for="(log, index) in selectedToken.logs"
								:key="index"
								class="font-mono text-xs mb-1 text-base-content/80"
							>
								{{ log }}
							</div>
						</div>
					</div>

					<div class="modal-action">
						<button
							@click="copyTokenData(selectedToken)"
							class="btn btn-primary"
						>
							Copy All Data
						</button>
						<button @click="selectedToken = null" class="btn">Close</button>
					</div>
				</div>
			</div>
			<form method="dialog" class="modal-backdrop">
				<button @click="selectedToken = null">close</button>
			</form>
		</dialog>
	</div>
</template>

<script>
const API_BASE_URL = 'http://localhost:8080'

export default {
	name: 'SnipingDashboard',
	data() {
		return {
			tokens: [],
			selectedToken: null,
			isLoading: false,
			isToggling: false,
			wsConnected: false,
			ws: null,
			selectedPlatform: '',
			availablePlatforms: [],
			monitorStatus: {
				isRunning: false,
				startTime: null,
				tokensFound: 0,
				lastTokenTime: null,
			},
			newTokenIds: new Set(),
		}
	},
	computed: {
		filteredTokens() {
			// For now, return all tokens. Can add filtering logic here later
			return this.tokens
		},
	},
	async mounted() {
		await this.loadPlatforms()
		await this.loadTokens()
		await this.loadMonitorStatus()
		this.connectWebSocket()
	},
	beforeUnmount() {
		this.disconnectWebSocket()
	},
	methods: {
		async loadPlatforms() {
			try {
				const response = await fetch(`${API_BASE_URL}/api/platforms`)
				const data = await response.json()
				if (data.success) {
					this.availablePlatforms = data.platforms || []
				}
			} catch (error) {
				console.error('Failed to load platforms:', error)
			}
		},

		async loadTokens() {
			this.isLoading = true
			try {
				let url = `${API_BASE_URL}/api/new-tokens`
				if (this.selectedPlatform) {
					url += `?platform=${encodeURIComponent(this.selectedPlatform)}`
				}
				const response = await fetch(url)
				const data = await response.json()
				if (data.success) {
					this.tokens = data.data || []
				}
			} catch (error) {
				console.error('Failed to load tokens:', error)
			} finally {
				this.isLoading = false
			}
		},

		onPlatformChange() {
			this.loadTokens()
		},

		async loadMonitorStatus() {
			try {
				const response = await fetch(`${API_BASE_URL}/api/monitor/status`)
				const data = await response.json()
				if (data.success) {
					this.monitorStatus = data.status
				}
			} catch (error) {
				console.error('Failed to load monitor status:', error)
			}
		},

		async toggleMonitoring() {
			this.isToggling = true
			try {
				const endpoint = this.monitorStatus.isRunning
					? '/api/monitor/stop'
					: '/api/monitor/start'
				const response = await fetch(`${API_BASE_URL}${endpoint}`, {
					method: 'POST',
				})
				const data = await response.json()

				if (data.success) {
					this.monitorStatus = data.status
				}
			} catch (error) {
				console.error('Failed to toggle monitoring:', error)
			} finally {
				this.isToggling = false
			}
		},

		connectWebSocket() {
			const wsUrl = `ws://localhost:8080/api/new-tokens/stream`
			this.ws = new WebSocket(wsUrl)

			this.ws.onopen = () => {
				this.wsConnected = true
				console.log('WebSocket connected')
			}

			this.ws.onclose = () => {
				this.wsConnected = false
				console.log('WebSocket disconnected')
				setTimeout(() => {
					if (!this.wsConnected) {
						this.connectWebSocket()
					}
				}, 5000)
			}

			this.ws.onerror = error => {
				console.error('WebSocket error:', error)
			}

			this.ws.onmessage = event => {
				try {
					const message = JSON.parse(event.data)
					this.handleWebSocketMessage(message)
				} catch (error) {
					console.error('Failed to parse WebSocket message:', error)
				}
			}
		},

		disconnectWebSocket() {
			if (this.ws) {
				this.ws.close()
				this.ws = null
			}
		},

		handleWebSocketMessage(message) {
			switch (message.type) {
				case 'new_token':
					this.addNewToken(message.data)
					break
				case 'status_update':
					this.monitorStatus = message.data
					break
			}
		},

		addNewToken(token) {
			this.tokens.unshift(token)
			this.newTokenIds.add(token.id)
			setTimeout(() => {
				this.newTokenIds.delete(token.id)
			}, 5000)

			// Keep only last 50 tokens
			if (this.tokens.length > 50) {
				this.tokens = this.tokens.slice(0, 50)
			}
		},

		async refreshTokens() {
			await this.loadTokens()
			await this.loadMonitorStatus()
		},

		viewToken(token) {
			this.selectedToken = token
		},

		copyAddress(address) {
			navigator.clipboard.writeText(address)
		},

		copyTokenData(token) {
			const data = JSON.stringify(token, null, 2)
			navigator.clipboard.writeText(data)
		},

		formatAddress(address) {
			if (!address) return 'N/A'
			return `${address.slice(0, 8)}...${address.slice(-8)}`
		},

		formatTime(timestamp) {
			return new Date(timestamp).toLocaleTimeString()
		},

		formatFullTime(timestamp) {
			return new Date(timestamp).toLocaleString()
		},

		formatUptime() {
			if (
				!this.monitorStatus.startTime ||
				this.monitorStatus.startTime === '0001-01-01T00:00:00Z' ||
				!this.monitorStatus.isRunning
			) {
				return '0m'
			}

			try {
				const start = new Date(this.monitorStatus.startTime)
				const now = new Date()

				// Validate dates and check for Go zero time
				if (
					isNaN(start.getTime()) ||
					isNaN(now.getTime()) ||
					start.getFullYear() < 1900
				) {
					return '0m'
				}

				const diff = Math.max(0, now - start) // Ensure positive difference
				const totalMinutes = Math.floor(diff / 60000)

				// Sanity check: if uptime is more than 1 year, something is wrong
				if (totalMinutes > 525600) {
					// 365 days * 24 hours * 60 minutes
					return '0m'
				}

				const hours = Math.floor(totalMinutes / 60)
				const minutes = totalMinutes % 60

				if (hours > 0) {
					return `${hours}h ${minutes}m`
				}
				return `${totalMinutes}m`
			} catch (error) {
				console.error('Error formatting uptime:', error)
				return '0m'
			}
		},

		formatLastToken() {
			if (
				!this.monitorStatus.lastTokenTime ||
				this.monitorStatus.lastTokenTime === '0001-01-01T00:00:00Z'
			) {
				return 'Never'
			}

			try {
				const last = new Date(this.monitorStatus.lastTokenTime)
				const now = new Date()

				// Validate dates
				if (isNaN(last.getTime()) || isNaN(now.getTime())) {
					return 'Never'
				}

				const diff = Math.max(0, now - last) // Ensure positive difference
				const totalMinutes = Math.floor(diff / 60000)

				if (totalMinutes < 1) return 'Just now'
				if (totalMinutes < 60) return `${totalMinutes}m ago`

				const hours = Math.floor(totalMinutes / 60)
				if (hours < 24) return `${hours}h ago`

				const days = Math.floor(hours / 24)
				return `${days}d ago`
			} catch (error) {
				console.error('Error formatting last token time:', error)
				return 'Never'
			}
		},

		getTokenAgeMinutes(timestamp) {
			const age = new Date() - new Date(timestamp)
			return Math.floor(age / 60000)
		},

		getAgeColor(timestamp) {
			const minutes = this.getTokenAgeMinutes(timestamp)
			if (minutes < 5) return 'badge-success'
			if (minutes < 30) return 'badge-warning'
			return 'badge-error'
		},

		getAgeStatus(timestamp) {
			const minutes = this.getTokenAgeMinutes(timestamp)
			if (minutes < 5) return 'Fresh'
			if (minutes < 30) return 'Recent'
			return 'Old'
		},

		isNewToken(token) {
			return this.newTokenIds.has(token.id)
		},

		getTokenIndex(token) {
			// Calculate sequential index based on position in filtered list
			const index = this.filteredTokens.findIndex(t => t.id === token.id)
			return index >= 0 ? (index + 1).toString() : token.id.toString()
		},

		getTokenDisplayName(token) {
			// Priority: Name > Symbol > Token ID
			if (token.name && token.name.trim()) {
				return token.name.trim()
			}
			if (token.symbol && token.symbol.trim()) {
				return `$${token.symbol.trim()}`
			}
			return `Token #${token.id}`
		},

		getPlatformColor(platform) {
			switch (platform?.toLowerCase()) {
				case 'pumpfun':
					return 'badge-primary'
				case 'raydium':
					return 'badge-secondary'
				case 'launchlab':
					return 'badge-accent'
				default:
					return 'badge-neutral'
			}
		},
	},
}
</script>
