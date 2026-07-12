<script lang="ts">
	import { signup } from '$lib/api';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { isLoggedIn } from '$lib/auth';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	const next = $derived($page.url.searchParams.get('next') || '/account');

	$effect(() => {
		if (isLoggedIn()) goto(next);
	});

	async function onSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';
		try {
			await signup(email, password);
			goto(next);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			loading = false;
		}
	}
</script>

<section class="panel auth">
	<p class="eyebrow">get started</p>
	<h1 class="brand-font">Sign up</h1>
	<p class="lead">Create an account to browse results and manage your API key.</p>
	<form onsubmit={onSubmit}>
		<label>
			Email
			<input type="email" bind:value={email} required autocomplete="username" />
		</label>
		<label>
			Password
			<input
				type="password"
				bind:value={password}
				required
				minlength="8"
				autocomplete="new-password"
			/>
		</label>
		{#if error}<p class="error">{error}</p>{/if}
		<button type="submit" disabled={loading}>{loading ? 'Creating…' : 'Create account'}</button>
	</form>
	<p class="muted switch">
		Already have an account?
		<a href={`/login${next !== '/account' ? `?next=${encodeURIComponent(next)}` : ''}`}>Log in</a>
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

	.switch {
		margin: 1.15rem 0 0;
		font-size: 0.9375rem;
	}

	.switch a {
		color: var(--accent);
		font-weight: 600;
	}
</style>
