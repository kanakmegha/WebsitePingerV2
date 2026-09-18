<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { Monitor, MonitorCheck } from '$lib/types';
	import { fetchMonitorById, fetchChecks, MOCK_MONITORS } from '$lib/services/api';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import UptimeChart from '$lib/components/UptimeChart.svelte';
	import StatusTimeline from '$lib/components/StatusTimeline.svelte';
	import CheckTable from '$lib/components/CheckTable.svelte';
	import {
		Globe,
		ShieldCheck,
		Server,
		MailCheck,
		ArrowLeft,
		ExternalLink
	} from '@lucide/svelte';

	let monitorId = $derived($page.params.id);
	let monitor = $state<Monitor>(MOCK_MONITORS[0]);
	let checks = $state<MonitorCheck[]>([]);
	let period = $state<'24h' | '7d' | '30d'>('24h');

	onMount(async () => {
		if (monitorId) {
			const fetchedMon = await fetchMonitorById(monitorId);
			if (fetchedMon) {
				monitor = fetchedMon;
			} else {
				const found = MOCK_MONITORS.find((m) => m.id === monitorId);
				if (found) monitor = found;
			}
			checks = await fetchChecks(monitorId);
		}
	});
</script>

<svelte:head>
	<title>{monitor.name} | Pinger Details</title>
</svelte:head>

<div class="space-y-8">
	<div>
		<a
			href="/"
			class="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-400 hover:text-white transition mb-4"
		>
			<ArrowLeft class="h-4 w-4" />
			Back to Dashboard
		</a>

		<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-slate-800 pb-6">
			<div>
				<div class="flex items-center gap-3">
					<h1 class="text-2xl font-extrabold text-white">{monitor.name}</h1>
					<StatusBadge status={monitor.status} size="lg" />
				</div>
				<a
					href={monitor.url}
					target="_blank"
					rel="noreferrer"
					class="mt-1 inline-flex items-center gap-1 text-xs font-mono text-emerald-400 hover:underline"
				>
					{monitor.url}
					<ExternalLink class="h-3 w-3" />
				</a>
			</div>

			<div class="flex items-center gap-3 text-xs font-mono text-slate-400">
				<span class="rounded-lg border border-slate-800 bg-slate-900 px-3 py-1.5">
					Type: <strong class="uppercase text-slate-200">{monitor.type}</strong>
				</span>
				<span class="rounded-lg border border-slate-800 bg-slate-900 px-3 py-1.5">
					Interval: <strong class="text-slate-200">{monitor.interval_seconds}s</strong>
				</span>
			</div>
		</div>
	</div>

	<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-6">
		<div class="flex items-center justify-between">
			<div>
				<h2 class="text-lg font-bold text-slate-100">Uptime & Latency Trend</h2>
				<p class="text-xs text-slate-400">Real-time response latency metrics</p>
			</div>

			<div class="flex items-center rounded-lg border border-slate-800 bg-slate-950 p-1 text-xs font-semibold">
				<button
					onclick={() => (period = '24h')}
					class="rounded-md px-3 py-1.5 transition {period === '24h' ? 'bg-emerald-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					24 Hours
				</button>
				<button
					onclick={() => (period = '7d')}
					class="rounded-md px-3 py-1.5 transition {period === '7d' ? 'bg-emerald-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					7 Days
				</button>
				<button
					onclick={() => (period = '30d')}
					class="rounded-md px-3 py-1.5 transition {period === '30d' ? 'bg-emerald-500 text-slate-950 font-bold' : 'text-slate-400 hover:text-white'}"
				>
					30 Days
				</button>
			</div>
		</div>

		<UptimeChart {period} />

		<StatusTimeline checksCount={40} />
	</div>

	<div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<Globe class="h-4 w-4 text-emerald-400" />
				HTTP Inspection
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Status Code:</span>
					<span class="font-bold text-emerald-400">200 OK</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Latency:</span>
					<span class="text-slate-200">{monitor.avg_latency_ms || 42} ms</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Size:</span>
					<span class="text-slate-200">14.2 KB</span>
				</div>
			</div>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<ShieldCheck class="h-4 w-4 text-emerald-400" />
				SSL Certificate
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Validity:</span>
					<span class="font-bold text-emerald-400">Valid</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Days Left:</span>
					<span class="font-bold text-slate-200">{monitor.ssl_days_remaining || 180} days</span>
				</div>
				<div class="flex justify-between truncate">
					<span class="text-slate-400">Issuer:</span>
					<span class="truncate text-slate-300">Google Trust LLC</span>
				</div>
			</div>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<Server class="h-4 w-4 text-blue-400" />
				DNS Records
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">A Record:</span>
					<span class="text-slate-200">142.250.190.46</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">AAAA Record:</span>
					<span class="text-slate-200">2404:6800::200e</span>
				</div>
				<div class="flex justify-between truncate">
					<span class="text-slate-400">MX Record:</span>
					<span class="truncate text-slate-300">smtp.google.com</span>
				</div>
			</div>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<MailCheck class="h-4 w-4 text-purple-400" />
				Email Security
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">SPF Record:</span>
					<span class="font-bold text-emerald-400">Valid</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">DMARC Policy:</span>
					<span class="font-bold text-emerald-400">Valid</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">DKIM Check:</span>
					<span class="text-emerald-400">Verified</span>
				</div>
			</div>
		</div>
	</div>

	<CheckTable {checks} />
</div>
