<script lang="ts">
	import {
		getBatch,
		getElement,
		listBatches,
		type Batch,
		type ElementDetail
	} from '$lib/api';
	import DiffTable from '$lib/DiffTable.svelte';
	import MetricsBars from '$lib/MetricsBars.svelte';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let detail = $state<ElementDetail | null>(null);
	let error = $state('');
	let loading = $state(true);
	let siblings = $state<string[]>([]);
	let batches = $state<Batch[]>([]);

	const team = $derived($page.params.team ?? '');
	const suite = $derived($page.params.suite ?? '');
	const batch = $derived($page.params.batch ?? '');
	const element = $derived(decodeURIComponent($page.params.element ?? ''));

	const changedOnly = $derived($page.url.searchParams.get('changed') !== '0');
	const blobsOnly = $derived($page.url.searchParams.get('blobs') === '1');
	const vs = $derived($page.url.searchParams.get('vs') ?? '');

	function setParam(key: string, value: string | null) {
		const url = new URL($page.url);
		if (value === null || value === '') url.searchParams.delete(key);
		else url.searchParams.set(key, value);
		goto(`${url.pathname}${url.search}`, { replaceState: true, noScroll: true, keepFocus: true });
	}

	async function load() {
		if (!team || !suite || !batch || !element) return;
		loading = true;
		try {
			const [el, batchDetail, batchList] = await Promise.all([
				getElement(team, suite, batch, element, vs || undefined),
				getBatch(team, suite, batch),
				listBatches(team, suite)
			]);
			detail = el;
			siblings = batchDetail.elements.map((e) => e.testcase);
			batches = batchList;
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			detail = null;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		void team;
		void suite;
		void batch;
		void element;
		void vs;
		void load();
	});

	$effect(() => {
		function onKey(e: KeyboardEvent) {
			if (e.target instanceof HTMLInputElement || e.target instanceof HTMLSelectElement || e.target instanceof HTMLTextAreaElement) {
				return;
			}
			const idx = siblings.indexOf(element);
			if (idx < 0) return;
			if (e.key === 'j' || e.key === 'ArrowDown') {
				e.preventDefault();
				const next = siblings[idx + 1];
				if (next) goto(`/t/${team}/${suite}/${batch}/e/${encodeURIComponent(next)}${$page.url.search}`);
			}
			if (e.key === 'k' || e.key === 'ArrowUp') {
				e.preventDefault();
				const prev = siblings[idx - 1];
				if (prev) goto(`/t/${team}/${suite}/${batch}/e/${encodeURIComponent(prev)}${$page.url.search}`);
			}
		}
		window.addEventListener('keydown', onKey);
		return () => window.removeEventListener('keydown', onKey);
	});

	const dstLabel = $derived(
		vs || detail?.comparison?.dst?.version || 'baseline'
	);
</script>

<p class="crumb">
	<a href="/">Teams</a> /
	<a href={`/t/${team}`}>{team}</a> /
	<a href={`/t/${team}/${suite}`}>{suite}</a> /
	<a href={`/t/${team}/${suite}/${batch}`}>{batch}</a> /
	{element}
</p>

{#if error}
	<ErrorBanner message={error} onretry={load} />
{:else if loading && !detail}
	<Skeleton lines={5} />
{:else if detail}
	<div class="title-row">
		<h1 class="brand-font">{detail.testcase}</h1>
		<span class={`verdict ${detail.verdict}`}>{detail.verdict}</span>
	</div>

	<p class="meta muted">
		score <span class="mono">{(detail.score ?? 0).toFixed(3)}</span>
		· comparing
		<strong class="mono">{detail.comparison.src.version}</strong>
		vs
		<strong class="mono">{dstLabel}</strong>
		<span class="nav-hint">· j/k next/prev</span>
	</p>

	<div class="toolbar">
		<label class="filter">
			<input
				type="checkbox"
				checked={changedOnly}
				onchange={(e) => setParam('changed', (e.currentTarget as HTMLInputElement).checked ? null : '0')}
			/>
			Changed only
		</label>
		<label class="filter">
			<input
				type="checkbox"
				checked={blobsOnly}
				onchange={(e) =>
					setParam('blobs', (e.currentTarget as HTMLInputElement).checked ? '1' : null)}
			/>
			Blobs only (checks)
		</label>
		<label class="compare">
			Compare to
			<select
				value={vs}
				onchange={(e) => setParam('vs', (e.currentTarget as HTMLSelectElement).value || null)}
			>
				<option value="">Suite baseline</option>
				{#each batches.filter((b) => b.slug !== batch) as b}
					<option value={b.slug}>{b.slug}{b.isBaseline ? ' (baseline)' : ''}</option>
				{/each}
			</select>
		</label>
	</div>

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
				<span class="n mono"
					>{detail.comparison.overview.metricsDurationCommonSrc}→{detail.comparison.overview
						.metricsDurationCommonDst}</span
				>
				<span class="l">common ms (src→dst)</span>
			</div>
		{/if}
	</div>

	<DiffTable
		title="Checks"
		cellar={detail.comparison.results}
		srcLabel={detail.comparison.src.version}
		{dstLabel}
		{changedOnly}
		{blobsOnly}
	/>
	<DiffTable
		title="Assumptions"
		cellar={detail.comparison.asserts}
		srcLabel={detail.comparison.src.version}
		{dstLabel}
		{changedOnly}
		{blobsOnly}
	/>
	<MetricsBars
		cellar={detail.comparison.metrics}
		srcLabel={detail.comparison.src.version}
		{dstLabel}
		{changedOnly}
	/>
{/if}

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem 1rem;
		align-items: baseline;
		margin-bottom: 0.4rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.035em;
		font-size: clamp(1.65rem, 3.5vw, 2.2rem);
	}

	.meta {
		margin: 0 0 0.95rem;
		font-size: 0.9rem;
	}

	.nav-hint {
		opacity: 0.75;
		font-size: 0.8rem;
	}

	.toolbar {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem 1.1rem;
		align-items: center;
		margin: 0 0 1.1rem;
		padding: 0.65rem 0.95rem;
		background: var(--card);
		border: 1px solid var(--hairline);
		border-radius: var(--radius-md);
		position: sticky;
		top: var(--header-h);
		z-index: 5;
	}

	.filter {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		margin: 0;
		color: var(--body);
		font-size: 0.875rem;
		cursor: pointer;
		user-select: none;
		width: auto;
		font-weight: 500;
		padding: 0.35rem 0.75rem;
		border-radius: var(--radius-full);
		border: 1px solid var(--hairline);
		background: var(--canvas);
	}

	.filter input {
		width: auto;
		min-height: 0;
		accent-color: var(--accent);
	}

	.compare {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		margin: 0;
		width: auto;
		font-size: 0.875rem;
		color: var(--muted);
		font-weight: 500;
	}

	.compare select {
		width: auto;
		min-width: 9rem;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 7.5rem));
		gap: 0.55rem;
		margin-bottom: 0.65rem;
	}

	.metrics-stats {
		grid-template-columns: repeat(auto-fit, minmax(6.5rem, 7.5rem));
		margin-bottom: 1.15rem;
	}

	.stats div {
		background: var(--card);
		border: 1px solid var(--hairline);
		border-radius: var(--radius-md);
		padding: 0.55rem 0.7rem;
	}

	.duration .n {
		font-size: 0.95rem;
	}

	.n {
		display: block;
		font-size: 1.2rem;
		font-weight: 650;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
	}

	.l {
		color: var(--muted);
		font-size: 0.66rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.07em;
	}

	@media (max-width: 720px) {
		.stats {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}
		.toolbar {
			top: 0;
		}
	}
</style>
