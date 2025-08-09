<template>
	<div class="min-h-screen bg-base-100" data-theme="crypto">
		<!-- Professional Crypto Admin Dashboard -->
		<div class="drawer lg:drawer-open">
			<input id="drawer-toggle" type="checkbox" class="drawer-toggle" />

			<!-- Main Content -->
			<div class="drawer-content flex flex-col">
				<!-- Top Navigation Bar -->
				<div class="navbar bg-base-200 shadow-lg border-b border-base-300">
					<div class="navbar-start">
						<!-- Mobile Menu Button -->
						<label
							for="drawer-toggle"
							class="btn btn-square btn-ghost lg:hidden"
						>
							<svg
								class="w-6 h-6"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M4 6h16M4 12h16M4 18h16"
								></path>
							</svg>
						</label>
					</div>

					<div class="navbar-center hidden lg:flex">
						<!-- Breadcrumb Navigation -->
						<div class="breadcrumbs text-[10px]">
							<ul>
								<li><a>Dashboard</a></li>
								<li class="text-primary font-semibold">
									{{ getCurrentViewTitle() }}
								</li>
							</ul>
						</div>
					</div>

					<div class="navbar-end gap-2">
						<!-- Wallet Connection -->
						<WalletConnection
							@wallet-connected="handleWalletConnected"
							@wallet-disconnected="handleWalletDisconnected"
						/>

						<!-- User Menu -->
						<div v-if="isConnected" class="dropdown dropdown-end">
							<div
								tabindex="0"
								role="button"
								class="btn btn-ghost btn-circle avatar"
							>
								<div
									class="w-10 rounded-full bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center border-2 border-primary/30"
								>
									<span class="text-primary font-bold text-sm">{{
										getUserInitials()
									}}</span>
								</div>
							</div>
							<ul
								tabindex="0"
								class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow-lg bg-base-200 rounded-box w-64 border border-base-300"
							>
								<li class="menu-title">
									<span class="text-base-content">Account</span>
								</li>
								<li>
									<div class="flex items-center gap-3 p-2">
										<div class="avatar">
											<div class="w-8 rounded-full bg-primary/20">
												<div
													class="flex items-center justify-center w-full h-full text-primary text-xs font-bold"
												>
													{{ getUserInitials() }}
												</div>
											</div>
										</div>
										<div>
											<div class="font-semibold">User {{ userId }}</div>
											<div class="text-xs opacity-70">Premium Account</div>
										</div>
									</div>
								</li>
								<div class="divider my-1"></div>
								<li>
									<a
										><svg
											class="w-4 h-4"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
											></path></svg
										>Profile Settings</a
									>
								</li>
								<li>
									<a
										><svg
											class="w-4 h-4"
											fill="none"
											stroke="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
											></path>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
											></path></svg
										>Preferences</a
									>
								</li>
								<div class="divider my-1"></div>
								<li>
									<a @click="logout" class="text-error"
										><svg
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
											></path></svg
										>Logout</a
									>
								</li>
							</ul>
						</div>
					</div>
				</div>

				<!-- Main Content Area -->
				<main class="flex-1 p-6 bg-base-100">
					<div class="max-w-7xl mx-auto">
						<!-- Page Header -->
						<div class="mb-6">
							<div class="flex items-center justify-between">
								<div>
									<h2 class="text-2xl font-bold text-base-content">
										{{ getCurrentViewTitle() }}
									</h2>
									<p class="text-base-content/70 mt-1">
										{{ getCurrentViewDescription() }}
									</p>
								</div>
								<div class="flex items-center gap-2">
									<!-- Quick Actions -->
									<button class="btn btn-outline btn-sm">
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
												d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
											></path>
										</svg>
										Refresh
									</button>
									<button class="btn btn-primary btn-sm">
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
												d="M12 6v6m0 0v6m0-6h6m-6 0H6"
											></path>
										</svg>
										New Analysis
									</button>
								</div>
							</div>
						</div>

						<!-- Dynamic Content -->
						<div class="space-y-6">
							<div v-if="currentView === 'dashboard'">
								<DashboardOverview @navigate="handleViewChange" />
							</div>
							<div v-else-if="currentView === 'telegram'">
								<MainContent
									:is-connected="isConnected"
									:user-id="userId"
									@connect="handleConnect"
									@2fa-required="handle2FARequired"
									@logout="logout"
									@disconnect="handleDisconnect"
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
							<div v-else-if="currentView === 'token-analysis'">
								<TokenAnalysis />
							</div>
						</div>
					</div>
				</main>

				<!-- Footer -->
				<footer
					class="footer footer-center p-4 bg-base-200 text-base-content border-t border-base-300"
				>
					<aside class="flex items-center gap-4">
						<p class="text-sm">© 2024 CryptoTrends Analytics Platform</p>
						<div class="flex items-center gap-2 text-xs opacity-70">
							<span>Status:</span>
							<div class="badge badge-success badge-xs">Online</div>
						</div>
					</aside>
				</footer>
			</div>

			<!-- Sidebar Navigation -->
			<div class="drawer-side">
				<label for="drawer-toggle" class="drawer-overlay"></label>
				<aside class="w-72 min-h-full bg-base-200 border-r border-base-300">
					<!-- Sidebar Header -->
					<div class="p-4 border-b border-base-300">
						<div class="flex items-center gap-3">
							<div class="avatar">
								<div
									class="w-10 rounded-lg bg-gradient-to-br from-primary to-secondary"
								>
									<div class="flex items-center justify-center w-full h-full">
										<svg
											class="w-6 h-6 text-white"
											fill="currentColor"
											viewBox="0 0 24 24"
										>
											<path
												d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"
											/>
										</svg>
									</div>
								</div>
							</div>
							<div>
								<h1 class="text-lg font-bold">CryptoTrends</h1>
								<p class="text-xs opacity-70">Analytics Platform</p>
							</div>
						</div>
					</div>

					<!-- Navigation Menu -->
					<div class="p-4">
						<ul class="menu menu-vertical w-full space-y-1">
							<!-- Dashboard -->
							<li>
								<a
									@click="handleViewChange('dashboard')"
									:class="{ active: currentView === 'dashboard' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7zm0 0a2 2 0 012-2h6l2 2h6a2 2 0 012 2v2H3V7z"
										></path>
									</svg>
									<div>
										<div class="font-medium">Dashboard</div>
										<div class="text-xs opacity-70">
											Overview & quick access
										</div>
									</div>
								</a>
							</li>

							<!-- Analytics Section -->
							<li class="menu-title mt-4">
								<span class="text-base-content/70">Analytics</span>
							</li>
							<li>
								<a
									@click="handleViewChange('telegram')"
									:class="{ active: currentView === 'telegram' }"
									class="flex items-center gap-3"
								>
									<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
										<path
											d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.07-.2-.08-.06-.19-.04-.27-.02-.12.02-1.96 1.25-5.54 3.69-.52.36-1 .53-1.42.52-.47-.01-1.37-.26-2.03-.48-.82-.27-1.47-.42-1.42-.88.03-.24.29-.49.8-.75 3.12-1.36 5.2-2.26 6.24-2.7 2.98-1.24 3.6-1.46 4.01-1.46.09 0 .28.02.4.09.11.06.19.14.22.25.02.1.02.21.01.3z"
										/>
									</svg>
									<div>
										<div class="font-medium">Telegram Signals</div>
										<div class="text-xs opacity-70">Track trading signals</div>
									</div>
								</a>
							</li>
							<li>
								<a
									@click="handleViewChange('market')"
									:class="{ active: currentView === 'market' }"
									class="flex items-center gap-3"
								>
									<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
										<path
											d="M3.5 18.49l6-6.01 4 4L22 6.92l-1.41-1.41-7.09 7.97-4-4L2 16.99z"
										/>
									</svg>
									<div>
										<div class="font-medium">Market Analysis</div>
										<div class="text-xs opacity-70">Real-time market data</div>
									</div>
								</a>
							</li>
							<li>
								<a
									@click="handleViewChange('token-analysis')"
									:class="{ active: currentView === 'token-analysis' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
										/>
									</svg>
									<div>
										<div class="font-medium">Token Analysis</div>
										<div class="text-xs opacity-70">Deep token insights</div>
									</div>
								</a>
							</li>

							<!-- Trading Section -->
							<li class="menu-title mt-4">
								<span class="text-base-content/70">Trading</span>
							</li>
							<li>
								<a
									@click="handleViewChange('solana-wallets')"
									:class="{ active: currentView === 'solana-wallets' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<rect
											x="2"
											y="7"
											width="20"
											height="10"
											rx="2"
											stroke-width="2"
										/>
										<path
											d="M16 12h.01"
											stroke-width="2"
											stroke-linecap="round"
										/>
									</svg>
									<div>
										<div class="font-medium">Profitable Wallets</div>
										<div class="text-xs opacity-70">Top performing wallets</div>
									</div>
								</a>
							</li>
							<li>
								<a
									@click="handleViewChange('portfolio')"
									:class="{ active: currentView === 'portfolio' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
										/>
									</svg>
									<div>
										<div class="font-medium">Portfolio</div>
										<div class="text-xs opacity-70">Manage your assets</div>
									</div>
								</a>
							</li>

							<!-- Advanced Section -->
							<li class="menu-title mt-4">
								<span class="text-base-content/70">Advanced</span>
							</li>
							<li>
								<a
									@click="handleViewChange('arbitration')"
									:class="{ active: currentView === 'arbitration' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
										/>
									</svg>
									<div>
										<div class="font-medium">Arbitrage</div>
										<div class="text-xs opacity-70">
											Cross-exchange opportunities
										</div>
									</div>
								</a>
							</li>
							<li>
								<a
									@click="handleViewChange('sniping')"
									:class="{ active: currentView === 'sniping' }"
									class="flex items-center gap-3"
								>
									<svg
										class="w-5 h-5"
										fill="none"
										stroke="currentColor"
										viewBox="0 0 24 24"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M13 10V3L4 14h7v7l9-11h-7z"
										/>
									</svg>
									<div>
										<div class="font-medium">Token Sniping</div>
										<div class="text-xs opacity-70">Early token detection</div>
									</div>
								</a>
							</li>
						</ul>
					</div>

					<!-- Sidebar Footer -->
					<div
						class="absolute bottom-0 left-0 right-0 p-4 border-t border-base-300 bg-base-200"
					>
						<div class="flex items-center justify-between text-xs opacity-70">
							<span>v2.1.0</span>
							<div class="flex items-center gap-1">
								<div
									class="w-2 h-2 bg-success rounded-full animate-pulse"
								></div>
								<span>Online</span>
							</div>
						</div>
					</div>
				</aside>
			</div>
		</div>

		<!-- Enhanced 2FA Modal -->
		<dialog :class="{ 'modal modal-open': show2FAModal }" class="modal">
			<div class="modal-box max-w-md">
				<div class="flex items-center gap-3 mb-4">
					<div class="avatar">
						<div
							class="w-12 rounded-full bg-primary/20 flex items-center justify-center"
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
									d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
								></path>
							</svg>
						</div>
					</div>
					<div>
						<h3 class="font-bold text-lg">Two-Factor Authentication</h3>
						<p class="text-sm opacity-70">Security verification required</p>
					</div>
				</div>

				<div class="alert alert-info mb-4">
					<svg
						class="w-6 h-6"
						fill="none"
						stroke="currentColor"
						viewBox="0 0 24 24"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						></path>
					</svg>
					<span>Please enter your 2FA code from the Telegram app.</span>
				</div>

				<div class="form-control w-full">
					<label class="label">
						<span class="label-text">Authentication Code</span>
					</label>
					<input
						type="text"
						v-model="twoFACode"
						placeholder="Enter a code"
						class="input input-bordered w-full"
						maxlength="60"
					/>
				</div>

				<div class="modal-action">
					<button @click="show2FAModal = false" class="btn btn-ghost">
						Cancel
					</button>
					<button
						@click="handle2FASubmit"
						class="btn btn-primary"
						:disabled="submitting2FA || !twoFACode"
					>
						<span
							v-if="submitting2FA"
							class="loading loading-spinner loading-sm"
						></span>
						{{ submitting2FA ? 'Verifying...' : 'Verify Code' }}
					</button>
				</div>
			</div>
			<form method="dialog" class="modal-backdrop">
				<button @click="show2FAModal = false">close</button>
			</form>
		</dialog>
	</div>
</template>

<script>
import DashboardOverview from './components/DashboardOverview.vue'
import MainContent from './components/MainContent.vue'
import MarketAnalysis from './components/MarketAnalysis.vue'
import ProfitableSolanaWallets from './components/ProfitableSolanaWallets.vue'
import Portfolio from './components/Portfolio.vue'
import ArbitrationDashboard from './components/ArbitrationDashboard.vue'
import SnipingDashboard from './components/SnipingDashboard.vue'
import TokenAnalysis from './components/TokenAnalysis.vue'
import WalletConnection from './components/WalletConnection.vue'

const API_BASE_URL = 'http://localhost:8080'

export default {
	name: 'App',
	components: {
		DashboardOverview,
		MainContent,
		MarketAnalysis,
		ProfitableSolanaWallets,
		Portfolio,
		ArbitrationDashboard,
		SnipingDashboard,
		TokenAnalysis,
		WalletConnection,
	},
	data() {
		return {
			isConnected: false,
			userId: null,
			show2FAModal: false,
			twoFACode: '',
			currentView: 'dashboard',
			walletAddress: null,
			currentPhone: '',
			submitting2FA: false,
		}
	},
	async mounted() {
		// Check authentication status on app startup
		console.log('App mounted, checking auth status...')
		await this.checkAuthStatus()
		console.log('Auth status checked:', {
			isConnected: this.isConnected,
			userId: this.userId,
		})
	},
	methods: {
		async checkAuthStatus() {
			try {
				console.log(
					'Fetching auth status from:',
					`${API_BASE_URL}/api/telegram/status`
				)
				const response = await fetch(`${API_BASE_URL}/api/telegram/status`)
				const data = await response.json()
				console.log('Auth status response:', data)

				if (data.status && data.status.authenticated && data.status.user_id) {
					this.isConnected = true
					this.userId = data.status.user_id
					console.log('Restored authenticated session:', data.status)
				} else {
					this.isConnected = false
					this.userId = null
					console.log('No authenticated session found')
				}
			} catch (error) {
				console.error('Failed to check auth status:', error)
				this.isConnected = false
				this.userId = null
			}
		},

		getCurrentViewTitle() {
			const titles = {
				dashboard: 'Dashboard Overview',
				telegram: 'Telegram Signal Analysis',
				market: 'Market Analysis',
				'token-analysis': 'Token Analysis',
				'solana-wallets': 'Profitable Wallets',
				portfolio: 'Portfolio Management',
				arbitration: 'Arbitrage Opportunities',
				sniping: 'Token Sniping',
			}
			return titles[this.currentView] || 'Dashboard'
		},

		getCurrentViewDescription() {
			const descriptions = {
				dashboard:
					'Welcome to your professional crypto trading analytics platform',
				telegram:
					'Track and analyze Telegram trading signals with advanced metrics',
				market: 'Real-time market data and comprehensive trading analytics',
				'token-analysis':
					'Deep dive analysis of Solana tokens and pump.fun projects',
				'solana-wallets':
					'Discover and track the most profitable Solana wallets',
				portfolio: 'Manage your crypto portfolio with advanced tracking tools',
				arbitration: 'Find profitable arbitrage opportunities across exchanges',
				sniping: 'Early detection and analysis of new token launches',
			}
			return (
				descriptions[this.currentView] ||
				'Professional crypto trading analytics platform'
			)
		},

		getUserInitials() {
			if (this.userId) {
				return this.userId.toString().slice(-2).toUpperCase()
			}
			return 'U'
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
			console.log('handleConnect called with userId:', userId)
			this.isConnected = true
			this.userId = userId
			console.log('Updated state:', {
				isConnected: this.isConnected,
				userId: this.userId,
			})

			// Force a re-check of auth status to ensure consistency
			setTimeout(() => {
				this.checkAuthStatus()
			}, 1000)
		},
		async handle2FASubmit() {
			if (!this.twoFACode) {
				return
			}

			this.submitting2FA = true
			try {
				const controller = new AbortController()
				const timeoutId = setTimeout(() => controller.abort(), 10000) // 10 second timeout

				const response = await fetch(
					`${API_BASE_URL}/api/telegram/verify-2fa`,
					{
						method: 'POST',
						headers: {
							'Content-Type': 'application/json',
						},
						body: JSON.stringify({
							phone: this.currentPhone,
							password: this.twoFACode,
						}),
						signal: controller.signal,
					}
				)

				clearTimeout(timeoutId)
				const data = await response.json()

				if (data.success) {
					this.isConnected = true
					this.userId = data.user_id
					this.show2FAModal = false
					this.twoFACode = ''
					console.log('2FA verification successful, refreshing page...')
					// Refresh the page to ensure clean state
					setTimeout(() => {
						window.location.reload()
					}, 500)
				} else {
					alert(data.message || '2FA verification failed')
				}
			} catch (error) {
				if (error.name === 'AbortError') {
					alert('Request timed out. Please try again.')
				} else {
					console.error('2FA verification error:', error)
					alert('Failed to verify 2FA code. Please try again.')
				}
			} finally {
				this.submitting2FA = false
			}
		},
		handle2FARequired(phone) {
			this.currentPhone = phone
			this.show2FAModal = true
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
		},
		handleDisconnect() {
			this.isConnected = false
			this.userId = null
			console.log('Telegram session disconnected due to expiration')
		},
	},
}
</script>

<style>
/* Remove all existing styles as we're using Tailwind now */
</style>
