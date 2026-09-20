<script lang="ts">
	import type { Monitor } from '$lib/types';
	import StatusBadge from './StatusBadge.svelte';
	import { Globe, ArrowUpRight, Trash2 } from '@lucide/svelte';
	import { deleteMonitor } from '$lib/services/api';
	import { monitorsStore } from '$lib/stores/monitors';

	let { monitor }: { monitor: Monitor } = $props();
	let isDeleting = $state(false);

	async function handleDelete(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();

		if (!confirm(`Are you sure you want to delete monitor "${monitor.name}"?`)) {
			return;
		}

		try {
			isDeleting = true;
			await deleteMonitor(monitor.id);
			monitorsStore.remove(monitor.id);
		} catch (err: any) {
			alert(err.message || 'Failed to delete monitor');
		} finally {
			isDeleting = false;
		}
	}
</script>

<a
	href="/monitor/{monitor.id}"
	class="group relative flex flex-col justify-between overflow-hidden rounded-xl border border-slate-800 bg-slate-900/60 p-5 backdrop-blur transition-all duration-200 hover:-translate-y-1 hover:border-slate-700 hover:shadow-xl hover:shadow-emerald-500/5 dark:bg-slate-900/80 {isDeleting ? 'opacity-50 pointer-events-none' : ''}"
>
	<div>
		<div class="flex items-start justify-between gap-3">
			<div class="min-w-0 flex-1">
				<div class="flex items-center gap-2">
					<StatusBadge status={monitor.status} size="sm" />
				</div>
				<h3 class="mt-2.5 truncate text-base font-bold text-slate-100 group-hover:text-emerald-400">
					{monitor.name}
				</h3>
				<p class="mt-1 flex items-center gap-1.5 truncate text-xs text-slate-400">
					<Globe class="h-3.5 w-3.5 text-slate-500" />
					<span class="truncate">{monitor.url}</span>
				</p>
			</div>

			<div class="flex items-center gap-1">
				<button
					type="button"
					onclick={handleDelete}
					title="Delete monitor"
					aria-label="Delete monitor"
					class="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-800/50 text-slate-400 hover:bg-red-500/20 hover:text-red-400 transition"
				>
					<Trash2 class="h-4 w-4" />
				</button>
				<div class="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-800/50 text-slate-400 group-hover:bg-emerald-500/10 group-hover:text-emerald-400">
					<ArrowUpRight class="h-4 w-4 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
				</div>
			</div>
		</div>
	</div>

	<div class="mt-6 grid grid-cols-3 gap-2 border-t border-slate-800/80 pt-4 text-xs">
		<div>
			<span class="text-[11px] text-slate-500">Latency</span>
			<p class="mt-0.5 font-mono font-semibold text-slate-200">
				{monitor.status === 'down' ? '—' : (monitor.avg_latency_ms !== undefined && monitor.avg_latency_ms !== null ? `${monitor.avg_latency_ms}ms` : 'N/A')}
			</p>
		</div>

		<div>
			<span class="text-[11px] text-slate-500">SSL Expiry</span>
			<p class="mt-0.5 font-mono font-semibold {monitor.ssl_days_remaining && monitor.ssl_days_remaining < 14 ? 'text-amber-400' : 'text-slate-200'}">
				{monitor.ssl_days_remaining ? `${monitor.ssl_days_remaining} days` : 'N/A'}
			</p>
		</div>

		<div>
			<span class="text-[11px] text-slate-500">24h Uptime</span>
			<p class="mt-0.5 font-mono font-semibold text-emerald-400">
				{monitor.uptime_pct_24h || 100}%
			</p>
		</div>
	</div>
</a>
