<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { Monitor, MonitorCheck } from '$lib/types';
	import { goto } from '$app/navigation';
	import { fetchMonitorById, fetchChecks, deleteMonitor } from '$lib/services/api';
	import { monitorsStore } from '$lib/stores/monitors';
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
		ExternalLink,
		Trash2
	} from '@lucide/svelte';

	let monitorId = $derived($page.params.id);
	let monitor = $state<Monitor | null>(null);
	let checks = $state<MonitorCheck[]>([]);
	let period = $state<'24h' | '7d' | '30d'>('24h');

	let sslCheckRecords = $state<MonitorCheck[]>([]);
	let dnsCheckRecords = $state<MonitorCheck[]>([]);
	let domainCheckRecord = $state<MonitorCheck | null>(null);
	let isDeleting = $state(false);

	async function handleDeleteMonitor() {
		if (!monitor) return;
		if (!confirm(`Are you sure you want to delete monitor "${monitor.name}"?`)) return;

		try {
			isDeleting = true;
			await deleteMonitor(monitor.id);
			monitorsStore.remove(monitor.id);
			goto('/');
		} catch (err: any) {
			alert(err.message || 'Failed to delete monitor');
		} finally {
			isDeleting = false;
		}
	}

	async function loadData() {
		if (!monitorId) return;
		const [fetchedMon, allChecks, sslChecks, dnsChecks, domainChecks] = await Promise.all([
			fetchMonitorById(monitorId),
			fetchChecks(monitorId),
			fetchChecks(monitorId, 'ssl'),
			fetchChecks(monitorId, 'dns'),
			fetchChecks(monitorId, 'domain')
		]);
		if (fetchedMon) {
			monitor = fetchedMon;
		}
		checks = allChecks;
		sslCheckRecords = sslChecks;
		dnsCheckRecords = dnsChecks;
		if (domainChecks.length > 0) domainCheckRecord = domainChecks[0];
	}

	onMount(() => {
		loadData();
		const interval = setInterval(loadData, 10000);
		return () => clearInterval(interval);
	});

	let latestHttpCheck = $derived(checks.find((c) => c.check_type === 'http' && c.http_result)?.http_result || monitor?.latest_http);
	let latestSslCheck = $derived(sslCheckRecords.find((c) => c.ssl_result)?.ssl_result || checks.find((c) => c.check_type === 'ssl' && c.ssl_result)?.ssl_result || monitor?.latest_ssl);
	let latestDnsCheck = $derived(dnsCheckRecords.find((c) => c.dns_result)?.dns_result || checks.find((c) => c.check_type === 'dns' && c.dns_result)?.dns_result || monitor?.latest_dns);
	let latestDomainCheck = $derived(domainCheckRecord?.domain_result || checks.find((c) => c.check_type === 'domain' && c.domain_result)?.domain_result || monitor?.latest_domain);

	let latestStatus = $derived(checks.length > 0 ? (checks[0].status === 'success' ? 'up' : 'down') : (monitor?.status || 'pending'));
</script>

<svelte:head>
	<title>{monitor?.name || 'Monitor Details'} | Pinger</title>
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

		{#if monitor}
			<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between border-b border-slate-800 pb-6">
				<div>
					<div class="flex items-center gap-3">
						<h1 class="text-2xl font-extrabold text-white">{monitor.name}</h1>
						<StatusBadge status={latestStatus} size="lg" />
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

				<div class="flex items-center gap-3">
					<button
						type="button"
						onclick={handleDeleteMonitor}
						disabled={isDeleting}
						class="inline-flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 px-3.5 py-2 text-xs font-semibold text-red-400 hover:bg-red-500/20 hover:border-red-500/50 transition disabled:opacity-50"
					>
						<Trash2 class="h-4 w-4" />
						{isDeleting ? 'Deleting...' : 'Delete Monitor'}
					</button>
				</div>
			</div>
		{:else}
			<div class="border-b border-slate-800 pb-6 animate-pulse">
				<div class="h-8 w-48 bg-slate-800 rounded"></div>
			</div>
		{/if}
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

		<UptimeChart {period} {checks} />

		<StatusTimeline {checks} {monitor} {sslCheckRecords} {dnsCheckRecords} />
	</div>

	<div class="grid grid-cols-1 gap-4 sm:gap-6 sm:grid-cols-2 lg:grid-cols-4">
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 sm:p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<Globe class="h-4 w-4 text-emerald-400" />
				HTTP Inspection
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Status Code:</span>
					<span class="font-bold {latestHttpCheck?.status_code === 200 ? 'text-emerald-400' : 'text-amber-400'}">
						{latestHttpCheck ? `${latestHttpCheck.status_code}` : 'N/A'}
					</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Latency:</span>
					<span class="text-slate-200">{latestHttpCheck ? `${latestHttpCheck.response_time_ms} ms` : 'N/A'}</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Size:</span>
					<span class="text-slate-200">{latestHttpCheck ? `${(latestHttpCheck.response_size_bytes / 1024).toFixed(1)} KB` : 'N/A'}</span>
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
					<span class="font-bold {latestSslCheck?.valid ? 'text-emerald-400' : 'text-red-400'}">
						{latestSslCheck ? (latestSslCheck.valid ? 'Valid' : 'Invalid') : 'N/A'}
					</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">Days Left:</span>
					<span class="font-bold text-slate-200">{latestSslCheck ? `${latestSslCheck.days_remaining} days` : 'N/A'}</span>
				</div>
				<div class="flex justify-between truncate">
					<span class="text-slate-400">Issuer:</span>
					<span class="truncate text-slate-300">{latestSslCheck ? latestSslCheck.issuer : 'N/A'}</span>
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
					<span class="text-slate-200 truncate">{latestDnsCheck?.a_records && latestDnsCheck.a_records.length > 0 ? latestDnsCheck.a_records[0] : 'None'}</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">AAAA Record:</span>
					<span class="text-slate-200 truncate">{latestDnsCheck?.aaaa_records && latestDnsCheck.aaaa_records.length > 0 ? latestDnsCheck.aaaa_records[0] : 'None'}</span>
				</div>
				<div class="flex justify-between truncate">
					<span class="text-slate-400">MX Record:</span>
					<span class="truncate text-slate-300">{latestDnsCheck?.mx_records && latestDnsCheck.mx_records.length > 0 ? latestDnsCheck.mx_records[0] : 'None'}</span>
				</div>
			</div>
		</div>

		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur shadow-lg">
			<div class="flex items-center gap-2 text-slate-300 font-bold text-sm">
				<MailCheck class="h-4 w-4 text-purple-400" />
				Email Security & WHOIS
			</div>
			<div class="mt-4 space-y-2 text-xs font-mono">
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">SPF Record:</span>
					<span class="font-bold {latestDnsCheck?.spf_valid ? 'text-emerald-400' : 'text-slate-500'}">
						{latestDnsCheck ? (latestDnsCheck.spf_valid ? 'Valid' : 'Missing') : 'N/A'}
					</span>
				</div>
				<div class="flex justify-between border-b border-slate-800/60 pb-1.5">
					<span class="text-slate-400">DMARC Policy:</span>
					<span class="font-bold {latestDnsCheck?.dmarc_valid ? 'text-emerald-400' : 'text-slate-500'}">
						{latestDnsCheck ? (latestDnsCheck.dmarc_valid ? 'Valid' : 'Missing') : 'N/A'}
					</span>
				</div>
				<div class="flex justify-between">
					<span class="text-slate-400">Domain Expiry:</span>
					<span class="text-slate-200">{latestDomainCheck ? `${latestDomainCheck.days_remaining} days` : 'N/A'}</span>
				</div>
			</div>
		</div>
	</div>

	<CheckTable {checks} />
</div>
