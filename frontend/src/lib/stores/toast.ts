import { writable } from 'svelte/store';

export interface Toast {
	id: string;
	message: string;
	type: 'success' | 'error' | 'info' | 'warning';
}

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	return {
		subscribe,
		show: (message: string, type: 'success' | 'error' | 'info' | 'warning' = 'success') => {
			const id = Math.random().toString(36).substring(2, 9);
			const newToast: Toast = { id, message, type };
			update((toasts) => [...toasts, newToast]);

			setTimeout(() => {
				update((toasts) => toasts.filter((t) => t.id !== id));
			}, 4000);
		},
		dismiss: (id: string) => {
			update((toasts) => toasts.filter((t) => t.id !== id));
		}
	};
}

export const toastStore = createToastStore();
