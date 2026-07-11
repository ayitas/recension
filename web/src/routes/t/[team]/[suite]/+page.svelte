<script lang="ts">
	import { listBatches, me, type Batch } from '$lib/api';
	import SdkSnippet from '$lib/SdkSnippet.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '$lib/auth';

	let batches = $state<Batch[]>([]);
	let error = $state('');
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

	$effect(() => {
		if (!isLoggedIn()) {
			goto('/login');
			return;
		}
		if (!team || !suite) return;
		listBatches(team, suite)
			.then((data) => {
				batches = [...data].sort(
					(a, b) => new Date(b.submittedAt).getTime() - new Date(a.submittedAt).getTime()
				);
				error = '';
			})
			.catch((err: Error) => {
				error = err.message;
			});
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
<h1>{suite}</h1>

{#if error}
	<p class="error">{error}</p>
{:else if batches.length === 0}
	<p class="muted">No batches yet. Submit your first revision to establish a baseline.</p>
	<SdkSnippet {team} {suite} {apiKey} />
{:else}
	<ul>
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
				<span class="muted">
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
	.crumb {
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	h1 {
		margin: 0 0 1.5rem;
		letter-spacing: -0.04em;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
	}
	li {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-top: 1px solid var(--line);
		align-items: baseline;
	}
	li:first-child {
		border-top: 0;
	}
	.main {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: 0.55rem 0.85rem;
		min-width: 0;
	}
	.baseline {
		font-size: 0.7rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--sent);
	}
	.scores {
		color: var(--muted);
		font-size: 0.88rem;
	}
	.scores.has-diff {
		color: var(--diff);
	}
	.muted {
		color: var(--muted);
		flex-shrink: 0;
	}
	.error {
		color: var(--diff);
	}
	.more {
		margin-top: 1.25rem;
	}
	.more summary {
		cursor: pointer;
		color: var(--accent);
		margin-bottom: 0.5rem;
	}
</style>
