import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiUrl = env.API_URL || 'http://192.168.100.7:8080'

  return {
    plugins: [vue()],
    server: {
      proxy: {
        '/api': {
          target: apiUrl,
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path
        },
        '/telegram': {
          target: apiUrl,
          changeOrigin: true,
          secure: false,
          rewrite: (path) => path
        }
      }
    }
  }
}) 