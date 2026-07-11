<script lang="ts">
	import type { Cellar } from '$lib/api';
	import DiffTable from '$lib/DiffTable.svelte';

	let {
		cellar,
		srcLabel = 'This version',
		dstLabel = 'Baseline',
		changedOnly = false
	}: {
		cellar: Cellar;
		srcLabel?: string;
		dstLabel?: string;
		changedOnly?: boolean;
	} = $props();

	type Kind = 'match' | 'changed' | 'fresh' | 'missing';

	type MetricRow = {
		name: string;
		kind: Kind;
		srcMs: number | null;
		dstMs: number | null;
		delta: number | null;
	};

	function parseMs(value?: string): number | null {
		if (value === undefined || value === '') return null;
		const n = Number(value);
		return Number.isFinite(n) ? n : null;
	}

	const allRows = $derived.by((): MetricRow[] => {
		const common = (cellar.commonKeys ?? []).map((c) => {
			const srcMs = parseMs(c.srcValue);
			const dstMs = parseMs(c.dstValue);
			const kind: Kind = c.score === 1 ? 'match' : 'changed';
			const delta = srcMs !== null && dstMs !== null ? srcMs - dstMs : null;
			return { name: c.name, kind, srcMs, dstMs, delta };
		});
		const fresh = (cellar.newKeys ?? []).map((c) => ({
			name: c.name,
			kind: 'fresh' as const,
			srcMs: parseMs(c.srcValue),
			dstMs: null,
			delta: null
		}));
		const missing = (cellar.missingKeys ?? []).map((c) => ({
			name: c.name,
			kind: 'missing' as const,
			srcMs: null,
			dstMs: parseMs(c.dstValue),
			delta: null
		}));
		return [...common, ...fresh, ...missing];
	});

	const rows = $derived(allRows.filter((r) => !(changedOnly && r.kind === 'match')));

	const maxMs = $derived.by(() => {
		let m = 1;
		for (const r of rows) {
			if (r.srcMs !== null) m = Math.max(m, r.srcMs);
			if (r.dstMs !== null) m = Math.max(m, r.dstMs);
		}
		return m;
	});

	function widthPct(ms: number | null): string {
		if (ms === null || ms <= 0) return '0%';
		return `${Math.max(2, (ms / maxMs) * 100)}%`;
	}

	function deltaLabel(row: MetricRow): string {
		if (row.kind === 'fresh') return 'new';
		if (row.kind === 'missing') return 'gone';
		if (row.delta === null) return '—';
		if (row.delta === 0) return 'same';
		const sign = row.delta > 0 ? '+' : '−';
		return `${sign}${Math.abs(row.delta)} ms`;
	}

	function deltaClass(row: MetricRow): string {
		if (row.kind === 'fresh') return 'fresh';
		if (row.kind === 'missing') return 'missing';
		if (row.delta === null || row.delta === 0) return 'same';
		return row.delta > 0 ? 'slower' : 'faster';
	}

	let showRaw = $state(false);
</script>

<section class="block">
	<div class="head">
		<h3>Metrics</h3>
		<span class="legend muted">
			<span class="swatch src"></span>{srcLabel}
			<span class="swatch dst"></span>{dstLabel}
		</span>
	</div>

	{#if allRows.length === 0}
		<p class="muted empty">No metrics recorded for this testcase.</p>
	{:else if rows.length === 0}
		<p class="muted empty">All metrics match baseline.</p>
	{:else}
		<ul class="metric-list">
			{#each rows as row}
				<li class={`row ${row.kind}`}>
					<div class="meta">
						<span class="key">{row.name}</span>
						<span class={`delta ${deltaClass(row)}`}>{deltaLabel(row)}</span>
					</div>
					<div class="bars" aria-label={`${row.name} duration comparison`}>
						<div class="track">
							<div
								class="bar src"
								style:width={widthPct(row.srcMs)}
								title={row.srcMs === null ? '—' : `${row.srcMs} ms (${srcLabel})`}
							></div>
							<span class="ms">{row.srcMs === null ? '—' : `${row.srcMs} ms`}</span>
						</div>
						<div class="track">
							<div
								class="bar dst"
								style:width={widthPct(row.dstMs)}
								title={row.dstMs === null ? '—' : `${row.dstMs} ms (${dstLabel})`}
							></div>
							<span class="ms muted">{row.dstMs === null ? '—' : `${row.dstMs} ms`}</span>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}

	{#if allRows.length > 0}
		<button type="button" class="raw-toggle" onclick={() => (showRaw = !showRaw)}>
			{showRaw ? 'Hide raw values' : 'Show raw values'}
		</button>
		{#if showRaw}
			<div class="raw">
				<DiffTable
					title="Raw metrics"
					{cellar}
					{srcLabel}
					{dstLabel}
					{changedOnly}
					blobsOnly={false}
				/>
			</div>
		{/if}
	{/if}
</section>

<style>
	.block {
		margin: 0 0 1.75rem;
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.75rem 1.25rem;
		margin-bottom: 0.85rem;
	}

	h3 {
		margin: 0;
		font-size: 1rem;
		letter-spacing: -0.02em;
	}

	.legend {
		display: inline-flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.45rem 0.65rem;
		font-size: 0.82rem;
	}

	.swatch {
		display: inline-block;
		width: 0.7rem;
		height: 0.7rem;
		border-radius: 0.15rem;
		margin-right: 0.15rem;
		vertical-align: middle;
	}

	.swatch.src {
		background: var(--accent);
	}

	.swatch.dst {
		background: rgba(107, 99, 88, 0.45);
	}

	.empty {
		margin: 0;
	}

	.metric-list {
		list-style: none;
		margin: 0;
		padding: 0;
		background: var(--card);
		border: 1px solid var(--line);
		border-radius: calc(var(--radius) + 0.15rem);
		box-shadow: var(--shadow-soft);
		overflow: hidden;
	}

	.row {
		padding: 0.95rem 1.1rem;
		border-top: 1px solid var(--line);
		animation: rise 320ms var(--ease) both;
	}

	.row:first-child {
		border-top: 0;
	}

	.row.changed {
		background: rgba(154, 52, 18, 0.04);
	}

	.row.fresh {
		background: rgba(15, 92, 76, 0.04);
	}

	.row.missing {
		background: rgba(107, 99, 88, 0.05);
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		justify-content: space-between;
		gap: 0.5rem 1rem;
		margin-bottom: 0.55rem;
	}

	.key {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.9rem;
	}

	.delta {
		font-variant-numeric: tabular-nums;
		font-size: 0.92rem;
		letter-spacing: -0.02em;
	}

	.delta.slower {
		color: var(--diff);
	}

	.delta.faster {
		color: var(--pass);
	}

	.delta.same {
		color: var(--muted);
	}

	.delta.fresh {
		color: var(--accent);
	}

	.delta.missing {
		color: var(--muted);
	}

	.bars {
		display: grid;
		gap: 0.35rem;
	}

	.track {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: 0.65rem;
		align-items: center;
		min-height: 0.85rem;
	}

	.bar {
		height: 0.55rem;
		border-radius: 0.2rem;
		max-width: 100%;
		transition: width 420ms var(--ease);
	}

	.bar.src {
		background: var(--accent);
	}

	.bar.dst {
		background: rgba(107, 99, 88, 0.42);
	}

	.ms {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.78rem;
		font-variant-numeric: tabular-nums;
		min-width: 4.5rem;
		text-align: right;
	}

	.raw-toggle {
		margin-top: 0.85rem;
		padding: 0.35rem 0.7rem;
		font-size: 0.85rem;
		background: transparent;
		border-color: var(--line-strong);
		color: var(--muted);
	}

	.raw-toggle:hover:not(:disabled) {
		background: rgba(15, 92, 76, 0.06);
		color: var(--accent);
		border-color: rgba(15, 92, 76, 0.35);
	}

	.raw {
		margin-top: 0.75rem;
	}

	.raw :global(.block) {
		margin-bottom: 0;
	}

	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(4px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (max-width: 640px) {
		.ms {
			min-width: 3.6rem;
		}
	}
</style>
