<template>
	<div class="space-y-6">
		<!-- Welcome Section -->
		<div
			class="hero bg-gradient-to-br from-primary/10 to-secondary/10 rounded-box"
		>
			<div class="hero-content text-center py-12">
				<div class="max-w-md">
					<h1 class="text-4xl font-bold text-gradient mb-4">
						Welcome to CryptoTrends
					</h1>
					<p class="text-lg opacity-80 mb-6">
						Professional crypto trading analytics platform powered by AI
					</p>
					<div class="flex flex-wrap gap-2 justify-center">
						<div class="badge badge-primary">Real-time Analysis</div>
						<div class="badge badge-secondary">AI-Powered</div>
						<div class="badge badge-accent">Multi-Chain</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Quick Stats -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			<div class="stat bg-base-200 rounded-box">
				<div class="stat-figure text-primary">
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
							d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
						></path>
					</svg>
				</div>
				<div class="stat-title">Active Signals</div>
				<div class="stat-value text-primary">1,247</div>
				<div class="stat-desc">↗︎ 12% from yesterday</div>
			</div>

			<div class="stat bg-base-200 rounded-box">
				<div class="stat-figure text-secondary">
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
				<div class="stat-title">Total Volume</div>
				<div class="stat-value text-secondary">$2.4M</div>
				<div class="stat-desc">↗︎ 8% from last week</div>
			</div>

			<div class="stat bg-base-200 rounded-box">
				<div class="stat-figure text-accent">
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
							d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
						></path>
					</svg>
				</div>
				<div class="stat-title">Win Rate</div>
				<div class="stat-value text-accent">73.2%</div>
				<div class="stat-desc">↗︎ 2.1% from last month</div>
			</div>

			<div class="stat bg-base-200 rounded-box">
				<div class="stat-figure text-success">
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
							d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
						></path>
					</svg>
				</div>
				<div class="stat-title">Active Users</div>
				<div class="stat-value text-success">12.4K</div>
				<div class="stat-desc">↗︎ 18% from last month</div>
			</div>
		</div>

		<!-- Feature Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
			<!-- Telegram Analysis -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-blue-500 to-cyan-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
								>
									<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
										<path
											d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.07-.2-.08-.06-.19-.04-.27-.02-.12.02-1.96 1.25-5.54 3.69-.52.36-1 .53-1.42.52-.47-.01-1.37-.26-2.03-.48-.82-.27-1.47-.42-1.42-.88.03-.24.29-.49.8-.75 3.12-1.36 5.2-2.26 6.24-2.7 2.98-1.24 3.6-1.46 4.01-1.46.09 0 .28.02.4.09.11.06.19.14.22.25.02.1.02.21.01.3z"
										/>
									</svg>
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Telegram Signals</h3>
							<p class="text-sm opacity-70">AI-powered signal analysis</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Track and analyze trading signals from Telegram channels with
						advanced metrics and performance tracking.
					</p>
					<div class="card-actions justify-end">
						<button
							@click="$emit('navigate', 'telegram')"
							class="btn btn-primary btn-sm"
						>
							Get Started
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Market Analysis -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-green-500 to-emerald-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
								>
									<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
										<path
											d="M3.5 18.49l6-6.01 4 4L22 6.92l-1.41-1.41-7.09 7.97-4-4L2 16.99z"
										/>
									</svg>
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Market Analysis</h3>
							<p class="text-sm opacity-70">Real-time market insights</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Comprehensive market analysis with real-time data, technical
						indicators, and trend predictions.
					</p>
					<div class="card-actions justify-end">
						<button
							@click="$emit('navigate', 'market')"
							class="btn btn-secondary btn-sm"
						>
							Explore
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Token Analysis -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-purple-500 to-pink-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
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
											d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
										/>
									</svg>
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Token Analysis</h3>
							<p class="text-sm opacity-70">Deep token insights</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Analyze Solana tokens and pump.fun projects with comprehensive
						metrics and security analysis.
					</p>
					<div class="card-actions justify-end">
						<button
							@click="$emit('navigate', 'token-analysis')"
							class="btn btn-accent btn-sm"
						>
							Analyze
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Profitable Wallets -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-orange-500 to-red-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
								>
									<svg
										class="w-6 h-6"
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
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Profitable Wallets</h3>
							<p class="text-sm opacity-70">Top performers</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Discover and track the most profitable Solana wallets and copy their
						trading strategies.
					</p>
					<div class="card-actions justify-end">
						<button
							@click="$emit('navigate', 'solana-wallets')"
							class="btn btn-warning btn-sm"
						>
							Discover
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Portfolio -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-indigo-500 to-blue-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
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
											d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
										/>
									</svg>
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Portfolio</h3>
							<p class="text-sm opacity-70">Asset management</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Manage your crypto portfolio with advanced tracking tools and
						performance analytics.
					</p>
					<div class="card-actions justify-end">
						<button
							@click="$emit('navigate', 'portfolio')"
							class="btn btn-info btn-sm"
						>
							Manage
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								></path>
							</svg>
						</button>
					</div>
				</div>
			</div>

			<!-- Advanced Tools -->
			<div
				class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 group"
			>
				<div class="card-body">
					<div class="flex items-center gap-3 mb-4">
						<div class="avatar">
							<div
								class="w-12 rounded-lg bg-gradient-to-br from-teal-500 to-cyan-500"
							>
								<div
									class="flex items-center justify-center w-full h-full text-white"
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
											d="M13 10V3L4 14h7v7l9-11h-7z"
										/>
									</svg>
								</div>
							</div>
						</div>
						<div>
							<h3 class="card-title text-lg">Advanced Tools</h3>
							<p class="text-sm opacity-70">Pro features</p>
						</div>
					</div>
					<p class="text-sm opacity-80 mb-4">
						Access arbitrage opportunities, token sniping, and other advanced
						trading tools.
					</p>
					<div class="dropdown dropdown-top dropdown-end">
						<div tabindex="0" role="button" class="btn btn-success btn-sm">
							Tools
							<svg
								class="w-4 h-4 ml-1"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M19 9l-7 7-7-7"
								></path>
							</svg>
						</div>
						<ul
							tabindex="0"
							class="dropdown-content z-[1] menu p-2 shadow bg-base-200 rounded-box w-52"
						>
							<li>
								<a @click="$emit('navigate', 'arbitration')">Arbitrage</a>
							</li>
							<li>
								<a @click="$emit('navigate', 'sniping')">Token Sniping</a>
							</li>
						</ul>
					</div>
				</div>
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="card bg-base-200 shadow-xl">
			<div class="card-body">
				<div class="flex items-center justify-between mb-4">
					<h2 class="card-title">Recent Activity</h2>
					<button class="btn btn-ghost btn-sm">
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
				</div>

				<div class="space-y-3">
					<div class="flex items-center gap-3 p-3 bg-base-100 rounded-lg">
						<div class="avatar">
							<div class="w-8 rounded-full bg-success/20">
								<div
									class="flex items-center justify-center w-full h-full text-success text-xs"
								>
									↗
								</div>
							</div>
						</div>
						<div class="flex-1">
							<p class="text-sm font-medium">New profitable signal detected</p>
							<p class="text-xs opacity-70">
								SOL/USDC pair showing 15% gain potential
							</p>
						</div>
						<div class="text-xs opacity-70">2m ago</div>
					</div>

					<div class="flex items-center gap-3 p-3 bg-base-100 rounded-lg">
						<div class="avatar">
							<div class="w-8 rounded-full bg-primary/20">
								<div
									class="flex items-center justify-center w-full h-full text-primary text-xs"
								>
									📊
								</div>
							</div>
						</div>
						<div class="flex-1">
							<p class="text-sm font-medium">Market analysis updated</p>
							<p class="text-xs opacity-70">
								New technical indicators available
							</p>
						</div>
						<div class="text-xs opacity-70">5m ago</div>
					</div>

					<div class="flex items-center gap-3 p-3 bg-base-100 rounded-lg">
						<div class="avatar">
							<div class="w-8 rounded-full bg-warning/20">
								<div
									class="flex items-center justify-center w-full h-full text-warning text-xs"
								>
									⚡
								</div>
							</div>
						</div>
						<div class="flex-1">
							<p class="text-sm font-medium">New token launched</p>
							<p class="text-xs opacity-70">Early detection on pump.fun</p>
						</div>
						<div class="text-xs opacity-70">12m ago</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
export default {
	name: 'DashboardOverview',
	emits: ['navigate'],
}
</script>
