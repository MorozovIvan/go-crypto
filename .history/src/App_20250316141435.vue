<template>
  <div class="container">
    <button @click="showMessage" class="hello-button">Click me!</button>
    <p v-if="message" class="message">{{ message }}</p>
  </div>
</template>

<script>
export default {
  name: "App",
  data() {
    return {
      message: "",
    };
  },
  methods: {
    showMessage() {
      this.message = "Hello World!";
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
}

.hello-button {
  padding: 12px 24px;
  font-size: 18px;
  background-color: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.hello-button:hover {
  background-color: #45a049;
}

.message {
  margin-top: 20px;
  font-size: 24px;
  color: #333;
}
</style>
