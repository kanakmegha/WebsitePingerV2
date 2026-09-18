<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	import { CheckCircle2, AlertCircle, Info, AlertTriangle, X } from '@lucide/svelte';
</script>

<div class="fixed bottom-5 right-5 z-50 flex flex-col gap-2.5 max-w-sm w-full pointer-events-none">
	{#each $toastStore as toast (toast.id)}
		<div
			class="pointer-events-auto flex items-center justify-between gap-3 rounded-xl border p-4 shadow-2xl backdrop-blur transition-all duration-300 animate-in fade-in slide-in-from-bottom-5 {toast.type === 'success'
				? 'border-emerald-500/30 bg-slate-900/95 text-emerald-300 shadow-emerald-500/10'
				: toast.type === 'error'
					? 'border-red-500/30 bg-slate-900/95 text-red-300 shadow-red-500/10'
					: toast.type === 'warning'
						? 'border-amber-500/30 bg-slate-900/95 text-amber-300 shadow-amber-500/10'
						: 'border-blue-500/30 bg-slate-900/95 text-blue-300'}"
		>
			<div class="flex items-center gap-3">
				{#if toast.type === 'success'}
					<CheckCircle2 class="h-5 w-5 text-emerald-400 shrink-0" />
				{:else if toast.type === 'error'}
					<AlertCircle class="h-5 w-5 text-red-400 shrink-0" />
				{:else if toast.type === 'warning'}
					<AlertTriangle class="h-5 w-5 text-amber-400 shrink-0" />
				{:else}
					<Info class="h-5 w-5 text-blue-400 shrink-0" />
				{/if}
				<p class="text-xs font-semibold text-slate-100">{toast.message}</p>
			</div>

			<button
				onclick={() => toastStore.dismiss(toast.id)}
				class="text-slate-400 hover:text-white transition"
			>
				<X class="h-4 w-4" />
			</button>
		</div>
	{/each}
</div>
