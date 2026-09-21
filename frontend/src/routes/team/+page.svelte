<script lang="ts">
	import { onMount } from 'svelte';
	import {
		inviteUser,
		fetchTenantMembers,
		fetchPendingInvites,
		cancelInvite
	} from '$lib/services/api';
	import { tenantStore } from '$lib/stores/tenant';
	import { toastStore } from '$lib/stores/toast';
	import type { TenantMember, PendingInvite } from '$lib/types';
	import {
		Users,
		UserPlus,
		Mail,
		ShieldCheck,
		Clock,
		Trash2,
		RotateCw,
		Loader2,
		AlertCircle,
		CheckCircle2,
		Building2
	} from '@lucide/svelte';

	let members = $state<TenantMember[]>([]);
	let invites = $state<PendingInvite[]>([]);
	let loadingData = $state(true);

	// Invite Form State
	let inviteEmail = $state('');
	let inviteRole = $state<'member' | 'admin'>('member');
	let sendingInvite = $state(false);
	let formError = $state('');
	let actionInProgressId = $state<string | null>(null);

	let selectedTenantName = $derived(
		$tenantStore.tenants.find((t) => t.id === $tenantStore.selectedTenantId)?.name || 'Organization'
	);

	async function loadTeamData() {
		loadingData = true;
		try {
			const [membersData, invitesData] = await Promise.all([
				fetchTenantMembers(),
				fetchPendingInvites()
			]);
			members = membersData;
			invites = invitesData;
		} catch (err: any) {
			toastStore.show('Failed to load team details', 'error');
		} finally {
			loadingData = false;
		}
	}

	onMount(() => {
		loadTeamData();
	});

	// Re-load when user switches active tenant
	$effect(() => {
		const tid = $tenantStore.selectedTenantId;
		if (tid) {
			loadTeamData();
		}
	});

	async function handleSendInvite(e: Event) {
		e.preventDefault();
		if (!inviteEmail) {
			formError = 'Please enter a recipient email address.';
			return;
		}

		sendingInvite = true;
		formError = '';

		try {
			const newInvite = await inviteUser(inviteEmail, inviteRole);
			toastStore.show(`Invitation sent to ${inviteEmail}`, 'success');
			invites = [newInvite, ...invites];
			inviteEmail = '';
			inviteRole = 'member';
		} catch (err: any) {
			formError = err.message || 'Failed to send invitation.';
		} finally {
			sendingInvite = false;
		}
	}

	async function handleResendInvite(inv: PendingInvite) {
		actionInProgressId = inv.id;
		try {
			await inviteUser(inv.email, inv.role);
			toastStore.show(`Invitation resent to ${inv.email}`, 'success');
			await loadTeamData();
		} catch (err: any) {
			toastStore.show(err.message || 'Failed to resend invite', 'error');
		} finally {
			actionInProgressId = null;
		}
	}

	async function handleCancelInvite(invId: string, email: string) {
		actionInProgressId = invId;
		try {
			await cancelInvite(invId);
			toastStore.show(`Cancelled invitation for ${email}`, 'info');
			invites = invites.filter((i) => i.id !== invId);
		} catch (err: any) {
			toastStore.show(err.message || 'Failed to cancel invitation', 'error');
		} finally {
			actionInProgressId = null;
		}
	}
</script>

<svelte:head>
	<title>Team & Members | Pinger</title>
</svelte:head>

<div class="space-y-8 pb-12">
	<!-- Page Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-slate-800 pb-5">
		<div>
			<div class="flex items-center gap-2">
				<h1 class="text-2xl font-extrabold text-white">Team Management</h1>
				<span class="inline-flex items-center gap-1 rounded-full border border-emerald-500/30 bg-emerald-500/10 px-2.5 py-0.5 text-xs font-bold text-emerald-400">
					<Building2 class="h-3.5 w-3.5" />
					{selectedTenantName}
				</span>
			</div>
			<p class="text-xs text-slate-400 mt-1">Manage team access, invite new collaborators, and view active workspace members.</p>
		</div>

		<button
			onclick={loadTeamData}
			disabled={loadingData}
			class="inline-flex items-center gap-2 rounded-xl border border-slate-800 bg-slate-900 px-3.5 py-2 text-xs font-semibold text-slate-300 hover:border-slate-700 hover:text-white transition active:scale-95 disabled:opacity-50"
		>
			<RotateCw class="h-3.5 w-3.5 {loadingData ? 'animate-spin text-emerald-400' : ''}" />
			Refresh
		</button>
	</div>

	<!-- Grid Container -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
		<!-- Section A: Invite User Form -->
		<div class="lg:col-span-1 space-y-6">
			<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
				<div class="flex items-center gap-3 border-b border-slate-800/80 pb-4">
					<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400">
						<UserPlus class="h-5 w-5" />
					</div>
					<div>
						<h2 class="text-sm font-bold text-white">Invite Collaborator</h2>
						<p class="text-[11px] text-slate-400">Send an email invitation link</p>
					</div>
				</div>

				<form onsubmit={handleSendInvite} class="space-y-4">
					{#if formError}
						<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
							<AlertCircle class="h-4 w-4 shrink-0" />
							<span>{formError}</span>
						</div>
					{/if}

					<div>
						<label for="invite-email" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
							Email Address
						</label>
						<div class="relative mt-2">
							<Mail class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
							<input
								id="invite-email"
								type="email"
								required
								bind:value={inviteEmail}
								placeholder="colleague@company.com"
								class="w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
							/>
						</div>
					</div>

					<div>
						<label for="invite-role" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
							Assigned Role
						</label>
						<div class="relative mt-2">
							<ShieldCheck class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
							<select
								id="invite-role"
								bind:value={inviteRole}
								class="w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 focus:border-emerald-500 focus:outline-none appearance-none"
							>
								<option value="member">Member (View & Monitor Access)</option>
								<option value="admin">Admin (Full Management Rights)</option>
							</select>
						</div>
					</div>

					<button
						type="submit"
						disabled={sendingInvite}
						class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50 mt-2"
					>
						{#if sendingInvite}
							<Loader2 class="h-4 w-4 animate-spin" />
							Sending Email...
						{:else}
							<UserPlus class="h-4 w-4" />
							Send Invitation
						{/if}
					</button>
				</form>
			</div>
		</div>

		<!-- Right Column: Members & Pending Invites -->
		<div class="lg:col-span-2 space-y-8">
			<!-- Section B: Current Members List -->
			<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
				<div class="flex items-center justify-between border-b border-slate-800/80 pb-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">
							<Users class="h-5 w-5" />
						</div>
						<div>
							<h2 class="text-sm font-bold text-white">Active Team Members</h2>
							<p class="text-[11px] text-slate-400">Current users with workspace access</p>
						</div>
					</div>
					<span class="rounded-full bg-slate-800 px-3 py-1 text-xs font-bold text-slate-300">
						{members.length} {members.length === 1 ? 'member' : 'members'}
					</span>
				</div>

				{#if loadingData}
					<div class="flex items-center justify-center py-8 space-y-2">
						<Loader2 class="h-6 w-6 animate-spin text-emerald-400" />
					</div>
				{:else if members.length === 0}
					<div class="py-8 text-center text-xs text-slate-500">No active members found.</div>
				{:else}
					<div class="divide-y divide-slate-800/60">
						{#each members as m (m.id)}
							<div class="flex items-center justify-between py-3.5 first:pt-0 last:pb-0">
								<div class="flex items-center gap-3">
									<div class="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-tr from-slate-800 to-slate-700 text-xs font-bold text-emerald-400 border border-slate-700">
										{m.email.substring(0, 2).toUpperCase()}
									</div>
									<div>
										<p class="text-xs font-bold text-white">{m.email}</p>
										<p class="text-[10px] text-slate-500">Joined {new Date(m.created_at).toLocaleDateString()}</p>
									</div>
								</div>
								<span class="inline-flex items-center rounded-full bg-slate-800 px-2.5 py-1 text-[10px] font-extrabold uppercase tracking-wider text-slate-300 border border-slate-700">
									{m.role}
								</span>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Section C: Pending Invites List -->
			<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
				<div class="flex items-center justify-between border-b border-slate-800/80 pb-4">
					<div class="flex items-center gap-3">
						<div class="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400">
							<Clock class="h-5 w-5" />
						</div>
						<div>
							<h2 class="text-sm font-bold text-white">Pending Invitations</h2>
							<p class="text-[11px] text-slate-400">Outstanding emails awaiting acceptance</p>
						</div>
					</div>
					<span class="rounded-full bg-slate-800 px-3 py-1 text-xs font-bold text-slate-300">
						{invites.length} pending
					</span>
				</div>

				{#if loadingData}
					<div class="flex items-center justify-center py-8">
						<Loader2 class="h-6 w-6 animate-spin text-amber-400" />
					</div>
				{:else if invites.length === 0}
					<div class="py-8 text-center text-xs text-slate-500">No pending invitations.</div>
				{:else}
					<div class="divide-y divide-slate-800/60">
						{#each invites as inv (inv.id)}
							<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 py-3.5 first:pt-0 last:pb-0">
								<div class="space-y-1">
									<div class="flex items-center gap-2">
										<p class="text-xs font-bold text-white">{inv.email}</p>
										<span class="rounded-full bg-amber-500/10 border border-amber-500/20 px-2 py-0.5 text-[9px] font-bold text-amber-400 uppercase">
											{inv.role}
										</span>
									</div>
									<p class="text-[10px] text-slate-500">
										Sent {new Date(inv.created_at).toLocaleDateString()} • Expires {new Date(inv.expires_at).toLocaleDateString()}
									</p>
								</div>

								<div class="flex items-center gap-2">
									<button
										onclick={() => handleResendInvite(inv)}
										disabled={actionInProgressId === inv.id}
										class="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-950 px-3 py-1.5 text-xs font-medium text-slate-300 hover:border-slate-700 hover:text-white transition disabled:opacity-50"
									>
										<RotateCw class="h-3 w-3 text-emerald-400 {actionInProgressId === inv.id ? 'animate-spin' : ''}" />
										Resend
									</button>
									<button
										onclick={() => handleCancelInvite(inv.id, inv.email)}
										disabled={actionInProgressId === inv.id}
										class="inline-flex items-center gap-1.5 rounded-lg border border-red-500/20 bg-red-500/10 px-3 py-1.5 text-xs font-medium text-red-400 hover:bg-red-500/20 transition disabled:opacity-50"
									>
										<Trash2 class="h-3 w-3" />
										Cancel
									</button>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
