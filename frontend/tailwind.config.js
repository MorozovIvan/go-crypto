/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        telegram: {
          blue: '#0088cc',
          light: '#e6f3f8',
          dark: '#005580',
          gray: '#f5f5f5',
          'gray-dark': '#e0e0e0',
          success: '#4CAF50',
          warning: '#FFA000',
          error: '#F44336',
        }
      },
      spacing: {
        '72': '18rem',
        '84': '21rem',
        '96': '24rem',
      },
      maxWidth: {
        '8xl': '88rem',
        '9xl': '96rem',
      },
      boxShadow: {
        'telegram': '0 2px 4px rgba(0, 136, 204, 0.1)',
        'telegram-hover': '0 4px 6px rgba(0, 136, 204, 0.15)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
  ],
} 