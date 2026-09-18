<script lang="ts">
	import { goto } from '$app/navigation';
	import { themeStore } from '$lib/stores/theme';
	import { alertStore } from '$lib/stores/alerts';
	import { authStore, isAuthenticated } from '$lib/stores/auth';
	import { tenantStore } from '$lib/stores/tenant';
	import { toastStore } from '$lib/stores/toast';
	import TenantSwitcher from '$lib/components/TenantSwitcher.svelte';
	import { Activity, Settings, ShieldAlert, LayoutDashboard, Plus, Sun, Moon, LogOut, LogIn, User } from '@lucide/svelte';

	let activeAlertCount = $derived($alertStore.filter((a) => a.status === 'triggered').length);

	function handleLogout() {
		authStore.logout();
		tenantStore.reset();
		toastStore.show('Logged out successfully', 'info');
		goto('/login');
	}
</script>

<header class="sticky top-0 z-50 border-b border-slate-800 bg-slate-950/80 backdrop-blur-md dark:bg-slate-950/90">
	<div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6">
		<div class="flex items-center gap-8">
			<a href="/" class="flex items-center gap-2.5 font-bold tracking-tight text-white transition hover:opacity-90">
				<div class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-emerald-500 to-cyan-500 shadow-lg shadow-emerald-500/20">
					<Activity class="h-5 w-5 text-slate-950" />
				</div>
				<span class="text-xl font-extrabold tracking-wider bg-gradient-to-r from-white via-slate-200 to-slate-400 bg-clip-text text-transparent">
					PINGER<span class="text-xs text-emerald-400 font-mono ml-1 px-1.5 py-0.5 rounded border border-emerald-500/30 bg-emerald-500/10">GO</span>
				</span>
			</a>

			{#if $isAuthenticated}
				<nav class="hidden items-center gap-1 md:flex">
					<a
						href="/"
						class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800/60 hover:text-white"
					>
						<LayoutDashboard class="h-4 w-4 text-emerald-400" />
						Dashboard
					</a>
					<a
						href="/alerts"
						class="relative flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800/60 hover:text-white"
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
						class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800/60 hover:text-white"
					>
						<Settings class="h-4 w-4 text-slate-400" />
						Settings
					</a>
				</nav>
			{/if}
		</div>

		<div class="flex items-center gap-3">
			{#if $isAuthenticated}
				<TenantSwitcher />

				<a
					href="/add-monitor"
					class="inline-flex items-center gap-2 rounded-lg bg-emerald-500 px-3.5 py-2 text-xs font-semibold text-slate-950 shadow-md shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95"
				>
					<Plus class="h-4 w-4 stroke-[3]" />
					New Monitor
				</a>

				<div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-lg border border-slate-800 bg-slate-900 text-xs text-slate-300 font-mono">
					<User class="h-3.5 w-3.5 text-emerald-400" />
					<span class="truncate max-w-[140px]">{$authStore.user?.email}</span>
				</div>

				<button
					onclick={handleLogout}
					title="Sign Out"
					class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-400 transition hover:border-slate-700 hover:text-red-400"
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
				class="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-400 transition hover:border-slate-700 hover:text-white"
			>
				<Sun class="hidden h-4 w-4 dark:block" />
				<Moon class="h-4 w-4 dark:hidden" />
			</button>
		</div>
	</div>
</header>
