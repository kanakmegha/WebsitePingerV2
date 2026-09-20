<script lang="ts">
	import { onMount } from 'svelte';
	import type { MonitorCheck, Monitor, TenantSettings } from '$lib/types';
	import { fetchTenantSettings } from '$lib/services/api';

	let {
		checks = [],
		monitor = null,
		sslCheckRecords = [],
		dnsCheckRecords = []
	}: {
		checks?: MonitorCheck[];
		monitor?: Monitor | null;
		sslCheckRecords?: MonitorCheck[];
		dnsCheckRecords?: MonitorCheck[];
	} = $props();

	let settings = $state<TenantSettings | null>(null);

	onMount(async () => {
		settings = await fetchTenantSettings();
	});

	// Group checks by type, using dedicated check records if available
	let httpChecks = $derived(checks.filter((c) => c.check_type === 'http'));
	let dnsChecks = $derived(dnsCheckRecords.length > 0 ? dnsCheckRecords : checks.filter((c) => c.check_type === 'dns'));
	let sslChecks = $derived(sslCheckRecords.length > 0 ? sslCheckRecords : checks.filter((c) => c.check_type === 'ssl'));

	function getLastCheckedText(typeChecks: MonitorCheck[]): string {
		if (typeChecks.length === 0) return 'Never checked';
		return new Date(typeChecks[0].checked_at).toLocaleString();
	}

	let httpLastChecked = $derived(getLastCheckedText(httpChecks));
	let dnsLastChecked = $derived(getLastCheckedText(dnsChecks));
	let sslLastChecked = $derived(getLastCheckedText(sslChecks));

	// Find interval for each check type (CheckConfig -> TenantSettings -> Defaults)
	function getInterval(type: string): number {
		const cfg = monitor?.check_configs?.find((c) => c.check_type === type);
		if (cfg && cfg.interval_seconds) return cfg.interval_seconds;
		if (settings) {
			switch (type) {
				case 'http': return settings.http_interval_seconds;
				case 'dns': return settings.dns_interval_seconds;
				case 'ssl': return settings.ssl_interval_seconds;
			}
		}
		switch (type) {
			case 'http': return 60;
			case 'dns': return 300;
			case 'ssl': return 3600;
			default: return 60;
		}
	}

	function formatInterval(seconds: number): string {
		if (seconds >= 86400) return `${Math.round(seconds / 86400)}d`;
		if (seconds >= 3600) return `${Math.round(seconds / 3600)}h`;
		if (seconds >= 60) return `${Math.round(seconds / 60)}m`;
		return `${seconds}s`;
	}

	function generateBlocks(typeChecks: MonitorCheck[]) {
		if (typeChecks.length === 0) return [];
		const slice = typeChecks.slice(0, 40).reverse();
		return slice.map((c) => ({
			id: c.id,
			status: c.status === 'success' ? ('up' as const) : ('down' as const),
			label: `${new Date(c.checked_at).toLocaleTimeString()} - ${c.status === 'success' ? 'Operational' : c.error || 'Failed'}`
		}));
	}

	let httpBlocks = $derived(generateBlocks(httpChecks));
	let dnsBlocks = $derived(generateBlocks(dnsChecks));
	let sslBlocks = $derived(generateBlocks(sslChecks));

	let httpUptime = $derived(
		httpChecks.length > 0
			? ((httpChecks.filter((c) => c.status === 'success').length / httpChecks.length) * 100).toFixed(1)
			: '100.0'
	);
</script>

<div class="space-y-4 pt-2">
	<div class="flex items-center justify-between text-xs text-slate-400">
		<span class="font-semibold text-slate-300">Check Type Status Timelines</span>
		<span class="font-medium text-emerald-400">{httpUptime}% HTTP Uptime</span>
	</div>

	<!-- HTTP Timeline (Top - aligned with graph interval) -->
	<div class="space-y-1">
		<div class="flex items-center justify-between text-[11px] font-mono text-slate-400">
			<div class="flex items-center gap-2">
				<span class="font-bold text-emerald-400">HTTP</span>
				<span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] text-slate-300 border border-slate-700">Interval: {formatInterval(getInterval('http'))}</span>
			</div>
			<div class="flex items-center gap-3 text-[10px]">
				<span class="text-slate-400">Last checked: <strong class="text-slate-200">{httpLastChecked}</strong></span>
				<span class="text-slate-500">•</span>
				<span class="text-slate-500">{httpBlocks.length} check{httpBlocks.length === 1 ? '' : 's'} recorded</span>
			</div>
		</div>
		<div class="flex h-5 w-full items-center gap-1 rounded-md border border-slate-800 bg-slate-900/60 p-1 backdrop-blur">
			{#if httpBlocks.length === 0}
				<div class="flex h-full w-full items-center justify-center text-[10px] text-slate-500 font-mono">No HTTP checks recorded yet</div>
			{:else}
				{#each httpBlocks as block (block.id)}
					<div
						title={block.label}
						class="h-full flex-1 rounded-xs transition-all hover:scale-y-125 hover:brightness-125 {block.status === 'up'
							? 'bg-emerald-500/80 shadow-xs shadow-emerald-500/20'
							: 'bg-red-500 shadow-xs shadow-red-500/20'}"
					></div>
				{/each}
			{/if}
		</div>
	</div>

	<!-- DNS Timeline -->
	<div class="space-y-1">
		<div class="flex items-center justify-between text-[11px] font-mono text-slate-400">
			<div class="flex items-center gap-2">
				<span class="font-bold text-blue-400">DNS</span>
				<span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] text-slate-300 border border-slate-700">Interval: {formatInterval(getInterval('dns'))}</span>
			</div>
			<div class="flex items-center gap-3 text-[10px]">
				<span class="text-slate-400">Last checked: <strong class="text-slate-200">{dnsLastChecked}</strong></span>
				<span class="text-slate-500">•</span>
				<span class="text-slate-500">{dnsBlocks.length} check{dnsBlocks.length === 1 ? '' : 's'} recorded</span>
			</div>
		</div>
		<div class="flex h-5 w-full items-center gap-1 rounded-md border border-slate-800 bg-slate-900/60 p-1 backdrop-blur">
			{#if dnsBlocks.length === 0}
				<div class="flex h-full w-full items-center justify-center text-[10px] text-slate-500 font-mono">No DNS checks recorded yet</div>
			{:else}
				{#each dnsBlocks as block (block.id)}
					<div
						title={block.label}
						class="h-full flex-1 rounded-xs transition-all hover:scale-y-125 hover:brightness-125 {block.status === 'up'
							? 'bg-blue-500/80 shadow-xs shadow-blue-500/20'
							: 'bg-red-500 shadow-xs shadow-red-500/20'}"
					></div>
				{/each}
			{/if}
		</div>
	</div>

	<!-- SSL Timeline -->
	<div class="space-y-1">
		<div class="flex items-center justify-between text-[11px] font-mono text-slate-400">
			<div class="flex items-center gap-2">
				<span class="font-bold text-teal-400">SSL</span>
				<span class="rounded bg-slate-800 px-1.5 py-0.5 text-[10px] text-slate-300 border border-slate-700">Interval: {formatInterval(getInterval('ssl'))}</span>
			</div>
			<div class="flex items-center gap-3 text-[10px]">
				<span class="text-slate-400">Last checked: <strong class="text-slate-200">{sslLastChecked}</strong></span>
				<span class="text-slate-500">•</span>
				<span class="text-slate-500">{sslBlocks.length} check{sslBlocks.length === 1 ? '' : 's'} recorded</span>
			</div>
		</div>
		<div class="flex h-5 w-full items-center gap-1 rounded-md border border-slate-800 bg-slate-900/60 p-1 backdrop-blur">
			{#if sslBlocks.length === 0}
				<div class="flex h-full w-full items-center justify-center text-[10px] text-slate-500 font-mono">No SSL checks recorded yet</div>
			{:else}
				{#each sslBlocks as block (block.id)}
					<div
						title={block.label}
						class="h-full flex-1 rounded-xs transition-all hover:scale-y-125 hover:brightness-125 {block.status === 'up'
							? 'bg-teal-500/80 shadow-xs shadow-teal-500/20'
							: 'bg-red-500 shadow-xs shadow-red-500/20'}"
					></div>
				{/each}
			{/if}
		</div>
	</div>
</div>
