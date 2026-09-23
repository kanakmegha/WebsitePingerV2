<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { themeStore } from '$lib/stores/theme';
	import { alertStore, unreadCount } from '$lib/stores/alerts';
	import { authStore, isAuthenticated } from '$lib/stores/auth';
	import { tenantStore } from '$lib/stores/tenant';
	import { toastStore } from '$lib/stores/toast';
	import TenantSwitcher from '$lib/components/TenantSwitcher.svelte';
	import {
		Activity,
		Settings,
		ShieldAlert,
		LayoutDashboard,
		Plus,
		Sun,
		Moon,
		LogOut,
		LogIn,
		User,
		Users,
		Menu,
		X
	} from '@lucide/svelte';

	let isMobileMenuOpen = $state(false);
	let activeAlertCount = $derived($unreadCount);

	// Close drawer on route navigation
	$effect(() => {
		const currentPath = $page.url.pathname;
		if (currentPath) {
			isMobileMenuOpen = false;
		}
	});

	function handleLogout() {
		isMobileMenuOpen = false;
		authStore.logout();
		tenantStore.reset();
		toastStore.show('Logged out successfully', 'info');
		goto('/login');
	}
</script>

<header class="sticky top-0 z-50 border-b border-slate-800 bg-slate-950/90 backdrop-blur-md pt-safe pl-safe pr-safe">
	<div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
		<div class="flex items-center gap-4 lg:gap-8">
			<a href="/" class="flex items-center gap-2.5 font-bold tracking-tight text-white transition hover:opacity-90">
				<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-emerald-500 to-cyan-500 shadow-lg shadow-emerald-500/20">
					<Activity class="h-5 w-5 text-slate-950" />
				</div>
				<span class="text-xl font-extrabold tracking-wider bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent flex items-center">
					PINGER<span class="text-xl text-emerald-400 font-mono ml-1.5 px-2 py-0.5 rounded border border-emerald-500/30 bg-emerald-500/10 font-extrabold inline-block">LIVE</span>
				</span>
			</a>

			{#if $isAuthenticated}
				<nav class="hidden items-center gap-1 md:flex">
					<a
						href="/"
						class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition {$page.url.pathname === '/' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'}"
					>
						<LayoutDashboard class="h-4 w-4 text-emerald-400" />
						Dashboard
					</a>
					<a
						href="/alerts"
						class="relative flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition {$page.url.pathname === '/alerts' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'}"
					>
						<ShieldAlert class="h-4 w-4 text-amber-400" />
						Alerts
						{#if activeAlertCount > 0}
							<span class="flex h-5 w-5 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white shadow-md shadow-red-500/30">
								{activeAlertCount}
							</span>
						{/if}
					</a>
					<a
						href="/settings"
						class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition {$page.url.pathname === '/settings' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'}"
					>
						<Settings class="h-4 w-4 text-slate-400" />
						Settings
					</a>
					<a
						href="/team"
						class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition {$page.url.pathname === '/team' ? 'bg-slate-800 text-white font-semibold' : 'text-slate-300 hover:bg-slate-800/60 hover:text-white'}"
					>
						<Users class="h-4 w-4 text-cyan-400" />
						Team
					</a>
				</nav>
			{/if}
		</div>

		<div class="flex items-center gap-2 sm:gap-3">
			{#if $isAuthenticated}
				<div class="hidden sm:block">
					<TenantSwitcher />
				</div>

				<div class="hidden lg:flex items-center gap-2 px-3 py-1.5 rounded-lg border border-slate-800 bg-slate-900 text-xs text-slate-300 font-mono">
					<User class="h-3.5 w-3.5 text-emerald-400" />
					<span class="truncate max-w-[130px]">{$authStore.user?.email}</span>
				</div>

				<button
					onclick={handleLogout}
					title="Sign Out"
					aria-label="Sign Out"
					class="hidden sm:flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-400 transition hover:border-slate-700 hover:text-red-400 min-h-[36px] min-w-[36px]"
				>
					<LogOut class="h-4 w-4" />
				</button>
			{:else}
				<a
					href="/login"
					class="inline-flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3.5 py-2 text-xs font-semibold text-slate-200 hover:border-slate-700 hover:text-white transition"
				>
					<LogIn class="h-4 w-4" />
					Sign In
				</a>
			{/if}

			<button
				onclick={() => themeStore.toggle()}
				aria-label="Toggle Dark Mode"
				class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-400 transition hover:border-slate-700 hover:text-white min-h-[36px] min-w-[36px]"
			>
				<Sun class="hidden h-4 w-4 dark:block" />
				<Moon class="h-4 w-4 dark:hidden" />
			</button>

			{#if $isAuthenticated}
				<button
					onclick={() => (isMobileMenuOpen = !isMobileMenuOpen)}
					aria-label="Toggle Navigation Menu"
					class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-300 md:hidden hover:border-slate-700 min-h-[36px] min-w-[36px]"
				>
					{#if isMobileMenuOpen}
						<X class="h-5 w-5 text-emerald-400" />
					{:else}
						<Menu class="h-5 w-5" />
					{/if}
				</button>
			{/if}
		</div>
	</div>

	<!-- Mobile Drawer Overlay & Navigation -->
	{#if $isAuthenticated && isMobileMenuOpen}
		<div
			onclick={() => (isMobileMenuOpen = false)}
			class="fixed inset-0 z-40 bg-slate-950/80 backdrop-blur-sm md:hidden"
			role="presentation"
		></div>

		<nav class="absolute left-0 right-0 top-full z-50 border-b border-slate-800 bg-slate-950 p-4 shadow-2xl md:hidden space-y-4">
			<div class="space-y-1 border-b border-slate-800 pb-3">
				<div class="px-3 py-1 text-[10px] font-bold uppercase tracking-wider text-slate-500">Organization</div>
				<TenantSwitcher />
			</div>

			<div class="space-y-1">
				<a
					href="/"
					onclick={() => (isMobileMenuOpen = false)}
					class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition {$page.url.pathname === '/' ? 'bg-emerald-500/10 text-emerald-400 font-semibold' : 'text-slate-200 hover:bg-slate-900'}"
				>
					<LayoutDashboard class="h-5 w-5 text-emerald-400" />
					Dashboard
				</a>
				<a
					href="/alerts"
					onclick={() => (isMobileMenuOpen = false)}
					class="flex items-center justify-between rounded-lg px-3 py-3 text-sm font-medium transition {$page.url.pathname === '/alerts' ? 'bg-emerald-500/10 text-emerald-400 font-semibold' : 'text-slate-200 hover:bg-slate-900'}"
				>
					<div class="flex items-center gap-3">
						<ShieldAlert class="h-5 w-5 text-amber-400" />
						Alerts
					</div>
					{#if activeAlertCount > 0}
						<span class="rounded-full bg-red-500 px-2 py-0.5 text-xs font-bold text-white">
							{activeAlertCount}
						</span>
					{/if}
				</a>
				<a
					href="/settings"
					onclick={() => (isMobileMenuOpen = false)}
					class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition {$page.url.pathname === '/settings' ? 'bg-emerald-500/10 text-emerald-400 font-semibold' : 'text-slate-200 hover:bg-slate-900'}"
				>
					<Settings class="h-5 w-5 text-slate-400" />
					Settings
				</a>
				<a
					href="/team"
					onclick={() => (isMobileMenuOpen = false)}
					class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition {$page.url.pathname === '/team' ? 'bg-emerald-500/10 text-emerald-400 font-semibold' : 'text-slate-200 hover:bg-slate-900'}"
				>
					<Users class="h-5 w-5 text-cyan-400" />
					Team
				</a>
			</div>

			<div class="pt-2 border-t border-slate-800">
				<div class="flex items-center justify-between px-1 py-1">
					<div class="flex items-center gap-2 text-xs font-mono text-slate-400 truncate">
						<User class="h-4 w-4 text-emerald-400 shrink-0" />
						<span class="truncate">{$authStore.user?.email}</span>
					</div>
					<button
						onclick={handleLogout}
						class="flex items-center gap-1.5 rounded-lg border border-slate-800 bg-slate-900 px-3 py-2 text-xs font-semibold text-red-400 hover:bg-red-500/10 transition"
					>
						<LogOut class="h-4 w-4" />
						Sign Out
					</button>
				</div>
			</div>
		</nav>
	{/if}
</header>
