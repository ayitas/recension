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

<section class="panel auth">
	<p class="eyebrow">welcome back</p>
	<h1>Log in</h1>
	<p class="muted">Sign in to browse teams, suites, and baselines.</p>
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
	<p class="muted switch">
		No account? <a href="/signup">Sign up</a>
	</p>
</section>

<style>
	.auth {
		max-width: 26rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		text-transform: uppercase;
		letter-spacing: 0.12em;
		font-size: 0.7rem;
		color: var(--accent);
	}

	h1 {
		margin: 0 0 0.35rem;
		letter-spacing: -0.035em;
	}

	form {
		display: grid;
		gap: 0.95rem;
		margin-top: 1.35rem;
	}

	.tip {
		margin-top: 1.35rem;
		font-size: 0.9rem;
	}

	.switch {
		margin: 0.85rem 0 0;
		font-size: 0.92rem;
	}

	.switch a {
		color: var(--accent);
		border-bottom: 1px solid rgba(15, 92, 76, 0.3);
	}

	code {
		font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
		font-size: 0.85em;
	}
</style>
