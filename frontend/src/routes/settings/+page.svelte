<script lang="ts">
	import { fetchTenantSettings, updateTenantSettings } from '$lib/services/api';
	import { toastStore } from '$lib/stores/toast';
	import type { TenantSettings } from '$lib/types';
	import { Mail, Webhook, MessageSquare, Save, CheckCircle2, Clock, ShieldAlert, AlertCircle } from '@lucide/svelte';

	let settings = $state<TenantSettings>({
		http_interval_seconds: 60,
		dns_interval_seconds: 300,
		ssl_interval_seconds: 3600,
		domain_interval_seconds: 86400,
		email_auth_interval_seconds: 300
	});

	let isLoading = $state(true);
	let isSaving = $state(false);
	let isSaved = $state(false);
	let errorMessage = $state('');

	// Notification channel local UI state
	let emailAlerts = $state(true);
	let emailAddress = $state('ops-alerts@company.com');
	let webhookAlerts = $state(true);
	let webhookUrl = $state('https://api.company.com/webhooks/pinger');
	let slackWebhook = $state('https://hooks.slack.com/services/T00/B00/XXXX');

	$effect(() => {
		fetchTenantSettings()
			.then((data) => {
				settings = data;
			})
			.finally(() => {
				isLoading = false;
			});
	});

	let validationErrors = $derived.by(() => {
		const errs: string[] = [];
		if (settings.http_interval_seconds < 30) errs.push('HTTP Uptime interval must be >= 30 seconds');
		if (settings.dns_interval_seconds < 300) errs.push('DNS Records interval must be >= 300 seconds (5 mins)');
		if (settings.ssl_interval_seconds < 3600) errs.push('SSL Expiry interval must be >= 3600 seconds (1 hour)');
		if (settings.domain_interval_seconds < 86400) errs.push('Domain Expiry interval must be >= 86400 seconds (24 hours)');
		if (settings.email_auth_interval_seconds < 300) errs.push('Email Auth interval must be >= 300 seconds (5 mins)');
		return errs;
	});

	async function handleSave(e: Event) {
		e.preventDefault();
		if (validationErrors.length > 0) {
			errorMessage = validationErrors[0];
			return;
		}

		isSaving = true;
		errorMessage = '';

		try {
			const updated = await updateTenantSettings(settings);
			settings = updated;
			isSaved = true;
			toastStore.show('Monitoring intervals & settings updated successfully!', 'success');
			setTimeout(() => (isSaved = false), 3000);
		} catch (err: any) {
			errorMessage = err.message || 'Failed to update settings';
		} finally {
			isSaving = false;
		}
	}
</script>

<svelte:head>
	<title>Monitoring Settings & Channels | Pinger</title>
</svelte:head>

<div class="mx-auto max-w-4xl space-y-8">
	<div class="border-b border-slate-800 pb-6">
		<h1 class="text-2xl font-extrabold text-white">Monitoring Settings</h1>
		<p class="mt-1 text-xs text-slate-400">Configure global tenant inspection frequencies and alert notification channels</p>
	</div>

	{#if isLoading}
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-8 text-center text-xs text-slate-400">
			Loading tenant configuration...
		</div>
	{:else}
		<form onsubmit={handleSave} class="space-y-6">
			{#if isSaved}
				<div class="flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-xs text-emerald-400">
					<CheckCircle2 class="h-4 w-4 shrink-0" />
					<span>Monitoring intervals and notification settings saved successfully! Existing monitors updated.</span>
				</div>
			{/if}

			{#if errorMessage}
				<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
					<AlertCircle class="h-4 w-4 shrink-0" />
					<span>{errorMessage}</span>
				</div>
			{/if}

			<!-- Part 4: Centralized Tenant Monitoring Frequencies UI -->
			<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-6">
				<div class="flex items-center gap-3 border-b border-slate-800 pb-4">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-800/60 text-emerald-400">
						<Clock class="h-5 w-5" />
					</div>
					<div>
						<h2 class="text-base font-bold text-white">Tenant Monitoring Frequencies</h2>
						<p class="text-xs text-slate-400">All monitors in your organization inherit these execution intervals</p>
					</div>
				</div>

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<!-- HTTP Interval -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="http-int" class="text-xs font-bold text-slate-200">HTTP Uptime Check</label>
							<span class="text-[10px] font-mono text-emerald-400 font-semibold">Min: 30s</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="http-int"
								type="number"
								min="30"
								bind:value={settings.http_interval_seconds}
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">sec</span>
						</div>
					</div>

					<!-- DNS Interval -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="dns-int" class="text-xs font-bold text-slate-200">DNS Records Check</label>
							<span class="text-[10px] font-mono text-emerald-400 font-semibold">Min: 300s (5m)</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="dns-int"
								type="number"
								min="300"
								bind:value={settings.dns_interval_seconds}
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">sec</span>
						</div>
					</div>

					<!-- SSL Interval -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="ssl-int" class="text-xs font-bold text-slate-200">SSL Certificate Expiry</label>
							<span class="text-[10px] font-mono text-emerald-400 font-semibold">Min: 3600s (1h)</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="ssl-int"
								type="number"
								min="3600"
								bind:value={settings.ssl_interval_seconds}
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">sec</span>
						</div>
					</div>

					<!-- Domain Expiry Interval -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="domain-int" class="text-xs font-bold text-slate-200">Domain Expiry (WHOIS)</label>
							<span class="text-[10px] font-mono text-emerald-400 font-semibold">Min: 86400s (24h)</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="domain-int"
								type="number"
								min="86400"
								bind:value={settings.domain_interval_seconds}
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">sec</span>
						</div>
					</div>

					<!-- Email Auth Interval -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2 sm:col-span-2">
						<div class="flex items-center justify-between">
							<label for="email-int" class="text-xs font-bold text-slate-200">SPF / DMARC Email Security</label>
							<span class="text-[10px] font-mono text-emerald-400 font-semibold">Min: 300s (5m)</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="email-int"
								type="number"
								min="300"
								bind:value={settings.email_auth_interval_seconds}
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">sec</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Notification Channels Section -->
			<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-800/60 text-slate-300">
							<Mail class="h-5 w-5 text-emerald-400" />
						</div>
						<div>
							<h3 class="text-sm font-bold text-white">Email Dispatch Channel</h3>
							<p class="text-xs text-slate-400">Receive incident reports via email</p>
						</div>
					</div>
					<input type="checkbox" bind:checked={emailAlerts} class="h-5 w-5 accent-emerald-500 rounded" />
				</div>

				{#if emailAlerts}
					<div>
						<label for="email" class="block text-xs font-bold uppercase tracking-wider text-slate-400">Recipient Address</label>
						<input
							id="email"
							type="email"
							bind:value={emailAddress}
							class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 px-4 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
						/>
					</div>
				{/if}
			</div>

			<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-800/60 text-slate-300">
							<Webhook class="h-5 w-5 text-blue-400" />
						</div>
						<div>
							<h3 class="text-sm font-bold text-white">HTTP Webhook Endpoint</h3>
							<p class="text-xs text-slate-400">POST JSON alerts to custom backend endpoints</p>
						</div>
					</div>
					<input type="checkbox" bind:checked={webhookAlerts} class="h-5 w-5 accent-emerald-500 rounded" />
				</div>

				{#if webhookAlerts}
					<div>
						<label for="webhook" class="block text-xs font-bold uppercase tracking-wider text-slate-400">Webhook URL</label>
						<input
							id="webhook"
							type="url"
							bind:value={webhookUrl}
							class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 px-4 py-2 text-xs font-mono text-slate-100 focus:border-emerald-500 focus:outline-none"
						/>
					</div>
				{/if}
			</div>

			<button
				type="submit"
				disabled={isSaving || validationErrors.length > 0}
				class="flex items-center justify-center gap-2 rounded-xl bg-emerald-500 px-6 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
			>
				<Save class="h-4 w-4 stroke-[2.5]" />
				{isSaving ? 'Saving Configuration...' : 'Save Monitoring Settings'}
			</button>
		</form>
	{/if}
</div>
