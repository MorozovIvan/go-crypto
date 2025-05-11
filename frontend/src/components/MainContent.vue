<template>
  <main class="flex-1 p-6">
    <div v-if="!isConnected" class="text-center">
      <TelegramConnectForm
        @connected="handleConnected"
        @2fa-required="handle2FARequired"
      />
    </div>
    <div v-else>
      <h2 class="text-2xl font-bold mb-4">Make me reach</h2>
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
  emits: ['connect-telegram', '2fa-required'],
  methods: {
    handleConnected(userId) {
      this.$emit('connect-telegram', userId)
    },
    handle2FARequired() {
      this.$emit('2fa-required')
    }
  }
}
</script> 