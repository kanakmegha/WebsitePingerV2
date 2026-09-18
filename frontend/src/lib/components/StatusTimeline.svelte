<script lang="ts">
	let { checksCount = 40, failIndexes = [12, 13] }: { checksCount?: number; failIndexes?: number[] } = $props();

	let blocks = $derived(
		Array.from({ length: checksCount }, (_, i) => {
			const isFail = failIndexes.includes(i);
			const minutesAgo = (checksCount - i) * 5;
			return {
				id: i,
				status: isFail ? 'down' : 'up',
				label: `${minutesAgo} mins ago - ${isFail ? 'Downtime (503 Error)' : '100% Operational (42ms)'}`
			};
		})
	);
</script>

<div class="space-y-2">
	<div class="flex items-center justify-between text-xs text-slate-400">
		<span>History (Last 40 checks)</span>
		<span class="font-medium text-emerald-400">99.98% Uptime</span>
	</div>

	<div class="flex h-7 w-full items-center gap-1 rounded-lg border border-slate-800 bg-slate-900/60 p-1.5 backdrop-blur">
		{#each blocks as block}
			<div
				title={block.label}
				class="h-full flex-1 rounded-xs transition-all hover:scale-y-125 hover:brightness-125 {block.status === 'up'
					? 'bg-emerald-500/80 shadow-xs shadow-emerald-500/20'
					: 'bg-red-500 shadow-xs shadow-red-500/20'}"
			></div>
		{/each}
	</div>
</div>
