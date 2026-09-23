<script lang="ts">
	import { fetchTenantSettings, updateTenantSettings } from '$lib/services/api';
	import { toastStore } from '$lib/stores/toast';
	import type { TenantSettings } from '$lib/types';
	import {
		isPushSupported,
		getNotificationPermission,
		getPushDiagnostics,
		subscribeUserToPush,
		unsubscribeUserFromPush,
		checkCurrentSubscription,
		type PushDiagnosticReport
	} from '$lib/services/push';
	import { Mail, Webhook, MessageSquare, Save, CheckCircle2, Clock, ShieldAlert, AlertCircle, Bell, BellOff, Info, Terminal } from '@lucide/svelte';

	let settings = $state<TenantSettings>({
		http_interval_seconds: 60,
		dns_interval_seconds: 300,
		ssl_interval_seconds: 3600,
		domain_interval_seconds: 86400,
		email_auth_interval_seconds: 300,
		ssl_min_expiry_days: 30,
		domain_min_expiry_days: 30
	});

	let isLoading = $state(true);
	let isSaving = $state(false);
	let isSaved = $state(false);
	let errorMessage = $state('');

	// Push Notifications state
	let pushSupported = $state(false);
	let pushPermission = $state<string>('default');
	let isPushSubscribed = $state(false);
	let isPushToggling = $state(false);
	let pushDiag = $state<PushDiagnosticReport>(getPushDiagnostics());

	function runDiagnostics() {
		pushDiag = getPushDiagnostics();
		toastStore.show(`Push Diagnostics: ${pushDiag.status} (${pushDiag.reason})`, 'info');
	}

	$effect(() => {
		pushDiag = getPushDiagnostics();
		pushSupported = isPushSupported();
		pushPermission = getNotificationPermission();

		if (pushSupported) {
			checkCurrentSubscription().then((sub) => {
				isPushSubscribed = !!sub;
			});
		}
	});

	async function togglePushNotifications(e: Event) {
		const target = e.target as HTMLInputElement;
		const shouldSubscribe = target.checked;
		isPushToggling = true;

		try {
			if (shouldSubscribe) {
				await subscribeUserToPush();
				isPushSubscribed = true;
				pushPermission = getNotificationPermission();
				toastStore.show('Web Push notifications enabled successfully!', 'success');
			} else {
				await unsubscribeUserFromPush();
				isPushSubscribed = false;
				toastStore.show('Web Push notifications disabled.', 'info');
			}
		} catch (err: any) {
			target.checked = !shouldSubscribe;
			toastStore.show(err.message || 'Failed to update push subscription', 'error');
		} finally {
			isPushToggling = false;
		}
	}

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
		if (settings.ssl_min_expiry_days !== undefined && settings.ssl_min_expiry_days < 1) errs.push('SSL warning days threshold must be at least 1 day');
		if (settings.domain_min_expiry_days !== undefined && settings.domain_min_expiry_days < 1) errs.push('Domain warning days threshold must be at least 1 day');
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
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
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

			<!-- Part 5: Minimum Expiry Warning Thresholds -->
			<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-6">
				<div class="flex items-center gap-3 border-b border-slate-800 pb-4">
					<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-amber-500/20 bg-amber-500/10 text-amber-400">
						<ShieldAlert class="h-5 w-5" />
					</div>
					<div>
						<h2 class="text-base font-bold text-white">Expiration Warning Thresholds</h2>
						<p class="text-xs text-slate-400">Specify the minimum days left to trigger a Warning state on your dashboard</p>
					</div>
				</div>

				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
					<!-- SSL Minimum Expiry Days -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="ssl-min" class="text-xs font-bold text-slate-200">SSL Certificate Warning</label>
							<span class="text-[10px] font-mono text-amber-400 font-semibold">Days left threshold</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="ssl-min"
								type="number"
								min="1"
								bind:value={settings.ssl_min_expiry_days}
								placeholder="30"
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-amber-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">days</span>
						</div>
					</div>

					<!-- Domain Minimum Expiry Days -->
					<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-2">
						<div class="flex items-center justify-between">
							<label for="domain-min" class="text-xs font-bold text-slate-200">Domain WHOIS Warning</label>
							<span class="text-[10px] font-mono text-amber-400 font-semibold">Days left threshold</span>
						</div>
						<div class="flex items-center gap-2">
							<input
								id="domain-min"
								type="number"
								min="1"
								bind:value={settings.domain_min_expiry_days}
								placeholder="30"
								class="w-full rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-mono text-slate-100 focus:border-amber-500 focus:outline-none"
							/>
							<span class="text-xs text-slate-400 font-mono">days</span>
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

			<!-- Web Push Notifications PWA Card -->
			<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-4">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-800/60 text-purple-400">
							{#if isPushSubscribed}
								<Bell class="h-5 w-5 text-emerald-400" />
							{:else}
								<BellOff class="h-5 w-5 text-slate-400" />
							{/if}
						</div>
						<div>
							<div class="flex flex-wrap items-center gap-2">
								<h3 class="text-sm font-bold text-white">Browser Web Push Notifications</h3>
								{#if pushDiag.status === 'insecure_context'}
									<span class="rounded-full bg-red-500/20 px-2.5 py-0.5 text-[10px] font-bold text-red-400 border border-red-500/30">
										Insecure Context (Not HTTPS)
									</span>
								{:else if pushDiag.status === 'ios_home_screen_required'}
									<span class="rounded-full bg-amber-500/20 px-2.5 py-0.5 text-[10px] font-bold text-amber-400 border border-amber-500/30">
										iOS: Home Screen PWA Required
									</span>
								{:else if pushDiag.status === 'api_unsupported'}
									<span class="rounded-full bg-amber-500/20 px-2.5 py-0.5 text-[10px] font-bold text-amber-400 border border-amber-500/30">
										Push API Unavailable
									</span>
								{:else if pushPermission === 'denied'}
									<span class="rounded-full bg-red-500/20 px-2.5 py-0.5 text-[10px] font-bold text-red-400 border border-red-500/30">
										Blocked in Browser
									</span>
								{:else if isPushSubscribed}
									<span class="rounded-full bg-emerald-500/20 px-2.5 py-0.5 text-[10px] font-bold text-emerald-400 border border-emerald-500/30">
										Enabled
									</span>
								{:else}
									<span class="rounded-full bg-slate-800 px-2.5 py-0.5 text-[10px] font-bold text-slate-400 border border-slate-700">
										Disabled
									</span>
								{/if}
							</div>
							<p class="text-xs text-slate-400 mt-0.5">Receive instant native alerts for downtime incidents & team invites on mobile and desktop</p>
						</div>
					</div>

					<input
						type="checkbox"
						checked={isPushSubscribed}
						disabled={!pushSupported || pushPermission === 'denied' || isPushToggling}
						onchange={togglePushNotifications}
						class="h-5 w-5 accent-emerald-500 rounded disabled:opacity-50 cursor-pointer shrink-0"
					/>
				</div>

				<!-- Diagnostic Feedback Boxes -->
				{#if pushDiag.status === 'insecure_context'}
					<div class="rounded-xl border border-red-500/30 bg-red-950/20 p-4 space-y-2 text-xs text-red-200">
						<div class="flex items-center gap-2 font-bold text-red-400">
							<AlertCircle class="h-4 w-4 shrink-0" />
							<span>Insecure Context Detected ({pushDiag.hostname})</span>
						</div>
						<p class="font-mono text-[11px] leading-relaxed text-red-300">
							{pushDiag.reason}
						</p>
						<div class="pt-1 text-[11px] font-semibold text-red-400">
							💡 Solution: {pushDiag.recommendation}
						</div>
					</div>
				{:else if pushDiag.status === 'ios_home_screen_required'}
					<div class="rounded-xl border border-amber-500/30 bg-amber-950/20 p-4 space-y-2 text-xs text-amber-200">
						<div class="flex items-center gap-2 font-bold text-amber-400">
							<Info class="h-4 w-4 shrink-0" />
							<span>iOS Home Screen PWA Required</span>
						</div>
						<p class="text-[11px] leading-relaxed text-amber-300">
							{pushDiag.reason}
						</p>
						<div class="pt-1 text-[11px] font-semibold text-amber-400">
							💡 Instructions: {pushDiag.recommendation}
						</div>
					</div>
				{:else if pushPermission === 'denied'}
					<div class="rounded-xl border border-red-500/30 bg-red-950/20 p-4 space-y-2 text-xs text-red-200">
						<div class="flex items-center gap-2 font-bold text-red-400">
							<AlertCircle class="h-4 w-4 shrink-0" />
							<span>Notification Permission Blocked</span>
						</div>
						<p class="text-[11px] text-red-300">
							Notification permissions are blocked in your browser settings. To allow alerts, click the lock icon in your browser address bar and enable Notifications for this site.
						</p>
					</div>
				{/if}

				<div class="flex items-center justify-between pt-2 border-t border-slate-800/80 text-[11px] text-slate-400 font-mono">
					<span>Environment: {pushDiag.hostname || 'localhost'} ({pushDiag.isSecureContext ? 'Secure Context' : 'Non-Secure Context'})</span>
					<button
						type="button"
						onclick={runDiagnostics}
						class="inline-flex items-center gap-1 text-slate-400 hover:text-emerald-400 transition"
					>
						<Terminal class="h-3.5 w-3.5" />
						Run Push Diagnostics
					</button>
				</div>
			</div>

			<button
				type="submit"
				disabled={isSaving || validationErrors.length > 0}
				class="flex w-full sm:w-auto items-center justify-center gap-2 rounded-xl bg-emerald-500 px-6 py-3.5 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50 min-h-[44px]"
			>
				<Save class="h-4 w-4 stroke-[2.5]" />
				{isSaving ? 'Saving Configuration...' : 'Save Monitoring Settings'}
			</button>
		</form>
	{/if}
</div>
