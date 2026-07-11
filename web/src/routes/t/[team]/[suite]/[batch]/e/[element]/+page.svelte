<script lang="ts">
	import { getElement, type ElementDetail } from '$lib/api';
	import DiffTable from '$lib/DiffTable.svelte';
	import { page } from '$app/stores';

	let detail = $state<ElementDetail | null>(null);
	let error = $state('');
	let changedOnly = $state(true);
	let blobsOnly = $state(false);

	const team = $derived($page.params.team ?? '');
	const suite = $derived($page.params.suite ?? '');
	const batch = $derived($page.params.batch ?? '');
	const element = $derived(decodeURIComponent($page.params.element ?? ''));

	$effect(() => {
		if (!team || !suite || !batch || !element) return;
		getElement(team, suite, batch, element)
			.then((data) => {
				detail = data;
				error = '';
			})
			.catch((err: Error) => {
				error = err.message;
			});
	});
</script>

<p class="crumb">
	<a href="/">Teams</a> /
	<a href={`/t/${team}`}>{team}</a> /
	<a href={`/t/${team}/${suite}`}>{suite}</a> /
	<a href={`/t/${team}/${suite}/${batch}`}>{batch}</a> /
	{element}
</p>

{#if error}
	<p class="error">{error}</p>
{:else if detail}
	<div class="title-row">
		<h1>{detail.testcase}</h1>
		<span class={`verdict ${detail.verdict}`}>{detail.verdict}</span>
	</div>

	<p class="meta">
		score {(detail.score ?? 0).toFixed(3)}
		· comparing
		<strong>{detail.comparison.src.version}</strong>
		vs baseline
		<strong>{detail.comparison.dst.version}</strong>
	</p>

	<div class="stats">
		<div>
			<span class="n">{detail.comparison.overview.keysCountCommon}</span>
			<span class="l">common</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.keysCountFresh}</span>
			<span class="l">fresh</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.keysCountMissing}</span>
			<span class="l">missing</span>
		</div>
	</div>

	<div class="filters">
		<label class="filter">
			<input type="checkbox" bind:checked={changedOnly} />
			Changed only
		</label>
		<label class="filter">
			<input type="checkbox" bind:checked={blobsOnly} />
			Blobs only
		</label>
	</div>

	<DiffTable
		title="Checks"
		cellar={detail.comparison.results}
		srcLabel={detail.comparison.src.version}
		dstLabel={detail.comparison.dst.version}
		{changedOnly}
		{blobsOnly}
	/>
	<DiffTable
		title="Assumptions"
		cellar={detail.comparison.asserts}
		srcLabel={detail.comparison.src.version}
		dstLabel={detail.comparison.dst.version}
		{changedOnly}
		{blobsOnly}
	/>
	<DiffTable
		title="Metrics"
		cellar={detail.comparison.metrics}
		srcLabel={detail.comparison.src.version}
		dstLabel={detail.comparison.dst.version}
		{changedOnly}
		{blobsOnly}
	/>
{/if}

<style>
	.crumb {
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: baseline;
		margin-bottom: 0.5rem;
	}
	h1 {
		margin: 0;
		letter-spacing: -0.04em;
	}
	.verdict {
		text-transform: uppercase;
		font-size: 0.8rem;
		letter-spacing: 0.1em;
	}
	.verdict.pass {
		color: var(--pass);
	}
	.verdict.diff {
		color: var(--diff);
	}
	.verdict.sent {
		color: var(--sent);
	}
	.meta {
		color: var(--muted);
		margin: 0 0 1.25rem;
	}
	.stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 8rem));
		gap: 0.75rem;
		margin-bottom: 1.75rem;
	}
	.stats div {
		background: var(--card);
		border: 1px solid var(--line);
		padding: 0.75rem 0.9rem;
	}
	.n {
		display: block;
		font-size: 1.4rem;
		letter-spacing: -0.03em;
	}
	.l {
		color: var(--muted);
		font-size: 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	.filters {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem 1.25rem;
		margin: 0 0 1.25rem;
	}
	.filter {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		margin: 0;
		color: var(--muted);
		font-size: 0.9rem;
		cursor: pointer;
		user-select: none;
	}
	.filter input {
		accent-color: var(--accent);
	}
	.error {
		color: var(--diff);
	}
</style>
