<template>
  <main class="flex-1 p-6">
    <div v-if="!isConnected" class="text-center">
      <TelegramConnectForm
        @connected="handleConnected"
        @2fa-required="handle2FARequired"
      />
    </div>
    <div v-else>
      <div class="flex justify-between items-center mb-4">
        <h2 class="text-2xl font-bold">Make me reach</h2>
        <button
          @click="handleLogout"
          class="px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors"
        >
          Logout
        </button>
      </div>
      <p>Connected to Telegram. Analysis features coming soon...</p>
    </div>
  </main>
</template>

<script>
import TelegramConnectForm from './TelegramConnectForm.vue'

export default {
  name: 'MainContent',
  components: {
    TelegramConnectForm
  },
  props: {
    isConnected: {
      type: Boolean,
      default: false
    }
  },
  emits: ['connect', '2fa-required', 'logout'],
  methods: {
    handleConnected(userId) {
      this.$emit('connect', userId)
    },
    handle2FARequired(phone) {
      this.$emit('2fa-required', phone)
    },
    handleLogout() {
      this.$emit('logout')
    }
  }
}
</script> 