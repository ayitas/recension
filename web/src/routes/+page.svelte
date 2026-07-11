<script lang="ts">
	import { createTeam, listTeams, slugify, type Team } from '$lib/api';
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
			if (error.includes('401') || error.includes('authentication')) {
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

<section class="hero">
	<p class="eyebrow">continuous regression testing</p>
	<h1>Establish a trusted reading of your software.</h1>
	<p class="lead">
		Create a team, add a suite, then submit behavior from the Go SDK. The first revision becomes your
		baseline.
	</p>
</section>

<section class="panel">
	<div class="panel-head">
		<h2>Teams</h2>
		{#if loading}<span class="muted">loading…</span>{/if}
	</div>

	{#if error}
		<p class="error">
			{error}. If you are not signed in, <a href="/login">log in</a>.
		</p>
	{:else if teams.length === 0 && !loading}
		<p class="empty muted">No teams yet. Create one to get started.</p>
	{:else}
		<ul class="list-card teams">
			{#each teams as team, i}
				<li style={`--i: ${i}`}>
					<a href={`/t/${team.slug}`}>{team.name}</a>
					<span class="muted">/{team.slug}</span>
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
	.hero {
		margin-bottom: 2.75rem;
	}

	.eyebrow {
		margin: 0 0 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.14em;
		font-size: 0.72rem;
		color: var(--accent);
	}

	h1 {
		margin: 0;
		max-width: 14ch;
		font-size: clamp(2.4rem, 7vw, 4.1rem);
		line-height: 0.96;
		letter-spacing: -0.05em;
	}

	.lead {
		max-width: 38rem;
		margin: 1.35rem 0 0;
		color: var(--muted);
		font-size: 1.12rem;
		line-height: 1.55;
	}

	.panel-head {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 1rem;
	}

	h2 {
		margin: 0;
		font-size: 1.1rem;
		letter-spacing: -0.02em;
	}

	h3 {
		margin: 0 0 0.85rem;
		font-size: 0.95rem;
	}

	.teams {
		margin-bottom: 0.25rem;
	}

	.teams > li {
		animation: rise 420ms var(--ease) both;
		animation-delay: calc(var(--i) * 45ms);
	}

	.empty {
		margin: 0 0 0.5rem;
	}

	.create {
		margin-top: 1.4rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--line);
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.85rem;
		margin-bottom: 0.95rem;
	}
</style>
