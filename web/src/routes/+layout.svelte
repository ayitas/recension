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
				<a class="cta" href="/signup">Sign up</a>
			{/if}
		</nav>
	</header>
	<main class="page-enter" id="main">
		{@render children()}
	</main>
</div>

<style>
	.shell {
		width: min(1100px, calc(100% - 2rem));
		margin: 0 auto;
		padding: 1.75rem 0 4.5rem;
	}

	.top {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.85rem 1.5rem;
		margin-bottom: 2.75rem;
		padding-bottom: 1.1rem;
		border-bottom: 1px solid color-mix(in srgb, var(--line) 80%, transparent);
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
		gap: 0.7rem;
		transition: transform 200ms var(--ease);
	}

	.brand:hover {
		color: inherit;
		transform: translateY(-1px);
	}

	.mark {
		width: 2.45rem;
		height: 2.45rem;
		border-radius: 0.5rem;
		object-fit: cover;
		flex-shrink: 0;
		box-shadow: 0 4px 14px rgba(15, 92, 76, 0.12);
		transition: box-shadow 200ms var(--ease);
	}

	.brand:hover .mark {
		box-shadow: 0 8px 20px rgba(15, 92, 76, 0.18);
	}

	.name {
		font-size: clamp(1.75rem, 4vw, 2.25rem);
		letter-spacing: -0.045em;
	}

	.tag {
		margin: 0;
		color: var(--muted);
		font-size: 0.92rem;
		max-width: 18rem;
		line-height: 1.35;
	}

	nav {
		display: flex;
		gap: 0.35rem 0.85rem;
		align-items: center;
	}

	nav a {
		color: var(--accent);
		padding: 0.35rem 0.55rem;
		border-radius: var(--radius);
	}

	nav a:hover {
		background: rgba(15, 92, 76, 0.08);
	}

	nav a.cta {
		background: var(--accent-soft);
		border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
	}

	nav a.cta:hover {
		background: var(--accent);
		color: #f7f3ea;
	}

	.linkish {
		border: 0;
		background: transparent;
		color: var(--muted);
		cursor: pointer;
		padding: 0.35rem 0.55rem;
		border-radius: var(--radius);
	}

	.linkish:hover:not(:disabled) {
		background: rgba(28, 25, 21, 0.05);
		color: var(--ink);
	}
</style>
