<script lang="ts">
	import { goto } from '$app/navigation';
	import { createOrganization, joinOrganization } from '$lib/services/api';
	import { toastStore } from '$lib/stores/toast';
	import { Building2, Plus, UserCheck, AlertCircle, ArrowRight } from '@lucide/svelte';

	let activeTab = $state<'create' | 'join'>('create');
	let orgName = $state('');
	let joinTenantId = $state('');
	let loading = $state(false);
	let errorMessage = $state('');

	async function handleCreate(e: Event) {
		e.preventDefault();
		if (!orgName.trim()) {
			errorMessage = 'Organization name is required';
			return;
		}

		loading = true;
		errorMessage = '';
		try {
			const org = await createOrganization(orgName.trim());
			toastStore.show(`Created & selected ${org.name}`, 'success');
			goto('/');
		} catch (err: any) {
			errorMessage = err.message || 'Failed to create organization';
		} finally {
			loading = false;
		}
	}

	async function handleJoin(e: Event) {
		e.preventDefault();
		if (!joinTenantId.trim()) {
			errorMessage = 'Valid Organization ID is required';
			return;
		}

		loading = true;
		errorMessage = '';
		try {
			const org = await joinOrganization(joinTenantId.trim());
			toastStore.show(`Joined ${org.name}`, 'success');
			goto('/');
		} catch (err: any) {
			errorMessage = err.message || 'Failed to join organization';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Organization Onboarding | Pinger</title>
</svelte:head>

<div class="mx-auto max-w-lg space-y-8 py-12">
	<div class="text-center space-y-2">
		<div class="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-tr from-emerald-500 to-cyan-500 shadow-xl shadow-emerald-500/20">
			<Building2 class="h-6 w-6 text-slate-950" />
		</div>
		<h1 class="text-2xl font-extrabold text-white">Organization Workspace Required</h1>
		<p class="text-xs text-slate-400">Create a new organization or join an existing workspace to get started.</p>
	</div>

	<!-- Tabs -->
	<div class="flex rounded-xl border border-slate-800 bg-slate-900/80 p-1.5 backdrop-blur text-xs font-semibold">
		<button
			onclick={() => { activeTab = 'create'; errorMessage = ''; }}
			class="flex-1 flex items-center justify-center gap-2 rounded-lg py-2.5 transition {activeTab === 'create' ? 'bg-emerald-500 text-slate-950 font-bold shadow-lg shadow-emerald-500/20' : 'text-slate-400 hover:text-white'}"
		>
			<Plus class="h-4 w-4" />
			Create New Organization
		</button>
		<button
			onclick={() => { activeTab = 'join'; errorMessage = ''; }}
			class="flex-1 flex items-center justify-center gap-2 rounded-lg py-2.5 transition {activeTab === 'join' ? 'bg-emerald-500 text-slate-950 font-bold shadow-lg shadow-emerald-500/20' : 'text-slate-400 hover:text-white'}"
		>
			<UserCheck class="h-4 w-4" />
			Join Existing Workspace
		</button>
	</div>

	{#if errorMessage}
		<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
			<AlertCircle class="h-4 w-4 shrink-0" />
			<span>{errorMessage}</span>
		</div>
	{/if}

	{#if activeTab === 'create'}
		<form onsubmit={handleCreate} class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
			<div>
				<label for="org-name" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
					Organization Name
				</label>
				<input
					id="org-name"
					type="text"
					required
					bind:value={orgName}
					placeholder="Acme Cloud Inc."
					class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 px-4 text-xs font-medium text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
				/>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
			>
				{loading ? 'Creating...' : 'Create & Open Workspace'}
				<ArrowRight class="h-4 w-4" />
			</button>
		</form>
	{:else}
		<form onsubmit={handleJoin} class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
			<div>
				<label for="join-id" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
					Organization ID (UUID)
				</label>
				<input
					id="join-id"
					type="text"
					required
					bind:value={joinTenantId}
					placeholder="e.g. 550e8400-e29b-41d4-a716-446655440000"
					class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 px-4 text-xs font-mono text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
				/>
				<p class="mt-1 text-[11px] text-slate-500">Ask your organization administrator for their Tenant ID key.</p>
			</div>

			<button
				type="submit"
				disabled={loading}
				class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
			>
				{loading ? 'Joining...' : 'Join Organization Workspace'}
				<ArrowRight class="h-4 w-4" />
			</button>
		</form>
	{/if}
</div>
