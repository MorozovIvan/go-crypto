<template>
	<div>
		<!-- Simple Test Button -->
		<button 
			@click="toggleModal" 
			class="test-btn"
			style="background: red; color: white; padding: 10px; border: none; cursor: pointer;"
		>
			Test Button ({{ clickCount }}) - Modal: {{ showModal ? 'OPEN' : 'CLOSED' }}
		</button>

		<!-- Simple Modal -->
		<div v-if="showModal" style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 9999; display: flex; align-items: center; justify-content: center;">
			<div style="background: white; padding: 20px; border-radius: 8px;">
				<h2>Test Modal</h2>
				<p>This is a simple test modal to verify Vue reactivity works.</p>
				<button @click="closeModal" style="background: blue; color: white; padding: 5px 10px; border: none; cursor: pointer;">Close</button>
			</div>
		</div>
	</div>
</template>

<script>
export default {
	name: 'PortfolioPriceTrendsSimple',
	data() {
		return {
			showModal: false,
			clickCount: 0
		}
	},
	mounted() {
		console.log('PortfolioPriceTrendsSimple: mounted')
	},
	methods: {
		toggleModal() {
			this.clickCount++
			console.log('PortfolioPriceTrendsSimple: toggleModal called', this.clickCount)
			this.showModal = !this.showModal
			console.log('PortfolioPriceTrendsSimple: showModal is now', this.showModal)
		},
		closeModal() {
			console.log('PortfolioPriceTrendsSimple: closeModal called')
			this.showModal = false
		}
	}
}
</script>
</template>