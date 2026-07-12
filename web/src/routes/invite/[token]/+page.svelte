<script lang="ts">
	import { acceptInvite, getInvite, type InviteInfo } from '$lib/api';
	import ErrorBanner from '$lib/ErrorBanner.svelte';
	import Skeleton from '$lib/Skeleton.svelte';
	import { isLoggedIn } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let info = $state<InviteInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let accepting = $state(false);

	const token = $derived($page.params.token ?? '');

	async function load() {
		if (!token) return;
		loading = true;
		try {
			info = await getInvite(token);
			error = '';
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
			info = null;
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		void load();
	});

	async function onAccept() {
		if (!isLoggedIn()) {
			goto(`/login?next=${encodeURIComponent(`/invite/${token}`)}`);
			return;
		}
		accepting = true;
		try {
			const res = await acceptInvite(token);
			goto(`/t/${res.team.slug}`);
		} catch (err) {
			error = err instanceof Error ? err.message : String(err);
		} finally {
			accepting = false;
		}
	}
</script>

<p class="crumb"><a href="/">Home</a> / Invite</p>
<h1 class="brand-font">Team invite</h1>

{#if error}
	<ErrorBanner message={error} onretry={load} />
{:else if loading}
	<Skeleton lines={3} />
{:else if info}
	<section class="panel">
		<p>
			You’ve been invited to <strong>{info.team.name}</strong>
			<span class="muted mono">/{info.team.slug}</span> as <strong>{info.role}</strong>.
		</p>
		{#if info.expired}
			<p class="error">This invite is expired or already used.</p>
		{:else}
			<p class="muted expires">Expires {new Date(info.expiresAt).toLocaleString()}</p>
			<button type="button" onclick={onAccept} disabled={accepting}>
				{accepting ? 'Joining…' : isLoggedIn() ? 'Join team' : 'Log in to join'}
			</button>
			{#if !isLoggedIn()}
				<p class="muted tip">
					No account? <a href={`/signup?next=${encodeURIComponent(`/invite/${token}`)}`}>Sign up</a> first,
					then return here.
				</p>
			{/if}
		{/if}
	</section>
{/if}

<style>
	h1 {
		margin: 0 0 1rem;
		letter-spacing: -0.035em;
		font-size: clamp(1.65rem, 3.5vw, 2.2rem);
	}

	.panel p {
		margin: 0 0 0.85rem;
		line-height: 1.5;
	}

	.expires {
		font-size: 0.88rem;
	}

	.tip {
		margin-top: 0.85rem;
		font-size: 0.9rem;
	}

	.tip a {
		color: var(--accent);
	}
</style>
