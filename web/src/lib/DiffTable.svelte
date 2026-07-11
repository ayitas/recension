<script lang="ts">
	import type { Cellar } from '$lib/api';
	import { charDiff, shouldHighlightString } from '$lib/stringDiff';
	import BlobCell from '$lib/BlobCell.svelte';

	let {
		title,
		cellar,
		srcLabel = 'This version',
		dstLabel = 'Baseline',
		changedOnly = false,
		blobsOnly = false
	}: {
		title: string;
		cellar: Cellar;
		srcLabel?: string;
		dstLabel?: string;
		changedOnly?: boolean;
		blobsOnly?: boolean;
	} = $props();

	function isBlob(type?: string, value?: string): boolean {
		return type === 'blob' || (!!value && value.startsWith('sha256:'));
	}

	function rowIsBlob(r: { srcType?: string; dstType?: string; srcValue?: string; dstValue?: string }) {
		return isBlob(r.srcType, r.srcValue) || isBlob(r.dstType, r.dstValue);
	}

	const allRows = $derived([
		...(cellar.commonKeys ?? []).map((c) => ({
			...c,
			kind: (c.score === 1 ? 'match' : 'changed') as 'match' | 'changed'
		})),
		...(cellar.newKeys ?? []).map((c) => ({ ...c, kind: 'fresh' as const, score: undefined })),
		...(cellar.missingKeys ?? []).map((c) => ({
			...c,
			kind: 'missing' as const,
			score: undefined
		}))
	]);

	const rows = $derived(
		allRows.filter((r) => {
			if (changedOnly && r.kind === 'match') return false;
			if (blobsOnly && !rowIsBlob(r)) return false;
			return true;
		})
	);

	const hiddenHint = $derived.by(() => {
		const parts: string[] = [];
		if (changedOnly) {
			const n = allRows.filter((r) => r.kind === 'match').length;
			if (n > 0) parts.push(`${n} matching hidden`);
		}
		if (blobsOnly) {
			const n = allRows.filter((r) => !rowIsBlob(r)).length;
			if (n > 0) parts.push(`${n} non-blobs hidden`);
		}
		return parts.join(' · ');
	});

	function valueParts(
		side: 'src' | 'dst',
		srcValue?: string,
		dstValue?: string,
		srcType?: string,
		dstType?: string,
		kind?: string
	) {
		const raw = side === 'src' ? srcValue : dstValue;
		if (raw === undefined || raw === '') return null;
		if (kind === 'match' || !shouldHighlightString(srcType, dstType)) {
			return [{ text: raw, kind: 'eq' as const }];
		}
		const a = srcValue ?? '';
		const b = dstValue ?? '';
		if (!a || !b) {
			return [{ text: raw, kind: (side === 'src' ? 'del' : 'ins') as 'del' | 'ins' }];
		}
		const diff = charDiff(a, b);
		return side === 'src' ? diff.src : diff.dst;
	}

	function blobTone(side: 'src' | 'dst', kind: string): '' | 'del' | 'ins' {
		if (kind !== 'changed') return '';
		return side === 'src' ? 'del' : 'ins';
	}
</script>

<section class="block">
	<div class="head">
		<h3>{title}</h3>
		{#if hiddenHint}
			<span class="hint">{hiddenHint}</span>
		{/if}
	</div>
	{#if allRows.length === 0}
		<p class="muted">No keys in this category.</p>
	{:else if rows.length === 0}
		<p class="muted">{blobsOnly ? 'No blob keys in this view.' : 'All keys match.'}</p>
	{:else}
		<table>
			<thead>
				<tr>
					<th>Key</th>
					<th>Status</th>
					<th>{srcLabel}</th>
					<th>{dstLabel}</th>
					<th>Score</th>
				</tr>
			</thead>
			<tbody>
				{#each rows as row}
					{@const srcParts = valueParts('src', row.srcValue, row.dstValue, row.srcType, row.dstType, row.kind)}
					{@const dstParts = valueParts('dst', row.srcValue, row.dstValue, row.srcType, row.dstType, row.kind)}
					<tr class={row.kind}>
						<td class="key">{row.name}</td>
						<td><span class={`tag ${row.kind}`}>{row.kind}</span></td>
						<td class="val">
							{#if isBlob(row.srcType, row.srcValue) && row.srcValue}
								<BlobCell digest={row.srcValue} label={row.name} tone={blobTone('src', row.kind)} />
							{:else if srcParts}
								{#each srcParts as part}<span class={part.kind}>{part.text}</span>{/each}
							{:else}
								—
							{/if}
						</td>
						<td class="val">
							{#if isBlob(row.dstType, row.dstValue) && row.dstValue}
								<BlobCell digest={row.dstValue} label={row.name} tone={blobTone('dst', row.kind)} />
							{:else if dstParts}
								{#each dstParts as part}<span class={part.kind}>{part.text}</span>{/each}
							{:else}
								—
							{/if}
						</td>
						<td>{row.score === undefined ? '—' : row.score.toFixed(3)}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</section>

<style>
	.block {
		margin: 0 0 1.75rem;
	}
	.head {
		display: flex;
		align-items: baseline;
		gap: 0.75rem;
		margin-bottom: 0.75rem;
	}
	h3 {
		margin: 0;
		font-size: 1rem;
		letter-spacing: -0.02em;
	}
	.hint {
		color: var(--muted);
		font-size: 0.8rem;
	}
	.muted {
		color: var(--muted);
		margin: 0;
	}
	table {
		width: 100%;
		border-collapse: collapse;
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		font-size: 0.95rem;
	}
	th,
	td {
		text-align: left;
		padding: 0.7rem 0.9rem;
		border-top: 1px solid var(--line);
		vertical-align: top;
	}
	th {
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--muted);
		border-top: 0;
	}
	.key {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.88rem;
	}
	.val {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.85rem;
		white-space: pre-wrap;
		word-break: break-word;
		max-width: 20rem;
	}
	.val .del {
		background: rgba(154, 52, 18, 0.22);
		box-decoration-break: clone;
	}
	.val .ins {
		background: rgba(15, 92, 76, 0.2);
		box-decoration-break: clone;
	}
	.tag {
		text-transform: uppercase;
		font-size: 0.7rem;
		letter-spacing: 0.08em;
	}
	.tag.match {
		color: var(--pass);
	}
	.tag.changed {
		color: var(--diff);
	}
	.tag.fresh {
		color: var(--accent);
	}
	.tag.missing {
		color: var(--muted);
	}
	tr.changed {
		background: rgba(154, 52, 18, 0.06);
	}
	tr.fresh {
		background: rgba(15, 92, 76, 0.06);
	}
	tr.missing {
		background: rgba(107, 99, 88, 0.08);
	}
</style>
