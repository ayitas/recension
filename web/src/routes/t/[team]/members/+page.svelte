<script lang="ts">
	import {
		addMember,
		listMembers,
		me,
		removeMember,
		roleAtLeast,
		updateMemberRole,
		type TeamMember,
		type TeamRole
	} from '$lib/api';
	import { isLoggedIn, type AuthUser } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	const ALL_ROLES: TeamRole[] = ['viewer', 'member', 'admin', 'owner'];

	let members = $state<TeamMember[]>([]);
	let user = $state<AuthUser | null>(null);
	let error = $state('');
	let formError = $state('');
	let email = $state('');
	let role = $state<TeamRole>('member');
	let adding = $state(false);
	let busyId = $state('');

	const team = $derived($page.params.team ?? '');
	const myMembership = $derived(members.find((m) => m.userId === user?.id));
	const myRole = $derived(myMembership?.role);
	const canManage = $derived(roleAtLeast(myRole, 'admin'));
	const canGrantOwner = $derived(myRole === 'owner');
	const inviteRoles = $derived(ALL_ROLES.filter((r) => r !== 'owner' || canGrantOwner));

	async function load() {
		if (!team) return;
		try {
			const [meUser, list] = await Promise.all([me(), listMembers(team)]);
			user = meUser;
			members = list.sort((a, b) => {
				const rank = (r: TeamRole) => ({ owner: 0, admin: 1, member: 2, viewer: 3 })[r];
				return rank(a.role) - rank(b.role) || a.email.localeCompare(b.email);
			});
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			if (error.toLowerCase().includes('authentication')) goto('/login');
		}
	}

	$effect(() => {
		if (!isLoggedIn()) {
			goto('/login');
			return;
		}
		void load();
	});

	function canEditMember(m: TeamMember): boolean {
		if (!canManage || !user) return false;
		if (m.userId === user.id) return false;
		if (m.role === 'owner' && !canGrantOwner) return false;
		return true;
	}

	function rolesFor(m: TeamMember): TeamRole[] {
		return ALL_ROLES.filter((r) => {
			if (r === 'owner' && !canGrantOwner) return false;
			if (m.role === 'owner' && r !== 'owner' && !canGrantOwner) return false;
			return true;
		});
	}

	async function onAdd(e: Event) {
		e.preventDefault();
		adding = true;
		formError = '';
		try {
			await addMember(team, email.trim(), role);
			email = '';
			role = 'member';
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			adding = false;
		}
	}

	async function onRoleChange(m: TeamMember, next: TeamRole) {
		if (next === m.role) return;
		busyId = m.userId;
		formError = '';
		try {
			await updateMemberRole(team, m.userId, next);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
			await load();
		} finally {
			busyId = '';
		}
	}

	async function onRemove(m: TeamMember) {
		if (!confirm(`Remove ${m.email} from ${team}?`)) return;
		busyId = m.userId;
		formError = '';
		try {
			await removeMember(team, m.userId);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			busyId = '';
		}
	}
</script>

<p class="crumb">
	<a href="/">Teams</a> /
	<a href={`/t/${team}`}>{team}</a> /
	Members
</p>

<div class="title-row">
	<h1>Members</h1>
	{#if myRole}
		<span class="badge">{myRole}</span>
	{/if}
</div>
<p class="muted lead">
	Team access is role-based: viewers read, members submit, admins manage suites and people, owners
	control ownership.
</p>

{#if error}
	<p class="error">{error}</p>
{:else}
	<ul class="list-card members">
		{#each members as m}
			<li>
				<div class="who">
					<span class="email">{m.email}</span>
					{#if m.userId === user?.id}
						<span class="you muted">you</span>
					{/if}
				</div>
				{#if canEditMember(m)}
					<div class="controls">
						<select
							aria-label={`Role for ${m.email}`}
							value={m.role}
							disabled={busyId === m.userId}
							onchange={(e) => onRoleChange(m, (e.currentTarget as HTMLSelectElement).value as TeamRole)}
						>
							{#each rolesFor(m) as r}
								<option value={r}>{r}</option>
							{/each}
						</select>
						<button
							type="button"
							class="danger"
							disabled={busyId === m.userId}
							onclick={() => onRemove(m)}
						>
							Remove
						</button>
					</div>
				{:else}
					<span class="role-label">{m.role}</span>
				{/if}
			</li>
		{/each}
	</ul>

	{#if canManage}
		<form class="panel create" onsubmit={onAdd}>
			<h3>Add member</h3>
			<p class="muted hint">User must already have a Recension account (same email as signup/login).</p>
			<div class="fields">
				<label>
					Email
					<input bind:value={email} type="email" placeholder="colleague@example.com" required />
				</label>
				<label>
					Role
					<select bind:value={role}>
						{#each inviteRoles as r}
							<option value={r}>{r}</option>
						{/each}
					</select>
				</label>
			</div>
			{#if formError}<p class="error">{formError}</p>{/if}
			<button type="submit" disabled={adding}>{adding ? 'Adding…' : 'Add member'}</button>
		</form>
	{:else if myRole}
		<p class="muted note">Only admins and owners can invite or change roles.</p>
	{/if}

	{#if formError && !canManage}
		<p class="error">{formError}</p>
	{/if}
{/if}

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		align-items: baseline;
		margin-bottom: 0.45rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.04em;
		font-size: clamp(1.9rem, 4vw, 2.6rem);
	}

	.badge {
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: var(--accent);
		background: var(--accent-soft);
		border: 1px solid rgba(15, 92, 76, 0.25);
		border-radius: var(--radius);
		padding: 0.25rem 0.55rem;
	}

	.lead {
		margin: 0 0 1.25rem;
		max-width: 36rem;
	}

	.members > li {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem 1rem;
		align-items: center;
		justify-content: space-between;
	}

	.who {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem;
		align-items: baseline;
		min-width: 12rem;
	}

	.email {
		font-weight: 500;
	}

	.you {
		font-size: 0.78rem;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.role-label {
		font-size: 0.9rem;
		color: var(--muted);
		text-transform: capitalize;
	}

	.controls {
		display: flex;
		flex-wrap: wrap;
		gap: 0.55rem;
		align-items: center;
	}

	.controls select {
		width: auto;
		min-width: 7.5rem;
	}

	.create {
		margin-top: 1.35rem;
	}

	h3 {
		margin: 0 0 0.35rem;
		font-size: 0.95rem;
	}

	.hint {
		margin: 0 0 0.85rem;
		font-size: 0.9rem;
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.85rem;
		margin-bottom: 0.95rem;
	}

	.note {
		margin-top: 1.25rem;
	}
</style>
