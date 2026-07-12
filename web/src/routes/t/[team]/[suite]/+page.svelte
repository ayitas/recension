<script lang="ts">
	import { listBatches, me, type Batch } from '$lib/api';
	import SdkSnippet from '$lib/SdkSnippet.svelte';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '$lib/auth';

	let batches = $state<Batch[]>([]);
	let error = $state('');
	let loading = $state(true);
	let apiKey = $state('YOUR_API_KEY');

	const team = $derived($page.params.team ?? '');
	const suite = $derived($page.params.suite ?? '');

	function scoreLine(batch: Batch): string {
		const n = batch.testcaseCount ?? 0;
		if (n === 0) return 'no testcases';
		const parts: string[] = [`${n} cases`];
		if (batch.isBaseline || (batch.sentCount ?? 0) > 0) {
			parts.push(`${batch.sentCount ?? 0} sent`);
		}
		if ((batch.diffCount ?? 0) > 0) {
			parts.push(`${batch.diffCount} diff`);
		}
		if ((batch.passCount ?? 0) > 0 && !batch.isBaseline) {
			parts.push(`${batch.passCount} pass`);
		}
		if (batch.avgScore !== undefined) {
			parts.push(`score ${batch.avgScore.toFixed(3)}`);
		}
		return parts.join(' · ');
	}

	async function load() {
		if (!team || !suite) return;
		loading = true;
		try {
			const data = await listBatches(team, suite);
			batches = [...data].sort(
				(a, b) => new Date(b.submittedAt).getTime() - new Date(a.submittedAt).getTime()
			);
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (!isLoggedIn()) {
			goto('/login');
			return;
		}
		void load();
		me()
			.then((u) => {
				apiKey = u.apiKey;
			})
			.catch(() => {});
	});
</script>

<p class="crumb">
	<a href="/">Teams</a> / <a href={`/t/${team}`}>{team}</a> / {suite}
</p>
<h1 class="brand-font">{suite}</h1>

{#if error}
	<ErrorBanner message={error} onretry={load} />
{:else if loading}
	<Skeleton lines={4} />
{:else if batches.length === 0}
	<p class="muted lead">No batches yet. Submit your first revision to establish a baseline.</p>
	<SdkSnippet {team} {suite} {apiKey} />
{:else}
	<ul class="list-card">
		{#each batches as batch}
			<li>
				<div class="main">
					<a href={`/t/${team}/${suite}/${batch.slug}`}>{batch.slug}</a>
					{#if batch.isBaseline}
						<span class="baseline">baseline</span>
					{/if}
					<span class="scores" class:has-diff={(batch.diffCount ?? 0) > 0}>
						{scoreLine(batch)}
					</span>
				</div>
				<span class="muted when">
					{new Date(batch.submittedAt).toLocaleString()}
					{#if batch.sealedAt}· sealed{/if}
				</span>
			</li>
		{/each}
	</ul>
	<details class="more">
		<summary>SDK commands</summary>
		<SdkSnippet {team} {suite} {apiKey} />
	</details>
{/if}

<style>
	h1 {
		margin: 0 0 1.15rem;
		letter-spacing: -0.035em;
		font-size: clamp(1.65rem, 3.5vw, 2.2rem);
	}

	.lead {
		margin: 0 0 0.5rem;
	}

	.main {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.45rem 0.75rem;
		min-width: 0;
	}

	.baseline {
		font-size: 0.66rem;
		font-weight: 650;
		text-transform: uppercase;
		letter-spacing: 0.07em;
		color: var(--sent);
		border: 1px solid color-mix(in srgb, var(--sent) 30%, transparent);
		border-radius: var(--radius);
		padding: 0.1rem 0.4rem;
	}

	.scores {
		color: var(--muted);
		font-size: 0.84rem;
		font-variant-numeric: tabular-nums;
	}

	.scores.has-diff {
		color: var(--diff);
	}

	.when {
		flex-shrink: 0;
		font-size: 0.84rem;
	}

	.more {
		margin-top: 1.15rem;
	}

	.more summary {
		cursor: pointer;
		color: var(--accent);
		margin-bottom: 0.55rem;
		list-style: none;
		font-weight: 550;
	}

	.more summary::-webkit-details-marker {
		display: none;
	}

	.more summary::before {
		content: '▸ ';
		display: inline-block;
		transition: transform 160ms var(--ease);
	}

	.more[open] summary::before {
		transform: rotate(90deg);
	}
</style>
