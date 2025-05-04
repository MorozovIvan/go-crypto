<template>
  <div class="container">
    <button @click="fetchBalance" class="balance-button">
      Get Wallet Balance
    </button>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="balance !== null" class="balance">
      <h2>Wallet Balance</h2>
      <p class="amount">{{ balance }} SOL</p>
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
    };
  },
  methods: {
    async fetchBalance() {
      this.loading = true;
      this.error = null;
      this.balance = null;

      try {
        const response = await fetch("http://localhost:8080/api/balance");
        const data = await response.json();

        if (data.error) {
          this.error = data.error;
        } else {
          this.balance = data.balance;
        }
      } catch (err) {
        this.error = "Failed to fetch balance. Please try again.";
      } finally {
        this.loading = false;
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

.balance-button {
  padding: 12px 24px;
  font-size: 18px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
  margin-bottom: 20px;
}

.balance-button:hover {
  background-color: #45a049;
}

.loading {
  color: #666;
  font-size: 18px;
  margin: 20px 0;
}

.error {
  color: #f44336;
  font-size: 18px;
  margin: 20px 0;
  text-align: center;
}

.balance {
  text-align: center;
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  margin-top: 20px;
}

.balance h2 {
  margin: 0 0 15px 0;
  color: #333;
}

.amount {
  font-size: 36px;
  font-weight: bold;
  color: #4caf50;
  margin: 0;
}
</style>
