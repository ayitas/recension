<script lang="ts">
	import { createTeam, listTeams, slugify, type Team } from '$lib/api';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { isLoggedIn } from '$lib/auth';
	import { goto } from '$app/navigation';

	let teams = $state<Team[]>([]);
	let error = $state('');
	let loading = $state(true);
	let name = $state('');
	let slug = $state('');
	let slugTouched = $state(false);
	let creating = $state(false);
	let formError = $state('');

	async function load() {
		loading = true;
		try {
			teams = await listTeams();
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			if (error.includes('401') || error.toLowerCase().includes('authentication')) {
				goto('/login');
			}
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (!isLoggedIn()) {
			goto('/login');
			return;
		}
		void load();
	});

	$effect(() => {
		if (!slugTouched) {
			slug = slugify(name);
		}
	});

	async function onCreate(e: Event) {
		e.preventDefault();
		creating = true;
		formError = '';
		try {
			const team = await createTeam(name.trim() || slug, slug);
			name = '';
			slug = '';
			slugTouched = false;
			await load();
			goto(`/t/${team.slug}`);
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			creating = false;
		}
	}
</script>

<section class="intro">
	<h1 class="brand-font">Teams</h1>
	<p class="muted lead">
		Pick a team to browse suites and baselines, or create one — you’ll be owner.
	</p>
</section>

<section class="panel">
	<div class="panel-head">
		<h2>Your teams</h2>
		{#if loading}<span class="muted">loading…</span>{/if}
	</div>

	{#if error}
		<ErrorBanner message={error} onretry={load} />
	{:else if loading}
		<Skeleton lines={3} title={false} />
	{:else if teams.length === 0}
		<p class="empty muted">No teams yet. Create one to get started.</p>
	{:else}
		<ul class="list-card teams">
			{#each teams as team, i}
				<li style={`--i: ${i}`}>
					<a href={`/t/${team.slug}`}>{team.name}</a>
					<span class="muted mono">/{team.slug}</span>
				</li>
			{/each}
		</ul>
	{/if}

	<form class="create" onsubmit={onCreate}>
		<h3>New team</h3>
		<div class="fields">
			<label>
				Name
				<input bind:value={name} placeholder="Acme Engineering" required />
			</label>
			<label>
				Slug
				<input
					bind:value={slug}
					placeholder="acme"
					required
					oninput={() => (slugTouched = true)}
				/>
			</label>
		</div>
		{#if formError}<p class="error">{formError}</p>{/if}
		<button type="submit" disabled={creating}>{creating ? 'Creating…' : 'Create team'}</button>
	</form>
</section>

<style>
	.intro {
		margin-bottom: 1.35rem;
	}

	h1 {
		margin: 0;
		font-size: clamp(1.85rem, 4vw, 2.4rem);
	}

	.lead {
		margin: 0.45rem 0 0;
		max-width: 36rem;
		font-size: 1rem;
		color: var(--body);
		line-height: 1.5;
	}

	.panel-head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 0.85rem;
	}

	h2 {
		margin: 0;
		font-size: 1rem;
		font-weight: 650;
		letter-spacing: -0.015em;
	}

	h3 {
		margin: 0 0 0.75rem;
		font-size: 0.9rem;
		font-weight: 650;
	}

	.teams {
		margin-bottom: 0.25rem;
	}

	.teams > li {
		animation: rise 360ms var(--ease) both;
		animation-delay: calc(var(--i) * 40ms);
	}

	.empty {
		margin: 0 0 0.5rem;
	}

	.create {
		margin-top: 1.15rem;
		padding-top: 1.05rem;
		border-top: 1px solid var(--hairline);
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.75rem;
		margin-bottom: 0.85rem;
	}
</style>
