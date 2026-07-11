<script lang="ts">
	import { createSuite, listSuites, slugify, type Suite } from '$lib/api';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '$lib/auth';

	let suites = $state<Suite[]>([]);
	let error = $state('');
	let name = $state('');
	let slug = $state('');
	let slugTouched = $state(false);
	let creating = $state(false);
	let formError = $state('');

	const team = $derived($page.params.team ?? '');

	async function load() {
		if (!team) return;
		try {
			suites = await listSuites(team);
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			if (error.includes('authentication')) goto('/login');
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
<h1>{team}</h1>

{#if error}
	<p class="error">{error}</p>
{:else if suites.length === 0}
	<p class="muted">No suites yet. A suite maps to one workflow under test.</p>
{:else}
	<ul>
		{#each suites as suite}
			<li>
				<a href={`/t/${team}/${suite.slug}`}>{suite.name}</a>
				<span class="muted">baseline: {suite.baselineBatchId ? 'set' : 'none'}</span>
			</li>
		{/each}
	</ul>
{/if}

<form class="create" onsubmit={onCreate}>
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
	.crumb {
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	h1 {
		margin: 0 0 1.5rem;
		letter-spacing: -0.04em;
	}
	h3 {
		margin: 0 0 0.75rem;
		font-size: 0.95rem;
	}
	ul {
		list-style: none;
		padding: 0;
		margin: 0;
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
	}
	li {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-top: 1px solid var(--line);
	}
	li:first-child {
		border-top: 0;
	}
	.muted {
		color: var(--muted);
	}
	.error {
		color: var(--diff);
	}
	.create {
		margin-top: 1.25rem;
		background: var(--card);
		border: 1px solid var(--line);
		box-shadow: var(--shadow);
		padding: 1.1rem 1.25rem;
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
