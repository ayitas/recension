<script lang="ts">
	import { me, rotateAPIKey } from '$lib/api';
	import type { AuthUser } from '$lib/auth';
	import { isLoggedIn } from '$lib/auth';
	import { goto } from '$app/navigation';
	import SdkSnippet from '$lib/SdkSnippet.svelte';

	let user = $state<AuthUser | null>(null);
	let error = $state('');
	let rotating = $state(false);
	let copied = $state(false);

	$effect(() => {
		if (!isLoggedIn()) {
			goto('/login');
			return;
		}
		me()
			.then((data) => {
				user = data;
				error = '';
			})
			.catch((err: Error) => {
				error = err.message;
				goto('/login');
			});
	});

	async function onRotate() {
		if (!confirm('Rotate API key? Existing SDK configs will stop working.')) return;
		rotating = true;
		try {
			user = await rotateAPIKey();
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			rotating = false;
		}
	}

	async function copyKey() {
		if (!user) return;
		await navigator.clipboard.writeText(user.apiKey);
		copied = true;
		setTimeout(() => (copied = false), 1500);
	}
</script>

<p class="crumb"><a href="/">Home</a> / Account</p>
<h1>Account</h1>

{#if error}
	<p class="error">{error}</p>
{/if}

{#if user}
	<section class="panel">
		<div class="row">
			<span class="label">Email</span>
			<span>{user.email}</span>
		</div>
		<div class="row">
			<span class="label">API key</span>
			<code>{user.apiKey}</code>
		</div>
		<div class="actions">
			<button type="button" onclick={copyKey}>{copied ? 'Copied' : 'Copy key'}</button>
			<button type="button" class="danger" onclick={onRotate} disabled={rotating}>
				{rotating ? 'Rotating…' : 'Rotate API key'}
			</button>
		</div>
		<p class="muted">
			Use this key with the Go SDK via <code>RECENSION_API_KEY</code> or
			<code>-api-key</code>.
		</p>
	</section>

	<section class="panel later">
		<h2>Quick start</h2>
		<p class="muted intro">
			After you create a team and suite, run a workflow like this (replace team/suite as needed).
		</p>
		<SdkSnippet team="acme" suite="students" apiKey={user.apiKey} />
	</section>
{/if}

<style>
	.crumb {
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	h1 {
		margin: 0 0 1.25rem;
		letter-spacing: -0.04em;
	}
	.panel {
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		padding: 1.25rem 1.4rem;
		max-width: 40rem;
	}
	.later {
		margin-top: 1.25rem;
		max-width: 44rem;
	}
	.later h2 {
		margin: 0 0 0.35rem;
		font-size: 1.05rem;
	}
	.intro {
		margin-bottom: 0.75rem;
	}
	.row {
		display: grid;
		gap: 0.35rem;
		margin-bottom: 1rem;
	}
	.label {
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--muted);
	}
	code {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.9rem;
		word-break: break-all;
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}
	button {
		border: 1px solid var(--accent);
		background: var(--accent-soft);
		color: var(--accent);
		padding: 0.55rem 0.85rem;
		cursor: pointer;
	}
	button.danger {
		border-color: var(--diff);
		background: rgba(154, 52, 18, 0.08);
		color: var(--diff);
	}
	.muted {
		color: var(--muted);
		margin: 0;
		font-size: 0.92rem;
		line-height: 1.45;
	}
	.error {
		color: var(--diff);
	}
</style>
