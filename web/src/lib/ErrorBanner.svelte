<script lang="ts">
	let {
		message = '',
		onretry
	}: {
		message: string;
		onretry?: () => void;
	} = $props();
</script>

{#if message}
	<div class="banner" role="alert" aria-live="assertive">
		<p class="error">{message}</p>
		{#if onretry}
			<button type="button" class="ghost" onclick={onretry}>Retry</button>
		{/if}
	</div>
{/if}

<style>
	.banner {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.65rem 1rem;
		padding: 0.85rem 1.1rem;
		margin: 0 0 1rem;
		border: 1px solid color-mix(in srgb, var(--diff) 35%, var(--hairline));
		background: color-mix(in srgb, var(--diff) 8%, var(--card));
		border-radius: var(--radius-md);
	}

	p {
		margin: 0;
		flex: 1;
		min-width: 12rem;
	}
</style>
