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
		<div>
			<span class="tab">shell</span>
			<h3>Submit your first version</h3>
		</div>
		<button type="button" class="ghost" onclick={copy}>{copied ? 'Copied' : 'Copy'}</button>
	</div>
	<p class="muted">
		Run this with the Go example (or your own workflow). The first revision becomes the baseline.
	</p>
	<pre class="code-well">{snippet}</pre>
</section>

<style>
	.setup {
		background: var(--bone);
		border: 1px solid var(--hairline);
		border-radius: var(--radius-md);
		padding: 1.5rem;
		margin-top: 1rem;
	}

	.head {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 1rem;
		margin-bottom: 0.45rem;
	}

	.tab {
		display: inline-block;
		font-family: var(--font-mono);
		font-size: 0.6875rem;
		color: var(--on-dark-mute);
		background: var(--surface-deep);
		padding: 0.375rem 0.75rem;
		border-radius: var(--radius-xs);
		margin-bottom: 0.55rem;
	}

	h3 {
		margin: 0;
		font-family: var(--font-brand);
		font-size: 1.15rem;
		letter-spacing: -0.03em;
		font-weight: 600;
	}

	.muted {
		color: var(--muted);
		margin: 0 0 1rem;
		font-size: 0.9375rem;
		line-height: 1.5;
	}

	pre {
		margin: 0;
	}
</style>
