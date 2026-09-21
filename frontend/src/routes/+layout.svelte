<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import Navbar from '$lib/components/Navbar.svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import { themeStore } from '$lib/stores/theme';
	import { authStore, isAuthenticated } from '$lib/stores/auth';
	import { tenantStore } from '$lib/stores/tenant';
	import { fetchUserTenants } from '$lib/services/api';
	import '../app.css';

	let { children } = $props();
	let isBootstrapping = $state(true);

	onMount(async () => {
		themeStore.init();
		authStore.init();

		const token = localStorage.getItem('pinger_token');
		if (token) {
			try {
				const tenants = await fetchUserTenants();
				if (!tenants || tenants.length === 0) {
					console.warn('[BOOTSTRAP] No tenants associated with user account.');
				}
			} catch (e) {
				console.error('[BOOTSTRAP ERROR] Tenant initialization failed:', e);
			}
		}
		isBootstrapping = false;
	});

	$effect(() => {
		const currentPath = $page.url.pathname;
		const isAuthPage = currentPath === '/login' || currentPath === '/register';
		const isInvitePage = currentPath.startsWith('/invite');
		const isOnboardingPage = currentPath === '/onboarding';

		if ($authStore.initialized && !$isAuthenticated && !isAuthPage && !isInvitePage) {
			const redirectTo = encodeURIComponent($page.url.pathname + $page.url.search);
			goto(`/login?redirectTo=${redirectTo}`);
		} else if ($authStore.initialized && $isAuthenticated && !isBootstrapping && $tenantStore.tenants.length === 0 && !isOnboardingPage && !isInvitePage) {
			goto('/onboarding');
		}
	});
</script>

<div class="min-h-screen bg-slate-950 font-sans text-slate-100 antialiased selection:bg-emerald-500 selection:text-slate-950">
	<div class="pointer-events-none fixed inset-0 z-0 bg-[radial-gradient(ellipse_80%_80%_at_50%_-20%,rgba(16,185,129,0.12),rgba(255,255,255,0))]"></div>
	
	<Navbar />

	<main class="relative z-10 mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
		{#if isBootstrapping && $isAuthenticated}
			<div class="flex flex-col items-center justify-center py-24 space-y-4">
				<div class="h-10 w-10 animate-spin rounded-full border-4 border-emerald-500/20 border-t-emerald-500"></div>
				<p class="text-sm font-medium text-slate-400">Initializing organization workspace...</p>
			</div>
		{:else}
			{@render children()}
		{/if}
	</main>

	<ToastContainer />
</div>
