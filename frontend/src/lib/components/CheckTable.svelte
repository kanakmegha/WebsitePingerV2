<script lang="ts">
	import type { MonitorCheck } from '$lib/types';
	import StatusBadge from './StatusBadge.svelte';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';

	let { checks = [] }: { checks?: MonitorCheck[] } = $props();

	let currentPage = $state(1);
	const pageSize = 10;

	let httpChecks = $derived(checks.filter((c) => c.check_type === 'http'));
	let totalPages = $derived(Math.ceil(httpChecks.length / pageSize) || 1);
	let paginatedChecks = $derived(httpChecks.slice((currentPage - 1) * pageSize, currentPage * pageSize));
</script>

<div class="overflow-hidden rounded-xl border border-slate-800 bg-slate-900/60 backdrop-blur dark:bg-slate-900/80 shadow-xl">
	<div class="flex items-center justify-between border-b border-slate-800 px-6 py-4">
		<div>
			<h3 class="text-base font-bold text-slate-100">Recent Inspection Checks</h3>
			<p class="text-xs text-slate-400">Audit history fetched directly from PostgreSQL engine (No cache)</p>
		</div>

		<div class="flex items-center gap-2 text-xs">
			<span class="text-slate-400">Page {currentPage} of {totalPages}</span>
			<div class="flex items-center gap-1">
				<button
					onclick={() => (currentPage = Math.max(1, currentPage - 1))}
					disabled={currentPage === 1}
					class="flex h-7 w-7 items-center justify-center rounded border border-slate-800 bg-slate-800/60 text-slate-300 transition hover:bg-slate-700 disabled:opacity-40"
				>
					<ChevronLeft class="h-4 w-4" />
				</button>
				<button
					onclick={() => (currentPage = Math.min(totalPages, currentPage + 1))}
					disabled={currentPage === totalPages}
					class="flex h-7 w-7 items-center justify-center rounded border border-slate-800 bg-slate-800/60 text-slate-300 transition hover:bg-slate-700 disabled:opacity-40"
				>
					<ChevronRight class="h-4 w-4" />
				</button>
			</div>
		</div>
	</div>

	<div class="overflow-x-auto">
		<table class="w-full text-left text-sm text-slate-300">
			<thead class="border-b border-slate-800 bg-slate-950/60 text-xs uppercase text-slate-400">
				<tr>
					<th scope="col" class="px-6 py-3.5 font-semibold">Timestamp</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Check Type</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Status Code</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Response Time</th>
					<th scope="col" class="px-6 py-3.5 font-semibold">Payload Size</th>
					<th scope="col" class="px-6 py-3.5 text-right font-semibold">Result</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-slate-800/60 font-medium">
				{#each paginatedChecks as chk (chk.id)}
					<tr class="transition hover:bg-slate-800/40">
						<td class="px-6 py-3.5 font-mono text-xs text-slate-400">
							{new Date(chk.checked_at).toLocaleString()}
						</td>

						<td class="px-6 py-3.5">
							<span class="rounded border border-slate-700 bg-slate-800 px-2 py-0.5 text-xs font-mono font-semibold uppercase text-slate-300">
								{chk.check_type}
							</span>
						</td>

						<td class="px-6 py-3.5 font-mono">
							{#if chk.http_result}
								<span class={chk.http_result.status_code >= 200 && chk.http_result.status_code < 400 ? 'text-emerald-400 font-bold' : 'text-red-400 font-bold'}>
									{chk.http_result.status_code}
								</span>
							{:else}
								<span class="text-slate-500">—</span>
							{/if}
						</td>

						<td class="px-6 py-3.5 font-mono text-emerald-400">
							{chk.http_result ? `${chk.http_result.response_time_ms} ms` : '—'}
						</td>

						<td class="px-6 py-3.5 font-mono text-slate-400 text-xs">
							{chk.http_result ? `${(chk.http_result.response_size_bytes / 1024).toFixed(1)} KB` : '—'}
						</td>

						<td class="px-6 py-3.5 text-right">
							<StatusBadge status={chk.status === 'success' ? 'up' : chk.status === 'rate_limited' ? 'rate_limited' : 'down'} size="sm" />
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
