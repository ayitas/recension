<script lang="ts">
	import { login } from '$lib/api';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '$lib/auth';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	$effect(() => {
		if (isLoggedIn()) goto('/');
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			await login(email, password);
			goto('/');
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			loading = false;
		}
	}
</script>

<section class="panel">
	<h1>Log in</h1>
	<p class="muted">Sign in with your account.</p>
	<form onsubmit={onSubmit}>
		<label>
			Email
			<input type="email" bind:value={email} required autocomplete="username" />
		</label>
		<label>
			Password
			<input type="password" bind:value={password} required autocomplete="current-password" />
		</label>
		{#if error}<p class="error">{error}</p>{/if}
		<button type="submit" disabled={loading}>{loading ? 'Signing in…' : 'Log in'}</button>
	</form>
	<p class="muted tip">
		Local bootstrap: <code>dev@recension.local</code> / <code>dev-password</code>
	</p>
</section>

<style>
	.panel {
		max-width: 26rem;
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		padding: 1.5rem;
	}
	h1 {
		margin: 0 0 0.35rem;
		letter-spacing: -0.03em;
	}
	.muted {
		color: var(--muted);
	}
	.tip {
		margin-top: 1.25rem;
		font-size: 0.9rem;
	}
	form {
		display: grid;
		gap: 0.9rem;
		margin-top: 1.25rem;
	}
	label {
		display: grid;
		gap: 0.35rem;
		font-size: 0.9rem;
	}
	input {
		border: 1px solid var(--line);
		padding: 0.6rem 0.7rem;
		background: #fff;
	}
	button {
		border: 1px solid var(--accent);
		background: var(--accent-soft);
		color: var(--accent);
		padding: 0.65rem 0.9rem;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.6;
	}
	.error {
		color: var(--diff);
		margin: 0;
		font-size: 0.9rem;
	}
	code {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.85em;
	}
</style>
