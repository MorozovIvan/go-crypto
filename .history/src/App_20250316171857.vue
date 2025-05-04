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
      telegramGroups: []
    }
  },
  methods: {
    async fetchBalance() {
      this.loading = true
      this.error = null
      try {
        const response = await fetch('http://localhost:8080/balance')
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
      try {
        const response = await fetch('http://localhost:8080/telegram/auth-link')
        if (!response.ok) {
          throw new Error('Failed to get auth link')
        }
        const data = await response.json()
        window.location.href = data.auth_link
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    async checkTelegramStatus() {
      try {
        const response = await fetch('http://localhost:8080/telegram/status')
        if (!response.ok) {
          throw new Error('Failed to check Telegram status')
        }
        const data = await response.json()
        this.telegramConnected = data.connected
        if (data.connected) {
          this.telegramGroups = data.groups
        }
      } catch (err) {
        this.error = err.message
      }
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
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: #f5f5f5;
  padding: 20px;
}

.card {
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-bottom: 30px;
  width: 100%;
  max-width: 600px;
}

h2 {
  margin: 0 0 20px 0;
  color: #333;
  text-align: center;
}

.balance-button, .telegram-button {
  padding: 12px 24px;
  font-size: 18px;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-bottom: 20px;
  width: 100%;
}

.balance-button {
  background-color: #4caf50;
}

.balance-button:hover {
  background-color: #45a049;
}

.telegram-button {
  background-color: #0088cc;
}

.telegram-button:hover {
  background-color: #0077b3;
}

.loading {
  color: #666;
  font-size: 18px;
  margin: 20px 0;
  text-align: center;
}

.error {
  color: #f44336;
  font-size: 18px;
  margin: 20px 0;
  text-align: center;
}

.balance {
  text-align: center;
}

.amount {
  font-size: 36px;
  font-weight: bold;
  color: #4caf50;
  margin: 0;
}

.telegram-section {
  margin-top: 20px;
}

.telegram-status {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
  color: #4caf50;
  font-weight: bold;
}

.status-icon {
  margin-right: 8px;
  font-size: 20px;
}

.groups-list {
  display: grid;
  gap: 15px;
  margin-top: 20px;
}

.group-item {
  background-color: #f8f9fa;
  padding: 15px;
  border-radius: 4px;
  border: 1px solid #dee2e6;
}

.group-item h4 {
  margin: 0 0 5px 0;
  color: #333;
}

.group-item p {
  margin: 5px 0;
  color: #666;
}

.group-type {
  font-size: 14px;
  color: #888;
  text-transform: capitalize;
}

.no-groups {
  text-align: center;
  color: #666;
  font-style: italic;
}

.telegram-auth {
  text-align: center;
}

.telegram-auth p {
  margin-bottom: 15px;
  color: #666;
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
</style>
