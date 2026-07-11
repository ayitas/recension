<script lang="ts">
	import { getBatch, promoteBatch, type BatchDetail } from '$lib/api';
	import { page } from '$app/stores';

	let detail = $state<BatchDetail | null>(null);
	let error = $state('');
	let promoting = $state(false);

	const team = $derived($page.params.team ?? '');
	const suite = $derived($page.params.suite ?? '');
	const batch = $derived($page.params.batch ?? '');

	async function load() {
		if (!team || !suite || !batch) return;
		try {
			detail = await getBatch(team, suite, batch);
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		}
	}

	$effect(() => {
		void load();
	});

	async function onPromote() {
		promoting = true;
		try {
			await promoteBatch(team, suite, batch);
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			promoting = false;
		}
	}
</script>

<p class="crumb">
	<a href="/">Teams</a> /
	<a href={`/t/${team}`}>{team}</a> /
	<a href={`/t/${team}/${suite}`}>{suite}</a> /
	{batch}
</p>

<div class="title-row">
	<h1>{batch}</h1>
	{#if detail}
		<button onclick={onPromote} disabled={promoting || detail.baselineBatchId === detail.batch.id}>
			{detail.baselineBatchId === detail.batch.id ? 'Current baseline' : 'Promote baseline'}
		</button>
	{/if}
</div>

{#if error}
	<p class="error">{error}</p>
{/if}

{#if detail}
	<p class="meta muted">
		{detail.elements.length} testcases
		{#if detail.baselineBatchId === detail.batch.id}
			· this batch is the baseline
		{/if}
	</p>

	<table class="data-table">
		<thead>
			<tr>
				<th>Testcase</th>
				<th>Verdict</th>
				<th>Score</th>
			</tr>
		</thead>
		<tbody>
			{#each detail.elements as el}
				<tr>
					<td>
						<a href={`/t/${team}/${suite}/${batch}/e/${encodeURIComponent(el.testcase)}`}>
							{el.testcase}
						</a>
					</td>
					<td><span class={`verdict ${el.verdict}`}>{el.verdict}</span></td>
					<td class="score">{(el.score ?? 0).toFixed(3)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 0.75rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.04em;
		font-size: clamp(1.9rem, 4vw, 2.6rem);
	}

	.meta {
		margin: 0 0 1.25rem;
	}

	.score {
		font-variant-numeric: tabular-nums;
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.9rem;
	}
</style>
