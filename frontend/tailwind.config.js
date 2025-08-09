/** @type {import('tailwindcss').Config} */
export default {
	content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
	theme: {
		extend: {
			fontFamily: {
				inter: ['Inter', 'sans-serif'],
			},
		},
	},
	plugins: [require('daisyui')],
	daisyui: {
		themes: [
			{
				crypto: {
					primary: '#6366f1',
					secondary: '#8b5cf6',
					accent: '#06b6d4',
					neutral: '#1f2937',
					'base-100': '#0f172a',
					'base-200': '#1e293b',
					'base-300': '#334155',
					'base-content': '#f1f5f9',
					info: '#3abff8',
					success: '#10b981',
					warning: '#f59e0b',
					error: '#ef4444',
				},
			},
			'light',
			'dark',
		],
		darkTheme: 'crypto',
	},
}
