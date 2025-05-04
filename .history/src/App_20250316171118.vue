<template>
  <div class="container">
    <div class="section">
      <h2>Solana Wallet Balance</h2>
      <button @click="fetchBalance" class="balance-button">Get Wallet Balance</button>
      <div v-if="loading" class="loading">Loading...</div>
      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="balance !== null" class="balance">
        <p class="amount">{{ balance }} SOL</p>
      </div>
    </div>

    <div class="section">
      <h2>Telegram Integration</h2>
      <div v-if="!telegramConnected" class="telegram-auth">
        <p>Connect your Telegram account to view your channels</p>
        <button @click="connectTelegram" class="telegram-button">
          Connect Telegram
        </button>
      </div>
      
      <div v-if="telegramLoading" class="loading">Loading...</div>
      <div v-if="telegramError" class="error">{{ telegramError }}</div>
      
      <div v-if="telegramConnected" class="telegram-section">
        <div class="telegram-status">
          <span class="status-icon">✓</span>
          <span>Connected to Telegram</span>
        </div>
        
        <h3>Your Telegram Channels</h3>
        <div v-if="groups.length === 0" class="no-groups">
          No channels found. Make sure you've started the bot with /start command.
        </div>
        <div v-else class="groups-list">
          <div v-for="group in groups" :key="group.id" class="group-item">
            <h4>{{ group.title }}</h4>
            <p v-if="group.username">@{{ group.username }}</p>
            <p class="group-type">{{ group.type }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "App",
  data() {
    return {
      balance: null,
      loading: false,
      error: null,
      telegramConnected: false,
      telegramLoading: false,
      telegramError: null,
      groups: [],
    };
  },
  methods: {
    async fetchBalance() {
      this.loading = true;
      this.error = null;
      this.balance = null;

      try {
        const response = await fetch('http://localhost:8080/api/balance');
        const data = await response.json();
        
        if (data.error) {
          this.error = data.error;
        } else {
          this.balance = data.balance;
        }
      } catch (err) {
        this.error = 'Failed to fetch balance. Please try again.';
      } finally {
        this.loading = false;
      }
    },
    async connectTelegram() {
      this.telegramLoading = true;
      this.telegramError = null;

      try {
        const response = await fetch('http://localhost:8080/api/telegram/auth');
        const data = await response.json();
        
        if (data.error) {
          this.telegramError = data.error;
        } else {
          window.open(data.auth_link, '_blank');
          this.telegramConnected = true;
          // Wait a bit for the user to authorize
          setTimeout(() => this.fetchGroups(), 2000);
        }
      } catch (err) {
        this.telegramError = 'Failed to connect Telegram. Please try again.';
      } finally {
        this.telegramLoading = false;
      }
    },
    async fetchGroups() {
      this.telegramLoading = true;
      this.telegramError = null;

      try {
        const response = await fetch('http://localhost:8080/api/telegram/groups');
        const data = await response.json();
        
        if (data.error) {
          this.telegramError = data.error;
          if (data.error.includes('not authorized')) {
            this.telegramConnected = false;
          }
        } else {
          this.groups = data.groups;
        }
      } catch (err) {
        this.telegramError = 'Failed to fetch groups. Please try again.';
      } finally {
        this.telegramLoading = false;
      }
    },
  },
};
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

.section {
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
</style>
