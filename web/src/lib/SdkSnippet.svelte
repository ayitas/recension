<script lang="ts">
	let {
		team,
		suite,
		apiKey = 'YOUR_API_KEY',
		apiURL = 'http://localhost:8080'
	}: {
		team: string;
		suite: string;
		apiKey?: string;
		apiURL?: string;
	} = $props();

	let copied = $state(false);

	const snippet = $derived(
		[
			`export RECENSION_API_KEY=${apiKey}`,
			`export RECENSION_API_URL=${apiURL}`,
			`export RECENSION_TEAM=${team}`,
			``,
			`cd examples/go/minimal`,
			`go run . -suite ${suite} -revision v1.0`
		].join('\n')
	);

	async function copy() {
		await navigator.clipboard.writeText(snippet);
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}
</script>

<section class="setup">
	<div class="head">
		<h3>Submit your first version</h3>
		<button type="button" onclick={copy}>{copied ? 'Copied' : 'Copy'}</button>
	</div>
	<p class="muted">
		Run this with the Go example (or your own workflow). The first revision becomes the baseline.
	</p>
	<pre>{snippet}</pre>
</section>

<style>
	.setup {
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		padding: 1.1rem 1.25rem;
		margin-top: 1rem;
	}
	.head {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
		margin-bottom: 0.4rem;
	}
	h3 {
		margin: 0;
		font-size: 1rem;
		letter-spacing: -0.02em;
	}
	button {
		border: 1px solid var(--accent);
		background: var(--accent-soft);
		color: var(--accent);
		padding: 0.4rem 0.7rem;
		cursor: pointer;
	}
	.muted {
		color: var(--muted);
		margin: 0 0 0.85rem;
		font-size: 0.92rem;
		line-height: 1.45;
	}
	pre {
		margin: 0;
		padding: 0.9rem 1rem;
		background: #1c1915;
		color: #f3efe6;
		overflow-x: auto;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.82rem;
		line-height: 1.5;
	}
</style>
