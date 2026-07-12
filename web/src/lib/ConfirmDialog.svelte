<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		title = 'Confirm',
		confirmLabel = 'Confirm',
		danger = false,
		busy = false,
		onconfirm,
		children
	}: {
		open?: boolean;
		title?: string;
		confirmLabel?: string;
		danger?: boolean;
		busy?: boolean;
		onconfirm: () => void | Promise<void>;
		children?: Snippet;
	} = $props();

	let dialog = $state<HTMLDialogElement | null>(null);

	$effect(() => {
		const el = dialog;
		if (!el) return;
		if (open && !el.open) el.showModal();
		if (!open && el.open) el.close();
	});

	function onCancel() {
		open = false;
	}

	async function onOk() {
		await onconfirm();
		open = false;
	}
</script>

<dialog
	bind:this={dialog}
	class="dlg"
	onclick={(e) => {
		if (e.target === dialog) onCancel();
	}}
	onclose={() => (open = false)}
>
	<form method="dialog" class="panel inner" onsubmit={(e) => e.preventDefault()}>
		<h2>{title}</h2>
		<div class="body">
			{#if children}{@render children()}{/if}
		</div>
		<div class="actions">
			<button type="button" class="ghost" onclick={onCancel} disabled={busy}>Cancel</button>
			<button type="button" class={danger ? 'danger' : ''} onclick={onOk} disabled={busy}>
				{busy ? 'Working…' : confirmLabel}
			</button>
		</div>
	</form>
</dialog>

<style>
	.dlg {
		border: 0;
		padding: 0;
		background: transparent;
		max-width: min(26rem, calc(100vw - 2rem));
	}

	.dlg::backdrop {
		background: rgba(28, 25, 21, 0.45);
	}

	:global(html[data-theme='dark']) .dlg::backdrop {
		background: rgba(0, 0, 0, 0.65);
	}

	.inner {
		margin: 0;
	}

	h2 {
		margin: 0 0 0.65rem;
		font-family: var(--font-brand);
		font-size: 1.35rem;
		letter-spacing: -0.03em;
		font-weight: 600;
		line-height: 1.15;
	}

	.body {
		color: var(--body);
		font-size: 0.95rem;
		line-height: 1.5;
		margin-bottom: 1.1rem;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 0.55rem;
	}
</style>
