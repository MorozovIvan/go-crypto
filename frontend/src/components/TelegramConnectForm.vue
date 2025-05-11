<template>
  <div class="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
    <h2 class="text-2xl font-bold mb-6 text-center">Connect to Telegram</h2>
    
    <!-- Phone Number Form -->
    <form v-if="!showVerification" @submit.prevent="handlePhoneSubmit" class="space-y-4">
      <div>
        <label for="phone" class="block text-sm font-medium text-gray-700 mb-1">Phone Number</label>
        <div class="relative">
          <span class="absolute inset-y-0 left-0 pl-3 flex items-center text-gray-500">+</span>
          <input
            type="tel"
            id="phone"
            v-model="phone"
            placeholder="1234567890"
            class="pl-8 w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0088cc] focus:border-transparent"
            :class="{ 'border-red-500': phoneError }"
            @input="formatPhoneNumber"
            maxlength="15"
            :disabled="loading"
          />
        </div>
        <p v-if="phoneError" class="mt-1 text-sm text-red-600">{{ phoneError }}</p>
        <p class="mt-1 text-sm text-gray-500">Enter your phone number with country code (e.g., 1234567890)</p>
      </div>
      <button
        type="submit"
        class="w-full bg-[#0088cc] text-white px-4 py-2 rounded-lg hover:bg-[#0077b3] transition-colors flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed"
        :disabled="loading || !isValidPhone"
      >
        <svg v-if="loading" class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>{{ loading ? 'Sending Code...' : 'Send Verification Code' }}</span>
      </button>
    </form>

    <!-- Verification Code Form -->
    <form v-else @submit.prevent="handleVerificationSubmit" class="space-y-4">
      <div>
        <label for="code" class="block text-sm font-medium text-gray-700 mb-1">Verification Code</label>
        <div class="relative">
          <input
            type="text"
            id="code"
            v-model="code"
            placeholder="Enter code"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0088cc] focus:border-transparent"
            :class="{ 'border-red-500': codeError }"
            maxlength="5"
            @input="formatCode"
            :disabled="loading"
          />
          <div v-if="countdown > 0" class="absolute right-3 top-1/2 transform -translate-y-1/2 text-sm text-gray-500">
            Resend in {{ countdown }}s
          </div>
        </div>
        <p v-if="codeError" class="mt-1 text-sm text-red-600">{{ codeError }}</p>
        <p class="mt-1 text-sm text-gray-500">Enter the code sent to your Telegram</p>
      </div>
      <div class="flex space-x-4">
        <button
          type="button"
          @click="resetForm"
          class="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading"
        >
          Back
        </button>
        <button
          type="submit"
          class="flex-1 bg-[#0088cc] text-white px-4 py-2 rounded-lg hover:bg-[#0077b3] transition-colors flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading || !isValidCode"
        >
          <svg v-if="loading" class="animate-spin h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span>{{ loading ? 'Verifying...' : 'Verify Code' }}</span>
        </button>
      </div>
    </form>

    <!-- Error Alert -->
    <div v-if="generalError" class="mt-4 p-3 bg-red-50 border border-red-200 rounded-md">
      <p class="text-sm text-red-600">{{ generalError }}</p>
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
      maxRetries: 3
    }
  },
  computed: {
    isValidPhone() {
      // Basic phone validation (at least 10 digits)
      return /^\d{10,15}$/.test(this.phone.replace(/\D/g, ''))
    },
    isValidCode() {
      // Basic code validation (5 digits)
      return /^\d{5}$/.test(this.code)
    }
  },
  methods: {
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

        const response = await fetch(`http://localhost:8080/api/telegram/auth/callback?phone=${encodeURIComponent('+' + this.phone)}`)
        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.error || 'Failed to start authentication')
        }

        const data = await response.json()
        this.hash = data.hash
        if (!this.hash) {
          throw new Error('Failed to get phone code hash')
        }

        this.showVerification = true
        this.startCountdown()
        this.retryCount = 0
      } catch (err) {
        this.phoneError = err.message
        this.generalError = 'Failed to send verification code. Please try again.'
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

        const response = await fetch('http://localhost:8080/api/telegram/auth/verify', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            phone: '+' + this.phone,
            code: this.code,
            hash: this.hash,
          }),
        })

        if (!response.ok) {
          const errorData = await response.json()
          
          if (errorData.error && errorData.error.includes('please wait')) {
            const waitTime = errorData.error.match(/\d+/)?.[0] || '30'
            throw new Error(`Please wait ${waitTime} seconds before trying again`)
          }

          if (errorData.error === '2FA_REQUIRED') {
            this.$emit('2fa-required')
            return
          }

          this.retryCount++
          if (this.retryCount >= this.maxRetries) {
            throw new Error('Too many failed attempts. Please try again with a new code.')
          }

          throw new Error(errorData.message || 'Verification failed')
        }

        const data = await response.json()
        if (data.success) {
          this.$emit('connected', data.user_id)
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
    }
  },
  beforeUnmount() {
    if (this.countdownInterval) {
      clearInterval(this.countdownInterval)
    }
  },
  emits: ['connected', '2fa-required']
}
</script> 