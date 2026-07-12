<script lang="ts">
	import { createSuite, listSuites, slugify, type Suite } from '$lib/api';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '$lib/auth';

	let suites = $state<Suite[]>([]);
	let error = $state('');
	let loading = $state(true);
	let name = $state('');
	let slug = $state('');
	let slugTouched = $state(false);
	let creating = $state(false);
	let formError = $state('');

	const team = $derived($page.params.team ?? '');

	async function load() {
		if (!team) return;
		loading = true;
		try {
			suites = await listSuites(team);
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			if (error.includes('authentication')) goto('/login');
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
		if (!slugTouched) slug = slugify(name);
	});

	async function onCreate(e: Event) {
		e.preventDefault();
		creating = true;
		formError = '';
		try {
			const suite = await createSuite(team, name.trim() || slug, slug);
			name = '';
			slug = '';
			slugTouched = false;
			await load();
			goto(`/t/${team}/${suite.slug}`);
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			creating = false;
		}
	}
</script>

<p class="crumb"><a href="/">Teams</a> / {team}</p>
<div class="title-row">
	<h1 class="brand-font">{team}</h1>
	<a class="members-link" href={`/t/${team}/members`}>Members</a>
</div>

{#if error}
	<ErrorBanner message={error} onretry={load} />
{:else if loading}
	<Skeleton lines={3} />
{:else if suites.length === 0}
	<p class="muted lead">No suites yet. A suite maps to one workflow under test.</p>
{:else}
	<ul class="list-card">
		{#each suites as suite}
			<li>
				<a href={`/t/${team}/${suite.slug}`}>{suite.name}</a>
				<span class="muted status">
					baseline: {suite.baselineBatchId ? 'set' : 'none'}
				</span>
			</li>
		{/each}
	</ul>
{/if}

<form class="panel create" onsubmit={onCreate}>
	<h3>New suite</h3>
	<div class="fields">
		<label>
			Name
			<input bind:value={name} placeholder="Students workflow" required />
		</label>
		<label>
			Slug
			<input bind:value={slug} placeholder="students" required oninput={() => (slugTouched = true)} />
		</label>
	</div>
	{#if formError}<p class="error">{formError}</p>{/if}
	<button type="submit" disabled={creating}>{creating ? 'Creating…' : 'Create suite'}</button>
</form>

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: baseline;
		justify-content: space-between;
		margin-bottom: 1.15rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.035em;
		font-size: clamp(1.65rem, 3.5vw, 2.2rem);
	}

	.members-link {
		font-size: 0.9rem;
		font-weight: 550;
		color: var(--accent);
		border-bottom: 1px solid transparent;
	}

	.members-link:hover {
		border-bottom-color: var(--accent);
	}

	.lead {
		margin: 0 0 1rem;
	}

	.status {
		font-size: 0.84rem;
	}

	.create {
		margin-top: 1.15rem;
	}

	h3 {
		margin: 0 0 0.75rem;
		font-size: 0.9rem;
		font-weight: 650;
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.75rem;
		margin-bottom: 0.85rem;
	}
</style>
