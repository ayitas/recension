<script lang="ts">
	import '../app.css';
	import { clearSession, isLoggedIn } from '$lib/auth';
	import { applyTheme, getStoredTheme, toggleTheme, type Theme } from '$lib/theme';
	import { listTeams, type Team } from '$lib/api';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let { children } = $props();
	let loggedIn = $state(false);
	let theme = $state<Theme>('light');
	let teams = $state<Team[]>([]);

	const teamSlug = $derived($page.params.team ?? '');
	const suiteSlug = $derived($page.params.suite ?? '');
	const inApp = $derived(loggedIn && $page.url.pathname !== '/login' && $page.url.pathname !== '/signup');

	$effect(() => {
		theme = getStoredTheme();
		applyTheme(theme);
	});

	$effect(() => {
		void $page.url.pathname;
		loggedIn = isLoggedIn();
	});

	$effect(() => {
		if (!loggedIn) {
			teams = [];
			return;
		}
		listTeams()
			.then((t) => {
				teams = t;
			})
			.catch(() => {
				teams = [];
			});
	});

	function logout() {
		clearSession();
		loggedIn = false;
		goto('/login');
	}

	function onTheme() {
		theme = toggleTheme();
	}

	function onTeamChange(e: Event) {
		const slug = (e.currentTarget as HTMLSelectElement).value;
		if (slug) goto(`/t/${slug}`);
	}
</script>

<a class="skip-link" href="#main">Skip to content</a>

<div class="shell">
	<header class="top">
		<div class="left">
			<a class="brand" href="/">
				<img class="mark" src="/logo-mark.png" alt="" width="40" height="40" />
				<span class="name brand-font">Recension</span>
			</a>
			{#if !inApp}
				<p class="tag">trusted behavior, version after version</p>
			{/if}
			{#if inApp && teams.length > 0}
				<label class="switcher">
					<span class="sr-only">Team</span>
					<select value={teamSlug} onchange={onTeamChange} aria-label="Switch team">
						{#if !teamSlug}
							<option value="">Select team…</option>
						{/if}
						{#each teams as t}
							<option value={t.slug}>{t.name}</option>
						{/each}
					</select>
				</label>
				{#if teamSlug}
					<a class="chip" href={`/t/${teamSlug}`}>Suites</a>
					<a class="chip" href={`/t/${teamSlug}/members`}>Members</a>
					{#if suiteSlug}
						<a class="chip" href={`/t/${teamSlug}/${suiteSlug}`}>Batches</a>
					{/if}
				{/if}
			{/if}
		</div>
		<nav>
			<button type="button" class="linkish" onclick={onTheme} aria-label="Toggle color theme">
				{theme === 'dark' ? 'Light' : 'Dark'}
			</button>
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
		width: min(1280px, calc(100% - 2rem));
		margin: 0 auto;
		padding: 0 0 3.5rem;
	}

	.top {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1.25rem;
		min-height: var(--header-h);
		margin: 0 0 1.75rem;
		padding: 0.35rem 0;
		border-bottom: 1px solid var(--hairline);
		position: sticky;
		top: 0;
		z-index: 20;
		background: var(--canvas);
	}

	.left {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.45rem 0.75rem;
		min-width: 0;
	}

	.brand {
		display: inline-flex;
		align-items: center;
		gap: 0.55rem;
	}

	.brand:hover {
		color: inherit;
	}

	.mark {
		width: 2rem;
		height: 2rem;
		border-radius: var(--radius-sm);
		object-fit: cover;
		flex-shrink: 0;
	}

	.name {
		font-size: 1.35rem;
		letter-spacing: -0.04em;
		font-weight: 700;
		line-height: 1;
	}

	.tag {
		margin: 0;
		color: var(--muted);
		font-size: 0.875rem;
		max-width: 16rem;
		line-height: 1.3;
	}

	.switcher select {
		width: auto;
		min-width: 8.5rem;
	}

	.chip {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--ink);
		padding: 0.375rem 0.875rem;
		border-radius: var(--radius-full);
		border: 1px solid var(--hairline);
		background: var(--card);
		line-height: 1;
	}

	.chip:hover {
		color: var(--accent);
		border-color: color-mix(in srgb, var(--accent) 35%, var(--hairline));
		background: var(--card);
	}

	nav {
		display: flex;
		gap: 0.25rem 0.35rem;
		align-items: center;
	}

	nav a {
		color: var(--ink);
		padding: 0.5rem 0.85rem;
		border-radius: var(--radius-full);
		font-size: 0.875rem;
		font-weight: 600;
		line-height: 1;
	}

	nav a:hover {
		background: color-mix(in srgb, var(--ink) 5%, transparent);
		color: var(--ink);
	}

	nav a.cta {
		background: var(--accent);
		color: var(--on-accent);
		border: 1px solid transparent;
		min-height: 36px;
		display: inline-flex;
		align-items: center;
		padding: 0.5rem 1.15rem;
	}

	nav a.cta:hover {
		background: var(--accent-deep);
		color: var(--on-accent);
	}

	.linkish {
		border: 0;
		background: transparent;
		color: var(--muted);
		cursor: pointer;
		padding: 0.5rem 0.85rem;
		border-radius: var(--radius-full);
		font-size: 0.875rem;
		font-weight: 600;
		min-height: 36px;
	}

	.linkish:hover:not(:disabled) {
		background: color-mix(in srgb, var(--ink) 5%, transparent);
		color: var(--ink);
	}

	@media (max-width: 720px) {
		.tag {
			display: none;
		}
	}
</style>
