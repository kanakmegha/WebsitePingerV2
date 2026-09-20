<script lang="ts">
	import { onMount } from 'svelte';
	import {
		monitorsStore,
		filteredMonitors,
		monitorStats,
		searchQuery,
		statusFilter,
		typeFilter
	} from '$lib/stores/monitors';
	import { fetchTenantSettings } from '$lib/services/api';
	import type { TenantSettings } from '$lib/types';
	import MonitorCard from '$lib/components/MonitorCard.svelte';
	import MonitorTable from '$lib/components/MonitorTable.svelte';
	import {
		Activity,
		ShieldCheck,
		AlertTriangle,
		ShieldAlert,
		Search,
		Filter,
		Grid,
		List
	} from '@lucide/svelte';

	let viewMode = $state<'grid' | 'table'>('table');
	let settings = $state<TenantSettings | null>(null);

	onMount(() => {
		monitorsStore.load();
		fetchTenantSettings().then((s) => (settings = s));

		const interval = setInterval(() => {
			monitorsStore.load();
		}, 10000);

		return () => clearInterval(interval);
	});

	let sslThreshold = $derived(settings?.ssl_min_expiry_days ?? 30);
	let domainThreshold = $derived(settings?.domain_min_expiry_days ?? 30);

	let warningCount = $derived(
		$monitorsStore.filter((m) => {
			const sslWarn = m.ssl_days_remaining !== undefined && m.ssl_days_remaining !== null && m.ssl_days_remaining <= sslThreshold;
			return sslWarn;
		}).length
	);
</script>

<svelte:head>
	<title>Pinger Dashboard | Website Monitoring SaaS</title>
</svelte:head>

<div class="space-y-8">
	<!-- Top Metric Summary Cards -->
	<div class="grid grid-cols-2 gap-3 sm:gap-4 lg:grid-cols-4">
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 sm:p-5 backdrop-blur dark:bg-slate-900/80 shadow-lg">
			<div class="flex items-center justify-between">
				<span class="text-[10px] sm:text-xs font-semibold uppercase tracking-wider text-slate-400">Total Monitors</span>
				<div class="flex h-7 w-7 sm:h-8 sm:w-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
					<Activity class="h-4 w-4" />
				</div>
			</div>
			<p class="mt-2 sm:mt-3 font-mono text-2xl sm:text-3xl font-extrabold text-white">{$monitorStats.total}</p>
			<p class="mt-1 text-[10px] sm:text-xs text-slate-400 truncate">Active monitoring targets</p>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 sm:p-5 backdrop-blur dark:bg-slate-900/80 shadow-lg">
			<div class="flex items-center justify-between">
				<span class="text-[10px] sm:text-xs font-semibold uppercase tracking-wider text-slate-400">Healthy</span>
				<div class="flex h-7 w-7 sm:h-8 sm:w-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-400">
					<ShieldCheck class="h-4 w-4" />
				</div>
			</div>
			<p class="mt-2 sm:mt-3 font-mono text-2xl sm:text-3xl font-extrabold text-emerald-400">{$monitorStats.up}</p>
			<p class="mt-1 text-[10px] sm:text-xs text-slate-400 truncate">Operational without issues</p>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 sm:p-5 backdrop-blur dark:bg-slate-900/80 shadow-lg">
			<div class="flex items-center justify-between">
				<span class="text-[10px] sm:text-xs font-semibold uppercase tracking-wider text-slate-400">Incidents</span>
				<div class="flex h-7 w-7 sm:h-8 sm:w-8 items-center justify-center rounded-lg bg-red-500/10 text-red-400">
					<AlertTriangle class="h-4 w-4" />
				</div>
			</div>
			<p class="mt-2 sm:mt-3 font-mono text-2xl sm:text-3xl font-extrabold {$monitorStats.down > 0 ? 'text-red-400' : 'text-slate-100'}">
				{$monitorStats.down}
			</p>
			<p class="mt-1 text-[10px] sm:text-xs text-slate-400 truncate">Failing inspection checks</p>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 sm:p-5 backdrop-blur dark:bg-slate-900/80 shadow-lg">
			<div class="flex items-center justify-between">
				<span class="text-[10px] sm:text-xs font-semibold uppercase tracking-wider text-slate-400">Warnings</span>
				<div class="flex h-7 w-7 sm:h-8 sm:w-8 items-center justify-center rounded-lg bg-amber-500/10 text-amber-400">
					<ShieldAlert class="h-4 w-4" />
				</div>
			</div>
			<p class="mt-2 sm:mt-3 font-mono text-2xl sm:text-3xl font-extrabold {warningCount > 0 ? 'text-amber-400' : 'text-slate-100'}">
				{warningCount}
			</p>
			<p class="mt-1 text-[10px] sm:text-xs text-slate-400 truncate">Expiring within {sslThreshold}d</p>
		</div>
	</div>

	<!-- Toolbar -->
	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between border-y border-slate-800/80 py-4">
		<div class="relative w-full sm:max-w-xs md:max-w-md">
			<Search class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
			<input
				type="text"
				bind:value={$searchQuery}
				placeholder="Search by name, URL, domain..."
				class="w-full rounded-xl border border-slate-800 bg-slate-900/80 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 placeholder-slate-500 focus:border-emerald-500 focus:outline-none focus:ring-1 focus:ring-emerald-500 min-h-[40px]"
			/>
		</div>

		<div class="flex flex-wrap items-center justify-between sm:justify-end gap-2.5">
			<div class="flex items-center gap-2">
				<div class="flex items-center gap-1.5 text-xs text-slate-400">
					<Filter class="h-3.5 w-3.5 text-slate-500 shrink-0" />
					<select
						bind:value={$statusFilter}
						class="rounded-lg border border-slate-800 bg-slate-900 px-2.5 py-2 text-xs font-semibold text-slate-200 focus:border-emerald-500 focus:outline-none"
					>
						<option value="all">All Statuses</option>
						<option value="up">Healthy (UP)</option>
						<option value="down">Down (FAILING)</option>
						<option value="warning">Warning (Expiry)</option>
					</select>
				</div>

				<select
					bind:value={$typeFilter}
					class="rounded-lg border border-slate-800 bg-slate-900 px-2.5 py-2 text-xs font-semibold text-slate-200 focus:border-emerald-500 focus:outline-none"
				>
					<option value="all">All Types</option>
					<option value="http">HTTP Uptime</option>
					<option value="ssl">SSL Certificate</option>
					<option value="dns">DNS Records</option>
					<option value="whois">WHOIS Domain</option>
				</select>
			</div>

			<div class="flex items-center rounded-lg border border-slate-800 bg-slate-900/80 p-1">
				<button
					onclick={() => (viewMode = 'table')}
					class="flex h-8 w-8 items-center justify-center rounded-md transition {viewMode === 'table' ? 'bg-emerald-500/20 text-emerald-400 font-bold' : 'text-slate-400 hover:text-slate-200'}"
					aria-label="Table View"
				>
					<List class="h-4 w-4" />
				</button>
				<button
					onclick={() => (viewMode = 'grid')}
					class="flex h-8 w-8 items-center justify-center rounded-md transition {viewMode === 'grid' ? 'bg-emerald-500/20 text-emerald-400 font-bold' : 'text-slate-400 hover:text-slate-200'}"
					aria-label="Grid View"
				>
					<Grid class="h-4 w-4" />
				</button>
			</div>
		</div>
	</div>

	<!-- Monitor List Display -->
	{#if viewMode === 'table'}
		<MonitorTable monitors={$filteredMonitors} />
	{:else}
		<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each $filteredMonitors as monitor (monitor.id)}
				<MonitorCard {monitor} />
			{/each}
		</div>
	{/if}
</div>
