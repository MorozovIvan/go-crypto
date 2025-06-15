<template>
  <div class="max-w-md p-6 mx-auto bg-white rounded-lg shadow-md">
    <h2 class="mb-6 text-2xl font-bold text-center">Connect to Telegram</h2>
    
    <!-- Phone Number Form -->
    <form v-if="!showVerification" @submit.prevent="handlePhoneSubmit" class="space-y-4">
      <div>
        <label for="phone" class="block mb-1 text-sm font-medium text-gray-700">Phone Number</label>
        <div class="relative">
          <span class="absolute inset-y-0 left-0 flex items-center pl-3 text-gray-500">+</span>
          <input
            type="tel"
            id="phone"
            v-model="phone"
            placeholder="1234567890"
            class="pl-8 w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-[#0088cc] focus:border-transparent"
            :class="{ 'border-red-500': phoneError }"
            @input="formatPhoneNumber"
            maxlength="15"
            :disabled="loading || isFlooded"
          />
        </div>
        <p v-if="phoneError" class="mt-1 text-sm text-red-600">{{ phoneError }}</p>
        <p class="mt-1 text-sm text-gray-500">Enter your phone number with country code (e.g., 1234567890)</p>
      </div>
      <button
        type="submit"
        class="w-full bg-[#0088cc] text-white px-4 py-2 rounded-lg hover:bg-[#0077b3] transition-colors flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed"
        :disabled="loading || !isValidPhone || isFlooded"
      >
        <svg v-if="loading" class="w-5 h-5 text-white animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span v-if="isFlooded">Please wait {{ formatWaitTime(floodWaitTime) }}</span>
        <span v-else>{{ loading ? 'Sending Code...' : 'Send Verification Code' }}</span>
      </button>
    </form>

    <!-- Verification Code Form -->
    <form v-else @submit.prevent="handleVerificationSubmit" class="space-y-4">
      <div>
        <label for="code" class="block mb-1 text-sm font-medium text-gray-700">Verification Code</label>
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
            :disabled="loading || isFlooded"
          />
          <div v-if="countdown > 0" class="absolute text-sm text-gray-500 transform -translate-y-1/2 right-3 top-1/2">
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
          class="flex-1 px-4 py-2 text-gray-700 transition-colors border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading || isFlooded"
        >
          Back
        </button>
        <button
          type="submit"
          class="flex-1 bg-[#0088cc] text-white px-4 py-2 rounded-lg hover:bg-[#0077b3] transition-colors flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed"
          :disabled="loading || !isValidCode || isFlooded"
        >
          <svg v-if="loading" class="w-5 h-5 text-white animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span v-if="isFlooded">Please wait {{ formatWaitTime(floodWaitTime) }}</span>
          <span v-else>{{ loading ? 'Verifying...' : 'Verify Code' }}</span>
        </button>
      </div>
    </form>

    <!-- Error Alert -->
    <div v-if="generalError" class="p-3 mt-4 border border-red-200 rounded-md bg-red-50">
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
      maxRetries: 3,
      isFlooded: false,
      floodWaitTime: 0,
      floodInterval: null
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
    }
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
        const response = await fetch(`http://localhost:8080/api/telegram/auth/callback?phone=${encodeURIComponent('+' + this.phone)}`)
        console.log('Response status:', response.status)
        const data = await response.json()
        console.log('Response data:', data)
        
        if (!response.ok) {
          console.error('Response not OK:', data)
          if (data.error && (data.error.includes('FLOOD_WAIT') || data.error.includes('code 420'))) {
            const waitTime = parseInt(data.error.match(/\((\d+)\)/)?.[1] || '30')
            this.startFloodTimer(waitTime)
            throw new Error(`Too many attempts. Please wait ${this.formatWaitTime(waitTime)} before trying again.`)
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
          this.generalError = 'Failed to send verification code. Please try again.'
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

        const response = await fetch('http://localhost:8080/api/telegram/verify-code', {
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
          
          if (errorData.error && errorData.error.includes('PHONE_PASSWORD_FLOOD')) {
            const waitTime = parseInt(errorData.error.match(/\d+/)?.[0] || '30')
            this.startFloodTimer(waitTime)
            throw new Error(`Too many attempts. Please wait ${this.formatWaitTime(waitTime)} before trying again.`)
          }

          if (errorData.error === '2FA_REQUIRED') {
            this.$emit('2fa-required', '+' + this.phone)
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
      this.resetFloodState()
    }
  },
  beforeDestroy() {
    if (this.countdownInterval) {
      clearInterval(this.countdownInterval)
    }
    this.resetFloodState()
  },
  emits: ['connected', '2fa-required']
}
</script> 