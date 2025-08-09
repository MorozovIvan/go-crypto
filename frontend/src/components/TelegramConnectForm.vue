<template>
	<div class="card bg-base-200 shadow-xl max-w-md mx-auto">
		<div class="card-body">
			<h2 class="card-title text-2xl justify-center mb-6">
				Connect to Telegram
			</h2>

			<!-- Phone Number Form -->
			<form
				v-if="!showVerification"
				@submit.prevent="handlePhoneSubmit"
				class="space-y-4"
			>
				<div class="form-control">
					<label class="label" for="phone">
						<span class="label-text">Phone Number</span>
					</label>
					<div class="relative">
						<span
							class="absolute inset-y-0 left-0 flex items-center pl-3 opacity-60"
							>+</span
						>
						<input
							type="tel"
							id="phone"
							v-model="phone"
							placeholder="1234567890"
							class="input input-bordered w-full pl-8"
							:class="{ 'input-error': phoneError }"
							@input="formatPhoneNumber"
							maxlength="15"
							:disabled="loading || isFlooded"
						/>
					</div>
					<label class="label">
						<span v-if="phoneError" class="label-text-alt text-error">{{
							phoneError
						}}</span>
						<span v-else class="label-text-alt"
							>Enter your phone number with country code</span
						>
					</label>
				</div>
				<button
					type="submit"
					class="btn btn-primary w-full"
					:disabled="loading || !isValidPhone || isFlooded"
				>
					<span
						v-if="loading"
						class="loading loading-spinner loading-sm"
					></span>
					<span v-if="isFlooded"
						>Please wait {{ formatWaitTime(floodWaitTime) }}</span
					>
					<span v-else>{{
						loading ? 'Sending Code...' : 'Send Verification Code'
					}}</span>
				</button>
			</form>

			<!-- Verification Code Form -->
			<form v-else @submit.prevent="handleVerificationSubmit" class="space-y-4">
				<div class="form-control">
					<label class="label" for="code">
						<span class="label-text">Verification Code</span>
					</label>
					<div class="relative">
						<input
							type="text"
							id="code"
							v-model="code"
							placeholder="Enter code"
							class="input input-bordered w-full"
							:class="{ 'border-red-500': codeError }"
							maxlength="5"
							@input="formatCode"
							:disabled="loading || isFlooded"
						/>
						<div
							v-if="countdown > 0"
							class="absolute text-sm text-gray-500 transform -translate-y-1/2 right-3 top-1/2"
						>
							Resend in {{ countdown }}s
						</div>
					</div>
					<label class="label">
						<span v-if="codeError" class="label-text-alt text-error">{{
							codeError
						}}</span>
						<span v-else class="label-text-alt"
							>Enter the code sent to your Telegram</span
						>
					</label>
				</div>
				<div class="flex space-x-4">
					<button
						type="button"
						@click="resetForm"
						class="btn btn-outline flex-1"
						:disabled="loading || isFlooded"
					>
						Back
					</button>
					<button
						type="submit"
						class="btn btn-primary flex-1"
						:disabled="loading || !isValidCode || isFlooded"
					>
						<span
							v-if="loading"
							class="loading loading-spinner loading-sm"
						></span>
						<span v-if="isFlooded"
							>Please wait {{ formatWaitTime(floodWaitTime) }}</span
						>
						<span v-else>{{ loading ? 'Verifying...' : 'Verify Code' }}</span>
					</button>
				</div>
			</form>

			<!-- Error Alert -->
			<div v-if="generalError" class="alert alert-error mt-4">
				<span>{{ generalError }}</span>
			</div>
		</div>
	</div>
</template>

<script>
export default {
	name: 'TelegramConnectForm',
	data() {
		return {
			phone: '',
			code: '',
			phoneError: '',
			codeError: '',
			generalError: '',
			loading: false,
			showVerification: false,
			hash: null,
			countdown: 0,
			countdownInterval: null,
			retryCount: 0,
			maxRetries: 3,
			isFlooded: false,
			floodWaitTime: 0,
			floodInterval: null,
		}
	},
	mounted() {
		// Clear any existing flood state when component is mounted
		this.resetFloodState()
	},
	computed: {
		isValidPhone() {
			// Basic phone validation (at least 10 digits)
			return /^\d{10,15}$/.test(this.phone.replace(/\D/g, ''))
		},
		isValidCode() {
			// Basic code validation (5 digits)
			return /^\d{5}$/.test(this.code)
		},
	},
	methods: {
		resetFloodState() {
			if (this.floodInterval) {
				clearInterval(this.floodInterval)
				this.floodInterval = null
			}
			this.isFlooded = false
			this.floodWaitTime = 0
		},
		formatWaitTime(seconds) {
			if (seconds < 60) {
				return `${seconds} seconds`
			} else if (seconds < 3600) {
				const minutes = Math.floor(seconds / 60)
				return `${minutes} minute${minutes > 1 ? 's' : ''}`
			} else {
				const hours = Math.floor(seconds / 3600)
				return `${hours} hour${hours > 1 ? 's' : ''}`
			}
		},
		startFloodTimer(seconds) {
			// Clear any existing flood timer
			this.resetFloodState()

			this.isFlooded = true
			this.floodWaitTime = seconds
			this.floodInterval = setInterval(() => {
				if (this.floodWaitTime > 0) {
					this.floodWaitTime--
				} else {
					this.resetFloodState()
				}
			}, 1000)
		},
		formatPhoneNumber() {
			// Remove all non-digit characters
			this.phone = this.phone.replace(/\D/g, '')
			this.phoneError = ''
			this.generalError = ''
		},
		formatCode() {
			// Remove all non-digit characters
			this.code = this.code.replace(/\D/g, '')
			this.codeError = ''
			this.generalError = ''
		},
		startCountdown() {
			this.countdown = 30
			this.countdownInterval = setInterval(() => {
				if (this.countdown > 0) {
					this.countdown--
				} else {
					clearInterval(this.countdownInterval)
				}
			}, 1000)
		},
		async handlePhoneSubmit() {
			this.phoneError = ''
			this.generalError = ''
			this.loading = true

			try {
				if (!this.isValidPhone) {
					throw new Error('Please enter a valid phone number')
				}

				console.log('Sending phone number:', '+' + this.phone)
				const response = await fetch(
					`http://localhost:8080/api/telegram/auth/callback?phone=${encodeURIComponent(
						'+' + this.phone
					)}`
				)
				console.log('Response status:', response.status)
				const data = await response.json()
				console.log('Response data:', data)

				if (!response.ok) {
					console.error('Response not OK:', data)
					if (
						data.error &&
						(data.error.includes('FLOOD_WAIT') ||
							data.error.includes('code 420'))
					) {
						const waitTime = parseInt(
							data.error.match(/\((\d+)\)/)?.[1] || '30'
						)
						this.startFloodTimer(waitTime)
						throw new Error(
							`Too many attempts. Please wait ${this.formatWaitTime(
								waitTime
							)} before trying again.`
						)
					}
					throw new Error(data.error || 'Failed to start authentication')
				}

				if (!data.success || !data.hash) {
					console.error('Missing success or hash:', data)
					throw new Error('Failed to get phone code hash')
				}

				console.log('Setting hash:', data.hash)
				this.hash = data.hash
				this.showVerification = true
				this.startCountdown()
				this.retryCount = 0
			} catch (err) {
				console.error('Error in handlePhoneSubmit:', err)
				if (err.message.includes('Too many attempts')) {
					this.generalError = err.message
				} else {
					this.phoneError = err.message
					this.generalError =
						'Failed to send verification code. Please try again.'
				}
			} finally {
				this.loading = false
			}
		},
		async handleVerificationSubmit() {
			this.codeError = ''
			this.generalError = ''
			this.loading = true

			try {
				if (!this.isValidCode) {
					throw new Error('Please enter a valid 5-digit code')
				}

				const response = await fetch(
					'http://localhost:8080/api/telegram/verify-code',
					{
						method: 'POST',
						headers: {
							'Content-Type': 'application/json',
						},
						body: JSON.stringify({
							phone: '+' + this.phone,
							code: this.code,
							hash: this.hash,
						}),
					}
				)

				if (!response.ok) {
					const errorData = await response.json()

					if (
						errorData.error &&
						errorData.error.includes('PHONE_PASSWORD_FLOOD')
					) {
						const waitTime = parseInt(errorData.error.match(/\d+/)?.[0] || '30')
						this.startFloodTimer(waitTime)
						throw new Error(
							`Too many attempts. Please wait ${this.formatWaitTime(
								waitTime
							)} before trying again.`
						)
					}

					if (errorData.error === '2FA_REQUIRED') {
						this.$emit('2fa-required', '+' + this.phone)
						return
					}

					this.retryCount++
					if (this.retryCount >= this.maxRetries) {
						throw new Error(
							'Too many failed attempts. Please try again with a new code.'
						)
					}

					throw new Error(errorData.message || 'Verification failed')
				}

				const data = await response.json()
				if (data.success) {
					this.$emit('connected', data.user_id)
					// Refresh the page to ensure clean state after authentication
					setTimeout(() => {
						window.location.reload()
					}, 1000)
				}
			} catch (err) {
				this.codeError = err.message
				this.generalError = 'Failed to verify code. Please try again.'
			} finally {
				this.loading = false
			}
		},
		resetForm() {
			this.phone = ''
			this.code = ''
			this.phoneError = ''
			this.codeError = ''
			this.generalError = ''
			this.showVerification = false
			this.hash = null
			this.retryCount = 0
			if (this.countdownInterval) {
				clearInterval(this.countdownInterval)
				this.countdown = 0
			}
			this.resetFloodState()
		},
	},
	beforeDestroy() {
		if (this.countdownInterval) {
			clearInterval(this.countdownInterval)
		}
		this.resetFloodState()
	},
	emits: ['connected', '2fa-required'],
}
</script>
