<script lang="ts">
	import type { Monitor } from '$lib/types';
	import StatusBadge from './StatusBadge.svelte';
	import { ArrowUpDown, ExternalLink, Globe } from '@lucide/svelte';

	let { monitors = [] }: { monitors?: Monitor[] } = $props();

	let sortKey = $state<'name' | 'avg_latency_ms' | 'ssl_days_remaining' | 'status'>('name');
	let sortAsc = $state(true);

	function toggleSort(key: typeof sortKey) {
		if (sortKey === key) {
			sortAsc = !sortAsc;
		} else {
			sortKey = key;
			sortAsc = true;
		}
	}

	function getActiveChecksList(m: Monitor): string[] {
		const list: string[] = [];
		if (m.check_configs && m.check_configs.length > 0) {
			m.check_configs.forEach((c) => {
				if (c.is_enabled) {
					list.push(c.check_type.toUpperCase());
				}
			});
		}
		if (list.length === 0 && m.type) {
			list.push(m.type.toUpperCase());
		}
		return list.length > 0 ? list : ['HTTP'];
	}

	let sortedMonitors = $derived(
		[...monitors].sort((a, b) => {
			let valA = a[sortKey] ?? 0;
			let valB = b[sortKey] ?? 0;

			if (typeof valA === 'string') {
				return sortAsc
					? (valA as string).localeCompare(valB as string)
					: (valB as string).localeCompare(valA as string);
			}

			return sortAsc ? (valA as number) - (valB as number) : (valB as number) - (valA as number);
		})
	);
</script>

<div class="overflow-hidden rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur dark:bg-slate-900/80 shadow-xl">
	<div class="overflow-x-auto">
		<table class="w-full text-left text-sm text-slate-300">
			<thead class="border-b border-slate-800 bg-slate-950/60 text-xs uppercase text-slate-400">
				<tr>
					<th scope="col" class="px-6 py-3.5">
						<button onclick={() => toggleSort('name')} class="flex items-center gap-1.5 font-semibold transition hover:text-white">
							Monitor / Domain
							<ArrowUpDown class="h-3.5 w-3.5 text-slate-500" />
						</button>
					</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Active Checks</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Status</th>
					<th scope="col" class="px-6 py-3.5">
						<button onclick={() => toggleSort('avg_latency_ms')} class="flex items-center gap-1.5 font-semibold transition hover:text-white">
							Response Time
							<ArrowUpDown class="h-3.5 w-3.5 text-slate-500" />
						</button>
					</th>
					<th scope="col" class="px-6 py-3.5">
						<button onclick={() => toggleSort('ssl_days_remaining')} class="flex items-center gap-1.5 font-semibold transition hover:text-white">
							SSL Expiry
							<ArrowUpDown class="h-3.5 w-3.5 text-slate-500" />
						</button>
					</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Last Checked</th>
					<th scope="col" class="px-6 py-3.5 text-right font-semibold">Action</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-800/60 font-medium">
				{#if sortedMonitors.length === 0}
					<tr>
						<td colspan="7" class="px-6 py-12 text-center text-slate-500">
							No monitors match the selected filter criteria.
						</td>
					</tr>
				{:else}
					{#each sortedMonitors as m (m.id)}
						<tr class="transition hover:bg-slate-800/40">
							<td class="px-6 py-4">
								<a href="/monitor/{m.id}" class="group flex items-center gap-2.5">
									<div class="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-800 bg-slate-800/60 text-slate-400 group-hover:border-slate-700 group-hover:text-emerald-400">
										<Globe class="h-4 w-4" />
									</div>
									<div>
										<p class="font-semibold text-slate-100 group-hover:text-emerald-400 transition">{m.name}</p>
										<p class="text-xs font-normal text-slate-400">{m.url}</p>
									</div>
								</a>
							</td>

							<td class="px-6 py-4">
								<div class="flex flex-wrap items-center gap-1 font-mono text-[11px]">
									{#each getActiveChecksList(m) as checkTag}
										<span class="rounded border border-slate-700 bg-slate-800/80 px-2 py-0.5 font-bold uppercase text-slate-300">
											{checkTag}
										</span>
									{/each}
								</div>
							</td>

							<td class="px-6 py-4">
								<StatusBadge status={m.status} size="sm" />
							</td>

							<td class="px-6 py-4 font-mono">
								{#if m.status === 'down'}
									<span class="text-red-400 font-bold">Failed</span>
								{:else}
									<span class="text-emerald-400">{m.avg_latency_ms || 42} ms</span>
								{/if}
							</td>

							<td class="px-6 py-4 font-mono">
								{#if m.ssl_days_remaining}
									<span class={m.ssl_days_remaining < 14 ? 'text-amber-400 font-bold' : 'text-slate-300'}>
										{m.ssl_days_remaining} days
									</span>
								{:else}
									<span class="text-slate-500">N/A</span>
								{/if}
							</td>

							<td class="px-6 py-4 text-xs text-slate-400 font-mono">
								{m.last_checked_at ? new Date(m.last_checked_at).toLocaleTimeString() : 'Just now'}
							</td>

							<td class="px-6 py-4 text-right">
								<a
									href="/monitor/{m.id}"
									class="inline-flex items-center gap-1 text-xs font-semibold text-emerald-400 hover:text-emerald-300 hover:underline"
								>
									Details
									<ExternalLink class="h-3.5 w-3.5" />
								</a>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>
