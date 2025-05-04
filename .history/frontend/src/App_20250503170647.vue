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
          Connected to Telegram
        </div>
        <div v-if="telegramGroups.length > 0" class="groups">
          <h3>Your Groups</h3>
          <ul>
            <li v-for="group in telegramGroups" :key="group.id">
              {{ group.title }} ({{ group.type }})
              <span v-if="group.username">@{{ group.username }}</span>
              <p class="description">{{ group.description }}</p>
            </li>
          </ul>
        </div>
        <div v-else class="no-groups">
          No groups found. Make sure you're a member of some groups.
        </div>
      </div>
      <div v-else>
        <button @click="connectTelegram" :disabled="loading">
          Connect Telegram
        </button>
        <div v-if="loading" class="loading">
          Connecting...
        </div>
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
      this.userId = null
      
      try {
        const response = await fetch('http://localhost:8080/api/telegram/auth')
        if (!response.ok) {
          const errorData = await response.json()
          throw new Error(errorData.error || 'Failed to get auth link')
        }
        const data = await response.json()
        if (!data.auth_link) {
          throw new Error('Invalid auth link response')
        }
        // Open in a new window
        window.open(data.auth_link, '_blank', 'width=600,height=600')
        
        // Start polling for connection status
        this.pollTelegramStatus()
      } catch (err) {
        console.error('Telegram connection error:', err)
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    async checkTelegramStatus() {
      try {
        const url = this.userId 
          ? `http://localhost:8080/api/telegram/groups?user_id=${this.userId}`
          : 'http://localhost:8080/api/telegram/groups'
          
        const response = await fetch(url)
        if (!response.ok) {
          throw new Error('Failed to check Telegram status')
        }
        const data = await response.json()
        
        // Update connection state based on response
        this.telegramConnected = data.connected || false
        this.telegramGroups = data.groups || []
        
        // If not connected, clear user data
        if (!data.connected) {
          this.userId = null
        }
      } catch (err) {
        this.error = err.message
        this.telegramConnected = false
        this.telegramGroups = []
        this.userId = null
      }
    },
    async pollTelegramStatus() {
      const maxAttempts = 30 // 30 seconds timeout
      let attempts = 0
      
      const poll = async () => {
        if (attempts >= maxAttempts) {
          this.error = 'Connection timeout. Please try again.'
          return
        }
        
        try {
          // First check if we have auth data in the URL
          const urlParams = new URLSearchParams(window.location.search)
          const authData = {
            id: urlParams.get('id'),
            first_name: urlParams.get('first_name'),
            last_name: urlParams.get('last_name'),
            username: urlParams.get('username'),
            photo_url: urlParams.get('photo_url'),
            auth_date: urlParams.get('auth_date'),
            hash: urlParams.get('hash')
          }
          
          // If we have auth data, send it to the backend
          if (authData.id && authData.hash) {
            const response = await fetch('http://localhost:8080/api/telegram/auth/callback', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify(authData)
            })
            
            if (response.ok) {
              const data = await response.json()
              if (data.user && data.user.id) {
                this.userId = data.user.id
                this.telegramConnected = true
                this.telegramGroups = data.groups || []
                // Clear URL parameters
                window.history.replaceState({}, document.title, window.location.pathname)
                return // Stop polling on success
              }
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
    this.checkTelegramStatus()
  }
}
</script>

<style>
.container {
  max-width: 800px;
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
}

.icon {
  font-weight: bold;
}

.groups {
  margin-top: 1rem;
}

.groups ul {
  list-style: none;
  padding: 0;
}

.groups li {
  padding: 0.5rem;
  border-bottom: 1px solid #eee;
}

.groups li:last-child {
  border-bottom: none;
}

.no-groups {
  color: #666;
  font-style: italic;
}

.description {
  font-size: 14px;
  color: #666;
  margin: 5px 0 0 0;
}

button {
  background-color: #4CAF50;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
}

button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

button:hover:not(:disabled) {
  background-color: #45a049;
}
</style> 