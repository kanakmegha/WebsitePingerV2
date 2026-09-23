import { writable, derived } from 'svelte/store';
import type { AlertEvent } from '$lib/types';
import { fetchAlerts, markAlertAsRead, markAllAlertsAsRead } from '$lib/services/api';

function createAlertStore() {
	const { subscribe, set, update } = writable<AlertEvent[]>([]);

	return {
		subscribe,
		set,
		loadAlerts: async () => {
			const data = await fetchAlerts();
			set(data);
		},
		markRead: async (id: string) => {
			try {
				await markAlertAsRead(id);
				update((list) =>
					list.map((a) => (a.id === id ? { ...a, status: 'read' as const } : a))
				);
			} catch (err) {
				console.error('Failed to mark alert as read:', err);
			}
		},
		markAllRead: async () => {
			try {
				await markAllAlertsAsRead();
				update((list) => list.map((a) => ({ ...a, status: 'read' as const })));
			} catch (err) {
				console.error('Failed to mark all alerts as read:', err);
			}
		}
	};
}

export const alertStore = createAlertStore();
export const alertFilter = writable<'all' | 'unread' | 'incident' | 'recovery' | 'invite_sent' | 'invite_accepted' | 'invites'>('all');

export const unreadCount = derived(alertStore, ($alerts) => {
	return $alerts.filter((a) => a.status === 'unread').length;
});

if (typeof window !== 'undefined') {
	unreadCount.subscribe(($count) => {
		if ('setAppBadge' in navigator) {
			if ($count > 0) {
				navigator.setAppBadge($count).catch(() => {});
			} else {
				navigator.clearAppBadge().catch(() => {});
			}
		}
	});
}

export const filteredAlerts = derived(
	[alertStore, alertFilter],
	([$alerts, $filter]) => {
		if ($filter === 'all') return $alerts;
		if ($filter === 'unread') return $alerts.filter((a) => a.status === 'unread');
		if ($filter === 'invites') {
			return $alerts.filter((a) => a.type === 'invite_sent' || a.type === 'invite_accepted');
		}
		if ($filter === 'incident' || $filter === 'recovery' || $filter === 'invite_sent' || $filter === 'invite_accepted') {
			return $alerts.filter((a) => a.type === $filter);
		}
		return $alerts;
	}
);
