<script lang="ts">
	import { downloadBlob, shortDigest } from '$lib/api';

	let {
		digest,
		label = 'blob',
		tone = ''
	}: {
		digest: string;
		label?: string;
		tone?: '' | 'del' | 'ins';
	} = $props();

	let busy = $state(false);
	let error = $state('');
	let copied = $state(false);
	let copyTimer: ReturnType<typeof setTimeout> | undefined;

	async function onDownload() {
		busy = true;
		error = '';
		try {
			await downloadBlob(digest, `${label}-${shortDigest(digest)}.bin`);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			busy = false;
		}
	}

	async function onCopy() {
		error = '';
		try {
			await navigator.clipboard.writeText(digest);
			copied = true;
			clearTimeout(copyTimer);
			copyTimer = setTimeout(() => {
				copied = false;
			}, 1200);
		} catch {
			error = 'could not copy';
		}
	}
</script>

<div class="blob" class:del={tone === 'del'} class:ins={tone === 'ins'}>
	<code title={digest}>{shortDigest(digest)}</code>
	<button type="button" onclick={onCopy}>{copied ? 'Copied' : 'Copy'}</button>
	<button type="button" onclick={onDownload} disabled={busy}>
		{busy ? '…' : 'Download'}
	</button>
</div>
{#if error}
	<p class="err">{error}</p>
{/if}

<style>
	.blob {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.45rem 0.65rem;
	}
	.blob.del code {
		background: rgba(154, 52, 18, 0.18);
	}
	.blob.ins code {
		background: rgba(15, 92, 76, 0.16);
	}
	code {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.8rem;
		padding: 0.15rem 0.35rem;
		background: rgba(28, 25, 21, 0.06);
	}
	button {
		border: 1px solid var(--line);
		background: var(--card);
		color: var(--accent);
		padding: 0.2rem 0.5rem;
		font-size: 0.75rem;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.6;
		cursor: wait;
	}
	.err {
		margin: 0.35rem 0 0;
		color: var(--diff);
		font-size: 0.75rem;
	}
</style>
