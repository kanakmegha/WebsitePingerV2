import { writable } from 'svelte/store';

export interface TenantDTO {
	id: string;
	name: string;
	role: string;
}

function createTenantStore() {
	const initialTenantId = typeof window !== 'undefined' ? localStorage.getItem('pinger_selected_tenant_id') : null;

	const { subscribe, set, update } = writable<{
		selectedTenantId: string | null;
		tenants: TenantDTO[];
		ready: boolean;
	}>({
		selectedTenantId: initialTenantId,
		tenants: [],
		ready: false
	});

	return {
		subscribe,
		setTenants: (tenants: TenantDTO[]) => {
			update((state) => {
				let selected = state.selectedTenantId;
				// If no tenant selected or current selected is not in user's tenant list, pick first tenant
				if (!selected || !tenants.some((t) => t.id === selected)) {
					selected = tenants.length > 0 ? tenants[0].id : null;
				}
				if (typeof window !== 'undefined') {
					if (selected) {
						localStorage.setItem('pinger_selected_tenant_id', selected);
					} else {
						localStorage.removeItem('pinger_selected_tenant_id');
					}
				}
				return { tenants, selectedTenantId: selected, ready: tenants.length === 0 || selected !== null };
			});
		},
		selectTenant: (tenantId: string) => {
			if (typeof window !== 'undefined') {
				localStorage.setItem('pinger_selected_tenant_id', tenantId);
			}
			update((state) => ({ ...state, selectedTenantId: tenantId, ready: true }));
		},
		reset: () => {
			if (typeof window !== 'undefined') {
				localStorage.removeItem('pinger_selected_tenant_id');
			}
			set({ selectedTenantId: null, tenants: [], ready: false });
		}
	};
}

export const tenantStore = createTenantStore();
