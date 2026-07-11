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
		<p class="empty">No teams yet. Create one to get started.</p>
	{:else}
		<ul>
			{#each teams as team}
				<li>
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
		margin-bottom: 2.5rem;
	}

	.eyebrow {
		margin: 0 0 0.75rem;
		text-transform: uppercase;
		letter-spacing: 0.14em;
		font-size: 0.72rem;
		color: var(--accent);
	}

	h1 {
		margin: 0;
		max-width: 14ch;
		font-size: clamp(2.4rem, 7vw, 4.2rem);
		line-height: 0.95;
		letter-spacing: -0.05em;
	}

	.lead {
		max-width: 38rem;
		margin: 1.25rem 0 0;
		color: var(--muted);
		font-size: 1.1rem;
		line-height: 1.55;
	}

	.panel {
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		padding: 1.25rem 1.5rem;
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
	}

	h3 {
		margin: 0 0 0.75rem;
		font-size: 0.95rem;
	}

	ul {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	li {
		display: flex;
		gap: 0.75rem;
		align-items: baseline;
		padding: 0.85rem 0;
		border-top: 1px solid var(--line);
	}

	li a {
		font-size: 1.2rem;
	}

	.empty {
		color: var(--muted);
		margin: 0 0 0.5rem;
	}

	.muted {
		color: var(--muted);
	}

	.error {
		color: var(--diff);
		line-height: 1.5;
	}

	.create {
		margin-top: 1.25rem;
		padding-top: 1.25rem;
		border-top: 1px solid var(--line);
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.75rem;
		margin-bottom: 0.85rem;
	}

	label {
		display: grid;
		gap: 0.3rem;
		font-size: 0.85rem;
	}

	input {
		border: 1px solid var(--line);
		padding: 0.55rem 0.65rem;
		background: #fff;
	}

	button {
		border: 1px solid var(--accent);
		background: var(--accent-soft);
		color: var(--accent);
		padding: 0.55rem 0.85rem;
		cursor: pointer;
	}

	button:disabled {
		opacity: 0.6;
	}
</style>
