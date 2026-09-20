import { writable, derived } from 'svelte/store';
import type { Monitor, CheckType, MonitorStatus } from '$lib/types';
import { fetchMonitors, MOCK_MONITORS } from '$lib/services/api';

function createMonitorStore() {
	const { subscribe, set, update } = writable<Monitor[]>([]);

	return {
		subscribe,
		set,
		load: async () => {
			const data = await fetchMonitors();
			set(data);
		},
		add: (monitor: Monitor) => {
			update((list) => [monitor, ...list]);
		},
		remove: (monitorId: string) => {
			update((list) => list.filter((m) => m.id !== monitorId));
		},
		updateStatus: (monitorId: string, status: MonitorStatus, latencyMs?: number) => {
			update((list) =>
				list.map((m) =>
					m.id === monitorId
						? {
								...m,
								status,
								last_checked_at: new Date().toISOString(),
								avg_latency_ms: latencyMs !== undefined ? latencyMs : m.avg_latency_ms
							}
						: m
				)
			);
		}
	};
}

export const monitorsStore = createMonitorStore();

// Derived Stores for Filters & Stats
export const searchQuery = writable<string>('');
export const statusFilter = writable<string>('all'); // all, up, down, warning
export const typeFilter = writable<string>('all'); // all, http, ssl, dns, whois

export const filteredMonitors = derived(
	[monitorsStore, searchQuery, statusFilter, typeFilter],
	([$monitors, $query, $status, $type]) => {
		return $monitors.filter((m) => {
			const matchesQuery =
				m.name.toLowerCase().includes($query.toLowerCase()) ||
				m.url.toLowerCase().includes($query.toLowerCase()) ||
				m.domain.toLowerCase().includes($query.toLowerCase());

			const matchesStatus = $status === 'all' || m.status === $status;
			const matchesType = $type === 'all' || m.type === $type;

			return matchesQuery && matchesStatus && matchesType;
		});
	}
);

export const monitorStats = derived(monitorsStore, ($monitors) => {
	const total = $monitors.length;
	const up = $monitors.filter((m) => m.status === 'up').length;
	const down = $monitors.filter((m) => m.status === 'down').length;
	const warning = $monitors.filter((m) => m.status === 'warning').length;

	const latencies = $monitors.map((m) => m.avg_latency_ms || 0).filter((l) => l > 0);
	const avgLatency =
		latencies.length > 0
			? Math.round(latencies.reduce((a, b) => a + b, 0) / latencies.length)
			: 0;

	return { total, up, down, warning, avgLatency };
});
