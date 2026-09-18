<script lang="ts">
	import { goto } from '$app/navigation';
	import { createMonitor } from '$lib/services/api';
	import { monitorsStore } from '$lib/stores/monitors';
	import { toastStore } from '$lib/stores/toast';
	import { ArrowLeft, Plus, AlertCircle, CheckSquare } from '@lucide/svelte';

	let name = $state('');
	let url = $state('https://');
	let isSubmitting = $state(false);
	let errorMessage = $state('');

	// Check selections (no interval inputs - inherited from tenant settings)
	let checkHttp = $state(true);
	let checkSsl = $state(true);
	let checkDns = $state(true);
	let checkDomain = $state(true);
	let checkEmailAuth = $state(false);

	let hasSelectedCheck = $derived(checkHttp || checkSsl || checkDns || checkDomain || checkEmailAuth);

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!hasSelectedCheck) {
			errorMessage = 'At least ONE check type must be selected';
			return;
		}

		const selectedChecks: string[] = [];
		if (checkHttp) selectedChecks.push('http');
		if (checkDns) selectedChecks.push('dns');
		if (checkSsl) selectedChecks.push('ssl');
		if (checkDomain) selectedChecks.push('domain');
		if (checkEmailAuth) selectedChecks.push('email_auth');

		isSubmitting = true;
		errorMessage = '';

		try {
			const newMon = await createMonitor({
				name,
				url,
				checks: selectedChecks
			});

			monitorsStore.add(newMon);
			toastStore.show('Monitor added successfully!', 'success');
			goto('/');
		} catch (err: any) {
			errorMessage = err.message || 'Failed to create monitor';
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>Add New Monitor | Pinger</title>
</svelte:head>

<div class="mx-auto max-w-2xl space-y-8">
	<div>
		<a href="/" class="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-400 hover:text-white transition mb-4">
			<ArrowLeft class="h-4 w-4" />
			Back to Dashboard
		</a>
		<h1 class="text-2xl font-extrabold text-white">Create Domain Monitor</h1>
		<p class="mt-1 text-xs text-slate-400">Configure continuous monitoring for your target domain or URL</p>
	</div>

	<form onsubmit={handleSubmit} class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-6">
		{#if errorMessage}
			<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
				<AlertCircle class="h-4 w-4 shrink-0" />
				<span>{errorMessage}</span>
			</div>
		{/if}

		<div>
			<label for="name" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
				Monitor Name
			</label>
			<input
				id="name"
				type="text"
				required
				bind:value={name}
				placeholder="e.g. PCAPPA Main Domain"
				class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 px-4 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
			/>
		</div>

		<div>
			<label for="url" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
				Target URL / Domain
			</label>
			<input
				id="url"
				type="url"
				required
				bind:value={url}
				placeholder="https://www.pcappa.org"
				class="mt-2 w-full rounded-xl border border-slate-800 bg-slate-950 px-4 py-2.5 text-sm font-mono text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
			/>
		</div>

		<!-- Check Selection Checkboxes (No Interval Inputs) -->
		<div class="space-y-3">
			<div class="flex items-center justify-between">
				<span class="block text-xs font-bold uppercase tracking-wider text-slate-300">
					Select Inspection Checks
				</span>
				<span class="text-[11px] text-slate-400">Execution intervals are inherited from your Organization Settings</span>
			</div>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<!-- HTTP Check -->
				<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-950 p-4 cursor-pointer hover:border-slate-700 transition">
					<input type="checkbox" bind:checked={checkHttp} class="mt-0.5 h-4 w-4 accent-emerald-500 rounded" />
					<div>
						<span class="text-xs font-bold text-slate-200 block">HTTP Uptime Check</span>
						<span class="text-[11px] text-slate-400">Response code & latency metrics</span>
					</div>
				</label>

				<!-- SSL Check -->
				<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-950 p-4 cursor-pointer hover:border-slate-700 transition">
					<input type="checkbox" bind:checked={checkSsl} class="mt-0.5 h-4 w-4 accent-emerald-500 rounded" />
					<div>
						<span class="text-xs font-bold text-slate-200 block">SSL Certificate Expiry</span>
						<span class="text-[11px] text-slate-400">TLS validity & days remaining</span>
					</div>
				</label>

				<!-- DNS Check -->
				<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-950 p-4 cursor-pointer hover:border-slate-700 transition">
					<input type="checkbox" bind:checked={checkDns} class="mt-0.5 h-4 w-4 accent-emerald-500 rounded" />
					<div>
						<span class="text-xs font-bold text-slate-200 block">DNS Records (A, AAAA, MX)</span>
						<span class="text-[11px] text-slate-400">Nameserver & IP routing lookup</span>
					</div>
				</label>

				<!-- Domain Expiry Check -->
				<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-950 p-4 cursor-pointer hover:border-slate-700 transition">
					<input type="checkbox" bind:checked={checkDomain} class="mt-0.5 h-4 w-4 accent-emerald-500 rounded" />
					<div>
						<span class="text-xs font-bold text-slate-200 block">Domain Expiry (WHOIS)</span>
						<span class="text-[11px] text-slate-400">Registrar expiration tracker</span>
					</div>
				</label>

				<!-- Email Auth Check -->
				<label class="flex items-start gap-3 rounded-xl border border-slate-800 bg-slate-950 p-4 cursor-pointer hover:border-slate-700 transition sm:col-span-2">
					<input type="checkbox" bind:checked={checkEmailAuth} class="mt-0.5 h-4 w-4 accent-emerald-500 rounded" />
					<div>
						<span class="text-xs font-bold text-slate-200 block">SPF / DMARC Email Security</span>
						<span class="text-[11px] text-slate-400">Verify email authentication DNS records</span>
					</div>
				</label>
			</div>
		</div>

		{#if !hasSelectedCheck}
			<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
				<AlertCircle class="h-4 w-4 shrink-0" />
				<span>Select at least one inspection check.</span>
			</div>
		{/if}

		<button
			type="submit"
			disabled={isSubmitting || !hasSelectedCheck}
			class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-sm font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
		>
			<Plus class="h-4 w-4 stroke-[3]" />
			{isSubmitting ? 'Saving Monitor...' : 'Save Monitor'}
		</button>
	</form>
</div>
