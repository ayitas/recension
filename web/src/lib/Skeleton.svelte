<script lang="ts">
	let {
		lines = 4,
		title = true
	}: {
		lines?: number;
		title?: boolean;
	} = $props();

	const widths = ['92%', '78%', '86%', '64%', '71%', '88%'];
</script>

<div class="wrap" aria-busy="true" aria-live="polite">
	<span class="sr-only">Loading…</span>
	{#if title}<div class="skeleton title"></div>{/if}
	{#each Array.from({ length: lines }, (_, i) => i) as i}
		<div class="skeleton line" style={`width: ${widths[i % widths.length]}`}></div>
	{/each}
	<div class="skeleton block"></div>
</div>

<style>
	.wrap {
		margin: 0.5rem 0 1rem;
	}
</style>
