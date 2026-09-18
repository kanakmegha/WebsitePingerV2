<script lang="ts">
	import { tenantStore } from '$lib/stores/tenant';
	import { monitorsStore } from '$lib/stores/monitors';
	import { Building2, ChevronDown } from '@lucide/svelte';

	let isOpen = $state(false);

	let selectedTenant = $derived(
		$tenantStore.tenants.find((t) => t.id === $tenantStore.selectedTenantId) || $tenantStore.tenants[0]
	);

	function handleSelect(id: string) {
		tenantStore.selectTenant(id);
		isOpen = false;
		monitorsStore.load(); // Reload monitors for newly selected tenant
	}
</script>

{#if $tenantStore.tenants.length > 0}
	<div class="relative">
		<button
			onclick={() => (isOpen = !isOpen)}
			class="flex items-center gap-2 rounded-lg border border-slate-800 bg-slate-900/90 px-3 py-1.5 text-xs font-semibold text-slate-200 hover:border-slate-700 hover:bg-slate-800/80 transition"
		>
			<Building2 class="h-3.5 w-3.5 text-emerald-400" />
			<span class="max-w-[130px] truncate">{selectedTenant ? selectedTenant.name : 'Select Org'}</span>
			<ChevronDown class="h-3.5 w-3.5 text-slate-400" />
		</button>

		{#if isOpen}
			<div class="absolute right-0 mt-2 w-48 rounded-xl border border-slate-800 bg-slate-900 py-1 shadow-xl shadow-slate-950/50 backdrop-blur-md z-50">
				<div class="px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider text-slate-500 border-b border-slate-800/60">
					Organizations ({$tenantStore.tenants.length})
				</div>
				{#each $tenantStore.tenants as t}
					<button
						onclick={() => handleSelect(t.id)}
						class="flex w-full items-center justify-between px-3 py-2 text-left text-xs font-medium transition hover:bg-slate-800/80 {t.id === $tenantStore.selectedTenantId ? 'text-emerald-400 font-semibold bg-emerald-500/10' : 'text-slate-300'}"
					>
						<span class="truncate">{t.name}</span>
						<span class="ml-2 text-[10px] font-mono text-slate-500 uppercase px-1 rounded border border-slate-800">{t.role}</span>
					</button>
				{/each}
			</div>
		{/if}
	</div>
{/if}
