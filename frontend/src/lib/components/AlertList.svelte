<script lang="ts">
	import type { AlertEvent } from '$lib/types';
	import { ShieldAlert, CheckCircle2 } from '@lucide/svelte';

	let { alerts = [] }: { alerts?: AlertEvent[] } = $props();
</script>

<div class="space-y-3">
	{#if alerts.length === 0}
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-12 text-center backdrop-blur">
			<CheckCircle2 class="mx-auto h-10 w-10 text-emerald-400 opacity-60" />
			<h4 class="mt-3 text-base font-bold text-slate-200">No Alerts Found</h4>
			<p class="mt-1 text-xs text-slate-400">All monitored targets are operational within defined thresholds.</p>
		</div>
	{:else}
		{#each alerts as alt (alt.id)}
			<div
				class="flex items-start justify-between gap-4 rounded-xl border p-4 backdrop-blur transition hover:border-slate-700 {alt.status === 'triggered'
					? 'border-red-500/30 bg-red-950/20 text-red-200'
					: 'border-emerald-500/30 bg-emerald-950/20 text-emerald-200'}"
			>
				<div class="flex items-start gap-3.5">
					<div
						class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border {alt.status === 'triggered'
							? 'border-red-500/30 bg-red-500/10 text-red-400'
							: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'}"
					>
						{#if alt.status === 'triggered'}
							<ShieldAlert class="h-5 w-5" />
						{:else}
							<CheckCircle2 class="h-5 w-5" />
						{/if}
					</div>

					<div>
						<div class="flex items-center gap-2">
							<span class="font-bold text-slate-100">{alt.monitor_name || 'Monitor'}</span>
							<span
								class="rounded-full px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider {alt.status === 'triggered'
									? 'bg-red-500/20 text-red-400 border border-red-500/30'
									: 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'}"
							>
								{alt.status}
							</span>
						</div>
						<p class="mt-1 font-mono text-xs text-slate-300">{alt.message}</p>
					</div>
				</div>

				<div class="shrink-0 text-right font-mono text-xs text-slate-400">
					{new Date(alt.created_at).toLocaleString()}
				</div>
			</div>
		{/each}
	{/if}
</div>
