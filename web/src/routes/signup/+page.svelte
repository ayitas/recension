<script lang="ts">
	import { signup } from '$lib/api';
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
			await signup(email, password);
			goto('/account');
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			loading = false;
		}
	}
</script>

<section class="panel auth">
	<p class="eyebrow">get started</p>
	<h1>Sign up</h1>
	<p class="muted">Create an account to browse results and manage your API key.</p>
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
		Already have an account? <a href="/login">Log in</a>
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

	.switch {
		margin: 1.15rem 0 0;
		font-size: 0.92rem;
	}

	.switch a {
		color: var(--accent);
		border-bottom: 1px solid rgba(15, 92, 76, 0.3);
	}
</style>
