<script lang="ts">
	import { login } from '$lib/api';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLoggedIn } from '$lib/auth';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	const next = $derived($page.url.searchParams.get('next') || '/');

	$effect(() => {
		if (isLoggedIn()) goto(next);
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			await login(email, password);
			goto(next);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			loading = false;
		}
	}
</script>

<section class="panel auth">
	<p class="eyebrow">welcome back</p>
	<h1 class="brand-font">Log in</h1>
	<p class="lead">Sign in to browse teams, suites, and baselines.</p>
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
		Local bootstrap: <code class="mono">dev@recension.local</code> /
		<code class="mono">dev-password</code>
	</p>
	<p class="muted switch">
		No account?
		<a href={`/signup${next !== '/' ? `?next=${encodeURIComponent(next)}` : ''}`}>Sign up</a>
	</p>
</section>

<style>
	.auth {
		max-width: 26rem;
	}

	.eyebrow {
		margin: 0 0 0.55rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--accent);
	}

	h1 {
		margin: 0 0 0.35rem;
		font-size: clamp(2rem, 5vw, 2.5rem);
	}

	.lead {
		margin: 0;
		color: var(--body);
		font-size: 1rem;
		line-height: 1.5;
	}

	form {
		display: grid;
		gap: 0.95rem;
		margin-top: 1.5rem;
	}

	.tip {
		margin-top: 1.35rem;
		font-size: 0.875rem;
	}

	.switch {
		margin: 0.85rem 0 0;
		font-size: 0.9375rem;
	}

	.switch a {
		color: var(--accent);
		font-weight: 600;
	}
</style>
