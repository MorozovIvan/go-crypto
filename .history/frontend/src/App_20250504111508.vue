<template>
  <div class="container">
    <div class="card">
      <h2>Wallet Balance</h2>
      <div v-if="balance !== null" class="balance">
        {{ balance }} ETH
      </div>
      <div v-else-if="loading" class="loading">
        Loading...
      </div>
      <div v-else class="error">
        {{ error }}
      </div>
    </div>

    <div class="card">
      <h2>Telegram Integration</h2>
      <div v-if="telegramConnected" class="telegram-connected">
        <div class="status">
          <span class="icon">✓</span>
          Connected as @{{ telegramUsername }}
        </div>
        <div v-if="telegramGroups.length > 0" class="groups">
          <h3>Your Groups</h3>
          <div class="groups-grid">
            <div v-for="group in telegramGroups" :key="group.id" class="group-card">
              <h4>{{ group.title }}</h4>
              <div class="group-info">
                <span class="type" :class="group.type">{{ group.type }}</span>
                <span v-if="group.username" class="username">@{{ group.username }}</span>
              </div>
              <div v-if="group.members" class="members">
                {{ group.members }} members
              </div>
              <p v-if="group.description" class="description">{{ group.description }}</p>
            </div>
          </div>
        </div>
        <div v-else class="no-groups">
          No groups found. Make sure you're a member of some groups.
        </div>
      </div>
      <div v-else>
        <button @click="connectTelegram" :disabled="loading" class="connect-button">
          <span v-if="!loading">Connect Telegram</span>
          <span v-else>Connecting...</span>
        </button>
        <div v-if="error" class="error">
          {{ error }}
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      balance: null,
      loading: false,
      error: null,
      telegramConnected: false,
      telegramGroups: [],
      telegramUsername: null,
      userId: null
    }
  },
  methods: {
    async fetchBalance() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('http://localhost:8080/api/balance')
        if (!response.ok) {
          throw new Error('Failed to fetch balance')
        }
        const data = await response.json()
        this.balance = data.balance
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    async connectTelegram() {
      this.loading = true
      this.error = null
      this.telegramConnected = false
      this.telegramGroups = []
      this.telegramUsername = null
      this.userId = null
      
      try {
        // Prompt for phone number
        const phone = prompt('Please enter your Telegram phone number (with country code, e.g., +1234567890):')
        if (!phone) {
          throw new Error('Phone number is required')
        }

        // Start authentication process
        const response = await fetch(`http://localhost:8080/api/telegram/auth/callback?phone=${encodeURIComponent(phone)}`)
        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.error || 'Failed to start authentication')
        }

        // Get the phone code hash from the response
        const authData = await response.json()
        const hash = authData.hash
        if (!hash) {
          throw new Error('Failed to get phone code hash')
        }

        // Show a message that the code has been sent
        alert('A verification code has been sent to your Telegram. Please check your Telegram app.')

        // Prompt for verification code
        const code = prompt('Please enter the verification code sent to your Telegram:')
        if (!code) {
          throw new Error('Verification code is required')
        }

        // Verify the code
        const verifyResponse = await fetch('http://localhost:8080/api/telegram/auth/verify', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            phone: phone,
            code: code,
            hash: hash,
          }),
        })

        if (!verifyResponse.ok) {
          const errorData = await verifyResponse.json()
          // Check if 2FA is required
          if (errorData.error && errorData.error.includes('2FA password is required')) {
            // Prompt for 2FA password
            const password = prompt('Please enter your 2FA password:')
            if (!password) {
              throw new Error('2FA password is required')
            }

            // Try to verify with 2FA password
            const verify2FAResponse = await fetch('http://localhost:8080/api/telegram/auth/verify', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
              },
              body: JSON.stringify({
                phone: phone,
                code: code,
                hash: hash,
                password: password,
              }),
            })

            if (!verify2FAResponse.ok) {
              const errorData = await verify2FAResponse.json()
              throw new Error(errorData.error || 'Failed to verify with 2FA password')
            }
          } else {
            throw new Error(errorData.error || 'Failed to verify code')
          }
        }

        // Show success message
        alert('Successfully authenticated with Telegram!')

        // Fetch groups after successful authentication
        const groupsResponse = await fetch('http://localhost:8080/api/telegram/groups')
        if (!groupsResponse.ok) {
          throw new Error('Failed to fetch groups')
        }
        const groupsData = await groupsResponse.json()
        this.telegramConnected = groupsData.connected
        this.telegramGroups = groupsData.groups || []
      } catch (err) {
        console.error('Telegram connection error:', err)
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    async pollTelegramStatus(authWindow) {
      const maxAttempts = 30 // 30 seconds timeout
      let attempts = 0
      
      const poll = async () => {
        if (attempts >= maxAttempts) {
          this.error = 'Connection timeout. Please try again.'
          if (authWindow) {
            authWindow.close()
          }
          return
        }
        
        try {
          const response = await fetch(`http://localhost:8080/api/telegram/groups${this.userId ? `?user_id=${this.userId}` : ''}`)
          if (response.ok) {
            const data = await response.json()
            if (data.connected && data.groups) {
              this.telegramConnected = true
              this.telegramGroups = data.groups
              if (authWindow) {
                authWindow.close()
              }
              return // Stop polling on success
            }
          }
        } catch (err) {
          console.error('Polling error:', err)
        }
        
        attempts++
        setTimeout(poll, 1000) // Poll every second
      }
      
      poll()
    }
  },
  mounted() {
    this.fetchBalance()
  }
}
</script>

<style>
.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.card {
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

h2 {
  margin-top: 0;
  color: #333;
}

.balance {
  font-size: 24px;
  font-weight: bold;
  color: #4CAF50;
}

.loading {
  color: #666;
}

.error {
  color: #f44336;
}

.telegram-connected {
  margin-top: 1rem;
}

.status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #4CAF50;
  margin-bottom: 1rem;
}

.icon {
  font-weight: bold;
}

.groups {
  margin-top: 1rem;
}

.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.group-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 1rem;
  border: 1px solid #e9ecef;
}

.group-card h4 {
  margin: 0 0 0.5rem 0;
  color: #333;
}

.group-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.type {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.875rem;
  font-weight: 500;
}

.type.channel {
  background: #e3f2fd;
  color: #1976d2;
}

.type.group {
  background: #e8f5e9;
  color: #2e7d32;
}

.type.supergroup {
  background: #fff3e0;
  color: #f57c00;
}

.username {
  color: #666;
  font-size: 0.875rem;
}

.members {
  color: #666;
  font-size: 0.875rem;
  margin-bottom: 0.5rem;
}

.description {
  font-size: 0.875rem;
  color: #666;
  margin: 0.5rem 0 0 0;
  line-height: 1.4;
}

.no-groups {
  color: #666;
  font-style: italic;
  text-align: center;
  padding: 2rem;
}

.connect-button {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 12px 24px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 200px;
  transition: background-color 0.2s;
}

.connect-button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.connect-button:hover:not(:disabled) {
  background-color: #45a049;
}
</style> 