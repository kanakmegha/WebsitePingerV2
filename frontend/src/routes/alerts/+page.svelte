<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { alertStore, alertFilter, filteredAlerts, unreadCount } from '$lib/stores/alerts';
	import AlertList from '$lib/components/AlertList.svelte';
	import { CheckCheck, RefreshCw } from '@lucide/svelte';

	let refreshInterval: any;
	let isRefreshing = $state(false);

	async function refreshAlerts() {
		isRefreshing = true;
		await alertStore.loadAlerts();
		isRefreshing = false;
	}

	onMount(() => {
		alertStore.loadAlerts();
		refreshInterval = setInterval(() => {
			alertStore.loadAlerts();
		}, 30000);
	});

	onDestroy(() => {
		if (refreshInterval) clearInterval(refreshInterval);
	});
</script>

<svelte:head>
	<title>Alerts Feed | Pinger</title>
</svelte:head>

<div class="space-y-8">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-slate-800 pb-6">
		<div>
			<h1 class="text-2xl font-extrabold text-white">Alert Events & Incidents</h1>
			<p class="mt-1 text-xs text-slate-400">State-driven alert transitions (Incidents & Recoveries)</p>
		</div>

		<div class="flex items-center gap-3">
			{#if $unreadCount > 0}
				<button
					onclick={() => alertStore.markAllRead()}
					class="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:border-slate-700 hover:text-white transition"
				>
					<CheckCheck class="h-3.5 w-3.5 text-emerald-400" />
					Mark All Read
				</button>
			{/if}

			<button
				onclick={refreshAlerts}
				disabled={isRefreshing}
				class="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3 py-1.5 text-xs font-semibold text-slate-300 hover:border-slate-700 hover:text-white transition disabled:opacity-50"
				title="Refresh Alerts"
			>
				<RefreshCw class="h-3.5 w-3.5 text-slate-400 {isRefreshing ? 'animate-spin' : ''}" />
				Refresh
			</button>

			<div class="flex items-center rounded-lg border border-slate-800 bg-slate-900 p-1 text-xs font-semibold">
				<button
					onclick={() => alertFilter.set('all')}
					class="rounded-md px-3.5 py-1.5 transition {$alertFilter === 'all' ? 'bg-emerald-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					All
				</button>
				<button
					onclick={() => alertFilter.set('unread')}
					class="rounded-md px-3.5 py-1.5 transition {$alertFilter === 'unread' ? 'bg-amber-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					Unread {$unreadCount > 0 ? `(${$unreadCount})` : ''}
				</button>
				<button
					onclick={() => alertFilter.set('incident')}
					class="rounded-md px-3.5 py-1.5 transition {$alertFilter === 'incident' ? 'bg-red-500 text-white font-bold' : 'text-slate-400 hover:text-white'}"
				>
					Incidents
				</button>
				<button
					onclick={() => alertFilter.set('recovery')}
					class="rounded-md px-3.5 py-1.5 transition {$alertFilter === 'recovery' ? 'bg-emerald-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					Recoveries
				</button>
				<button
					onclick={() => alertFilter.set('invites')}
					class="rounded-md px-3.5 py-1.5 transition {$alertFilter === 'invites' ? 'bg-cyan-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					Invites
				</button>
			</div>
		</div>
	</div>

	<AlertList alerts={$filteredAlerts} />
</div>
