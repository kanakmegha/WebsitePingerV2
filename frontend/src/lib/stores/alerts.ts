import { writable, derived } from 'svelte/store';
import type { AlertEvent } from '$lib/types';
import { MOCK_ALERTS } from '$lib/services/api';

function createAlertStore() {
	const { subscribe, set, update } = writable<AlertEvent[]>(MOCK_ALERTS);

	return {
		subscribe,
		set,
		add: (alert: AlertEvent) => update((list) => [alert, ...list]),
		resolve: (id: string) =>
			update((list) =>
				list.map((a) => (a.id === id ? { ...a, status: 'resolved' as const } : a))
			)
	};
}

export const alertStore = createAlertStore();
export const alertFilter = writable<'all' | 'triggered' | 'resolved'>('all');

export const filteredAlerts = derived(
	[alertStore, alertFilter],
	([$alerts, $filter]) => {
		if ($filter === 'all') return $alerts;
		return $alerts.filter((a) => a.status === $filter);
	}
);
