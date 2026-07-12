<script lang="ts">
	import {
		addMember,
		createInvite,
		listMembers,
		me,
		removeMember,
		roleAtLeast,
		updateMemberRole,
		type TeamMember,
		type TeamRole
	} from '$lib/api';
	import ConfirmDialog from '$lib/ConfirmDialog.svelte';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { isLoggedIn, type AuthUser } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	const ALL_ROLES: TeamRole[] = ['viewer', 'member', 'admin', 'owner'];

	let members = $state<TeamMember[]>([]);
	let user = $state<AuthUser | null>(null);
	let error = $state('');
	let loading = $state(true);
	let formError = $state('');
	let email = $state('');
	let role = $state<TeamRole>('member');
	let inviteRole = $state<TeamRole>('member');
	let adding = $state(false);
	let busyId = $state('');
	let inviteBusy = $state(false);
	let inviteLink = $state('');
	let inviteCopied = $state(false);
	let removeTarget = $state<TeamMember | null>(null);
	let confirmOpen = $state(false);

	const team = $derived($page.params.team ?? '');
	const myMembership = $derived(members.find((m) => m.userId === user?.id));
	const myRole = $derived(myMembership?.role);
	const canManage = $derived(roleAtLeast(myRole, 'admin'));
	const canGrantOwner = $derived(myRole === 'owner');
	const inviteRoles = $derived(ALL_ROLES.filter((r) => r !== 'owner' || canGrantOwner));

	async function load() {
		if (!team) return;
		loading = true;
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

	function askRemove(m: TeamMember) {
		removeTarget = m;
		confirmOpen = true;
	}

	async function doRemove() {
		if (!removeTarget) return;
		busyId = removeTarget.userId;
		formError = '';
		try {
			await removeMember(team, removeTarget.userId);
			await load();
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			busyId = '';
			removeTarget = null;
		}
	}

	async function onCreateInvite() {
		inviteBusy = true;
		formError = '';
		try {
			const inv = await createInvite(team, inviteRole);
			inviteLink = `${window.location.origin}${inv.path}`;
		} catch (err) {
			formError = err instanceof Error ? err.message : String(err);
		} finally {
			inviteBusy = false;
		}
	}

	async function copyInvite() {
		if (!inviteLink) return;
		await navigator.clipboard.writeText(inviteLink);
		inviteCopied = true;
		setTimeout(() => (inviteCopied = false), 1500);
	}
</script>

<p class="crumb">
	<a href="/">Teams</a> /
	<a href={`/t/${team}`}>{team}</a> /
	Members
</p>

<div class="title-row">
	<h1 class="brand-font">Members</h1>
	{#if myRole}
		<span class="badge">{myRole}</span>
	{/if}
</div>
<p class="muted lead">
	Viewers read, members submit, admins manage suites and people, owners control ownership.
</p>

{#if error}
	<ErrorBanner message={error} onretry={load} />
{:else if loading}
	<Skeleton lines={4} />
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
							onclick={() => askRemove(m)}
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
			<h3>Add existing user</h3>
			<p class="muted hint">User must already have a Recension account.</p>
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

		<section class="panel create">
			<h3>Invite link</h3>
			<p class="muted hint">Share a link — recipient logs in (or signs up) then joins with the chosen role.</p>
			<div class="fields">
				<label>
					Role
					<select bind:value={inviteRole}>
						{#each inviteRoles as r}
							<option value={r}>{r}</option>
						{/each}
					</select>
				</label>
			</div>
			<div class="invite-actions">
				<button type="button" onclick={onCreateInvite} disabled={inviteBusy}>
					{inviteBusy ? 'Creating…' : 'Create invite link'}
				</button>
				{#if inviteLink}
					<code class="link mono">{inviteLink}</code>
					<button type="button" class="ghost" onclick={copyInvite}>
						{inviteCopied ? 'Copied' : 'Copy'}
					</button>
				{/if}
			</div>
		</section>
	{:else if myRole}
		<p class="muted note">Only admins and owners can invite or change roles.</p>
	{/if}
{/if}

<ConfirmDialog
	bind:open={confirmOpen}
	title="Remove member?"
	confirmLabel="Remove"
	danger
	busy={busyId !== ''}
	onconfirm={doRemove}
>
	{#if removeTarget}
		<p>Remove <strong>{removeTarget.email}</strong> from <span class="mono">{team}</span>?</p>
	{/if}
</ConfirmDialog>

<style>
	.title-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem;
		align-items: baseline;
		margin-bottom: 0.35rem;
	}

	h1 {
		margin: 0;
		letter-spacing: -0.035em;
		font-size: clamp(1.65rem, 3.5vw, 2.2rem);
	}

	.badge {
		font-size: 0.68rem;
		font-weight: 650;
		text-transform: uppercase;
		letter-spacing: 0.07em;
		color: var(--accent);
		background: var(--accent-soft);
		border: 1px solid color-mix(in srgb, var(--accent) 25%, transparent);
		border-radius: var(--radius);
		padding: 0.2rem 0.45rem;
	}

	.lead {
		margin: 0 0 1.1rem;
		max-width: 36rem;
		font-size: 0.95rem;
	}

	.members > li {
		display: flex;
		flex-wrap: wrap;
		gap: 0.65rem 1rem;
		align-items: center;
		justify-content: space-between;
	}

	.who {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
		align-items: baseline;
		min-width: 12rem;
	}

	.email {
		font-weight: 550;
	}

	.you {
		font-size: 0.72rem;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.role-label {
		font-size: 0.88rem;
		color: var(--muted);
		text-transform: capitalize;
	}

	.controls {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		align-items: center;
	}

	.controls select {
		width: auto;
		min-width: 7rem;
	}

	.create {
		margin-top: 1.15rem;
	}

	h3 {
		margin: 0 0 0.3rem;
		font-size: 0.92rem;
		font-weight: 650;
	}

	.hint {
		margin: 0 0 0.75rem;
		font-size: 0.88rem;
	}

	.fields {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.75rem;
		margin-bottom: 0.85rem;
	}

	.invite-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.55rem;
		align-items: center;
	}

	.link {
		font-size: 0.78rem;
		word-break: break-all;
		max-width: 100%;
	}

	.note {
		margin-top: 1.1rem;
	}
</style>
