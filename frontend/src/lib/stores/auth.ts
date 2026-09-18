import { writable, derived } from 'svelte/store';
import type { User } from '$lib/types';

interface AuthState {
	token: string | null;
	user: User | null;
	initialized: boolean;
}

function createAuthStore() {
	const initialToken = typeof window !== 'undefined' ? localStorage.getItem('pinger_token') : null;
	const initialUserStr = typeof window !== 'undefined' ? localStorage.getItem('pinger_user') : null;
	let initialUser: User | null = null;

	if (initialUserStr) {
		try {
			initialUser = JSON.parse(initialUserStr);
		} catch (e) {
			initialUser = null;
		}
	}

	const { subscribe, set, update } = writable<AuthState>({
		token: initialToken,
		user: initialUser,
		initialized: false
	});

	return {
		subscribe,
		init: () => {
			if (typeof window !== 'undefined') {
				const token = localStorage.getItem('pinger_token');
				const userStr = localStorage.getItem('pinger_user');
				let user: User | null = null;
				if (userStr) {
					try { user = JSON.parse(userStr); } catch (e) {}
				}
				set({ token, user, initialized: true });
			}
		},
		login: (token: string, userId: string, email: string) => {
			const user: User = { id: userId, email };
			if (typeof window !== 'undefined') {
				localStorage.setItem('pinger_token', token);
				localStorage.setItem('pinger_user', JSON.stringify(user));
			}
			set({ token, user, initialized: true });
		},
		logout: () => {
			if (typeof window !== 'undefined') {
				localStorage.removeItem('pinger_token');
				localStorage.removeItem('pinger_user');
			}
			set({ token: null, user: null, initialized: true });
		}
	};
}

export const authStore = createAuthStore();

export const isAuthenticated = derived(authStore, ($auth) => !!$auth.token);
