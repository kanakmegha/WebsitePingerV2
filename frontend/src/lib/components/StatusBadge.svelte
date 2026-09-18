<script lang="ts">
	import type { MonitorStatus } from '$lib/types';

	let { status = 'up', size = 'md' }: { status?: MonitorStatus | 'success' | 'failure' | 'rate_limited'; size?: 'sm' | 'md' | 'lg' } = $props();

	let isUp = $derived(status === 'up' || status === 'success');
	let isDown = $derived(status === 'down' || status === 'failure');
	let isWarning = $derived(status === 'warning');
	let isRateLimited = $derived(status === 'rate_limited');

	let sizeClasses = $derived(
		size === 'sm'
			? 'px-2 py-0.5 text-xs'
			: size === 'lg'
				? 'px-3.5 py-1.5 text-sm font-semibold'
				: 'px-2.5 py-1 text-xs font-medium'
	);
</script>

<span
	class="inline-flex items-center gap-1.5 rounded-full border transition-colors {sizeClasses} {isUp
		? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400 dark:text-emerald-400'
		: isDown
			? 'border-red-500/30 bg-red-500/10 text-red-400 dark:text-red-400'
			: isWarning
				? 'border-amber-500/30 bg-amber-500/10 text-amber-400 dark:text-amber-400'
				: isRateLimited
					? 'border-blue-500/30 bg-blue-500/10 text-blue-400'
					: 'border-slate-500/30 bg-slate-500/10 text-slate-400'}"
>
	<span
		class="h-2 w-2 rounded-full {isUp
			? 'animate-pulse bg-emerald-500'
			: isDown
				? 'bg-red-500'
				: isWarning
					? 'bg-amber-400'
					: 'bg-blue-400'}"
	></span>
	<span class="capitalize">
		{isUp ? 'Healthy' : isDown ? 'Down' : isWarning ? 'Warning' : isRateLimited ? 'Rate Limited' : status}
	</span>
</span>
