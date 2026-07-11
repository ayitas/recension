<script lang="ts">
	import { getElement, type ElementDetail } from '$lib/api';
	import DiffTable from '$lib/DiffTable.svelte';
	import MetricsBars from '$lib/MetricsBars.svelte';
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

	<p class="meta muted">
		score {(detail.score ?? 0).toFixed(3)}
		· comparing
		<strong>{detail.comparison.src.version}</strong>
		vs baseline
		<strong>{detail.comparison.dst.version}</strong>
	</p>

	<div class="stats">
		<div>
			<span class="n">{detail.comparison.overview.keysCountCommon}</span>
			<span class="l">checks common</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.keysCountFresh}</span>
			<span class="l">checks fresh</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.keysCountMissing}</span>
			<span class="l">checks missing</span>
		</div>
	</div>

	<div class="stats metrics-stats">
		<div>
			<span class="n">{detail.comparison.overview.metricsCountCommon}</span>
			<span class="l">metrics common</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.metricsCountFresh}</span>
			<span class="l">metrics fresh</span>
		</div>
		<div>
			<span class="n">{detail.comparison.overview.metricsCountMissing}</span>
			<span class="l">metrics missing</span>
		</div>
		{#if detail.comparison.overview.metricsCountCommon > 0}
			<div class="duration">
				<span class="n"
					>{detail.comparison.overview.metricsDurationCommonSrc}→{detail.comparison.overview
						.metricsDurationCommonDst}</span
				>
				<span class="l">common ms (src→dst)</span>
			</div>
		{/if}
	</div>

	<div class="filters">
		<label class="filter">
			<input type="checkbox" bind:checked={changedOnly} />
			Changed only
		</label>
		<label class="filter">
			<input type="checkbox" bind:checked={blobsOnly} />
			Blobs only (checks)
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
	<MetricsBars
		cellar={detail.comparison.metrics}
		srcLabel={detail.comparison.src.version}
		dstLabel={detail.comparison.dst.version}
		{changedOnly}
	/>
{/if}

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: baseline;
		margin-bottom: 0.55rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.04em;
		font-size: clamp(1.9rem, 4vw, 2.6rem);
	}

	.meta {
		margin: 0 0 1.35rem;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 8.5rem));
		gap: 0.75rem;
		margin-bottom: 0.85rem;
	}

	.metrics-stats {
		grid-template-columns: repeat(auto-fit, minmax(7.5rem, 8.5rem));
		margin-bottom: 1.5rem;
	}

	.stats div {
		background: var(--card);
		border: 1px solid var(--line);
		border-radius: var(--radius);
		padding: 0.8rem 0.95rem;
		box-shadow: var(--shadow-soft);
		transition:
			transform 180ms var(--ease),
			box-shadow 180ms var(--ease);
	}

	.stats div:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow);
	}

	.duration .n {
		font-size: 1.05rem;
	}

	.n {
		display: block;
		font-size: 1.45rem;
		letter-spacing: -0.03em;
		font-variant-numeric: tabular-nums;
	}

	.l {
		color: var(--muted);
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}

	.filters {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem 1.25rem;
		margin: 0 0 1.35rem;
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
		width: auto;
	}

	.filter input {
		width: auto;
		accent-color: var(--accent);
	}
</style>
