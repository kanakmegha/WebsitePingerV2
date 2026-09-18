import { writable } from 'svelte/store';

type Theme = 'dark' | 'light';

function createThemeStore() {
	const initial: Theme =
		typeof window !== 'undefined' && localStorage.getItem('theme') === 'light'
			? 'light'
			: 'dark';

	const { subscribe, set, update } = writable<Theme>(initial);

	return {
		subscribe,
		toggle: () => {
			update((current) => {
				const next = current === 'dark' ? 'light' : 'dark';
				if (typeof window !== 'undefined') {
					localStorage.setItem('theme', next);
					if (next === 'dark') {
						document.documentElement.classList.add('dark');
					} else {
						document.documentElement.classList.remove('dark');
					}
				}
				return next;
			});
		},
		init: () => {
			if (typeof window !== 'undefined') {
				const theme = localStorage.getItem('theme') || 'dark';
				set(theme as Theme);
				if (theme === 'dark') {
					document.documentElement.classList.add('dark');
				} else {
					document.documentElement.classList.remove('dark');
				}
			}
		}
	};
}

export const themeStore = createThemeStore();
