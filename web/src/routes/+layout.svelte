<script lang="ts">
	import '../app.css';
	import { clearSession, isLoggedIn } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let { children } = $props();
	let loggedIn = $state(false);

	$effect(() => {
		loggedIn = isLoggedIn();
	});

	// re-check when route changes
	$effect(() => {
		void $page.url.pathname;
		loggedIn = isLoggedIn();
	});

	function logout() {
		clearSession();
		loggedIn = false;
		goto('/login');
	}
</script>

<div class="shell">
	<header class="top">
		<div class="left">
			<a class="brand" href="/">
				<img class="mark" src="/logo-mark.png" alt="" width="40" height="40" />
				<span class="name">Recension</span>
			</a>
			<p class="tag">trusted behavior, version after version</p>
		</div>
		<nav>
			{#if loggedIn}
				<a href="/account">Account</a>
				<button type="button" class="linkish" onclick={logout}>Log out</button>
			{:else}
				<a href="/login">Log in</a>
				<a href="/signup">Sign up</a>
			{/if}
		</nav>
	</header>
	<main>
		{@render children()}
	</main>
</div>

<style>
	.shell {
		width: min(1100px, calc(100% - 2rem));
		margin: 0 auto;
		padding: 2rem 0 4rem;
	}

	.top {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1.5rem;
		margin-bottom: 2.5rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--line);
	}

	.left {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem 1.25rem;
	}

	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.65rem;
	}

	.mark {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 0.45rem;
		object-fit: cover;
		flex-shrink: 0;
	}

	.name {
		font-size: clamp(1.8rem, 4vw, 2.4rem);
		letter-spacing: -0.04em;
	}

	.tag {
		margin: 0;
		color: var(--muted);
		font-size: 0.95rem;
	}

	nav {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	nav a {
		color: var(--accent);
	}

	.linkish {
		border: 0;
		background: transparent;
		color: var(--muted);
		cursor: pointer;
		padding: 0;
	}
</style>
