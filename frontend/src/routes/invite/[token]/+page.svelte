<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { fetchInviteDetails, acceptInvite } from '$lib/services/api';
	import { authStore } from '$lib/stores/auth';
	import { toastStore } from '$lib/stores/toast';
	import type { InviteDetails } from '$lib/types';
	import { Building2, UserCheck, AlertCircle, CheckCircle2, Loader2, KeyRound, Mail, LogOut } from '@lucide/svelte';

	let token = $derived(page.params.token);

	let invite = $state<InviteDetails | null>(null);
	let loading = $state(true);
	let accepting = $state(false);
	let error = $state('');
	let successMessage = $state('');

	let password = $state('');
	let confirmPassword = $state('');

	let isLoggedIn = $derived(!!$authStore.token);
	let currentUserEmail = $derived($authStore.user?.email || '');
	let isMatchingEmail = $derived(
		isLoggedIn && invite && currentUserEmail.toLowerCase() === invite.email.toLowerCase()
	);

	onMount(async () => {
		if (!token) {
			error = 'No invitation token provided in URL.';
			loading = false;
			return;
		}

		try {
			invite = await fetchInviteDetails(token);
		} catch (err: any) {
			error = err.message || 'Invitation is invalid or has expired.';
		} finally {
			loading = false;
		}
	});

	function handleLogoutAndSwitch() {
		authStore.logout();
	}

	async function handleAccept() {
		if (!invite || !token) return;

		error = '';

		// Validation for unauthenticated new users
		if (!isLoggedIn) {
			if (!password) {
				error = 'Please enter a password to set up your account.';
				return;
			}
			if (password.length < 6) {
				error = 'Password must be at least 6 characters long.';
				return;
			}
			if (password !== confirmPassword) {
				error = 'Passwords do not match. Please check and try again.';
				return;
			}
		}

		accepting = true;

		try {
			const res = await acceptInvite(token, !isLoggedIn ? password : undefined);
			successMessage = `Successfully joined ${invite.tenant_name}! Redirecting to dashboard...`;
			toastStore.show(successMessage, 'success');
			setTimeout(() => {
				goto('/');
			}, 1200);
		} catch (err: any) {
			error = err.message || 'Failed to accept invitation. Please try again.';
		} finally {
			accepting = false;
		}
	}
</script>

<svelte:head>
	<title>{invite ? `Join ${invite.tenant_name}` : 'Organization Invitation'} | Pinger</title>
</svelte:head>

<div class="mx-auto max-w-lg space-y-8 py-12 px-4">
	<div class="text-center space-y-2">
		<div class="inline-flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-tr from-emerald-500 to-cyan-500 shadow-xl shadow-emerald-500/20">
			<Building2 class="h-7 w-7 text-slate-950" />
		</div>
		<h1 class="text-2xl font-extrabold text-white">Team Invitation</h1>
		<p class="text-xs text-slate-400">Join organization workspace on Pinger</p>
	</div>

	<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-6">
		{#if loading}
			<div class="flex flex-col items-center justify-center py-8 space-y-3">
				<Loader2 class="h-8 w-8 animate-spin text-emerald-400" />
				<p class="text-xs font-medium text-slate-400">Validating invitation token...</p>
			</div>
		{:else if error && !invite}
			<div class="rounded-xl border border-red-500/30 bg-red-500/10 p-4 space-y-2 text-center">
				<AlertCircle class="h-6 w-6 text-red-400 mx-auto" />
				<h3 class="text-sm font-bold text-red-300">Invitation Invalid or Expired</h3>
				<p class="text-xs text-red-400">{error}</p>
			</div>
			<div class="pt-2 text-center">
				<a href="/" class="inline-flex items-center gap-1 text-xs font-bold text-slate-400 hover:text-white transition">
					Return to Home
				</a>
			</div>
		{:else if successMessage}
			<div class="rounded-xl border border-emerald-500/30 bg-emerald-500/10 p-5 space-y-3 text-center">
				<CheckCircle2 class="h-8 w-8 text-emerald-400 mx-auto" />
				<h3 class="text-base font-bold text-emerald-300">Welcome to the Team!</h3>
				<p class="text-xs text-emerald-400">{successMessage}</p>
			</div>
		{:else if invite}
			<div class="space-y-6">
				<!-- Organization Summary -->
				<div class="rounded-xl border border-slate-800 bg-slate-950 p-4 space-y-3">
					<div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
						<span class="text-xs font-semibold text-slate-400">Organization</span>
						<span class="text-sm font-bold text-white flex items-center gap-1.5">
							<Building2 class="h-4 w-4 text-emerald-400" />
							{invite.tenant_name}
						</span>
					</div>
					<div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
						<span class="text-xs font-semibold text-slate-400">Invited Email</span>
						<span class="text-xs font-medium text-slate-200">{invite.email}</span>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-xs font-semibold text-slate-400">Role</span>
						<span class="inline-flex items-center rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-bold text-emerald-400 border border-emerald-500/20 uppercase tracking-wide">
							{invite.role}
						</span>
					</div>
				</div>

				{#if error}
					<div class="rounded-lg border border-red-500/30 bg-red-500/10 p-3 flex items-start gap-2">
						<AlertCircle class="h-4 w-4 text-red-400 shrink-0 mt-0.5" />
						<p class="text-xs text-red-300">{error}</p>
					</div>
				{/if}

				{#if isLoggedIn}
					<!-- Scenario A: User is Logged In -->
					{#if isMatchingEmail}
						<div class="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-3 flex items-center justify-between text-xs">
							<span class="text-slate-400">Signed in as</span>
							<span class="font-bold text-emerald-400">{currentUserEmail}</span>
						</div>

						<button
							onclick={handleAccept}
							disabled={accepting}
							class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
						>
							{#if accepting}
								<Loader2 class="h-4 w-4 animate-spin" />
								Joining Organization...
							{:else}
								<UserCheck class="h-4 w-4" />
								Accept Invitation
							{/if}
						</button>
					{:else}
						<div class="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 space-y-2 text-xs text-amber-300">
							<p class="font-bold">Email Mismatch</p>
							<p>
								This invitation was sent to <strong class="text-white">{invite.email}</strong>, but you are currently signed in as <strong class="text-white">{currentUserEmail}</strong>.
							</p>
						</div>

						<button
							onclick={handleLogoutAndSwitch}
							class="flex w-full items-center justify-center gap-2 rounded-xl border border-slate-700 bg-slate-800 py-2.5 text-xs font-bold text-slate-200 hover:bg-slate-700 hover:text-white transition"
						>
							<LogOut class="h-4 w-4 text-slate-400" />
							Sign out to accept with {invite.email}
						</button>
					{/if}
				{:else}
					<!-- Scenario B: New User Onboarding Form -->
					<div class="space-y-4">
						<div class="space-y-1">
							<h3 class="text-sm font-bold text-white">Create Your Account</h3>
							<p class="text-xs text-slate-400">Set a password to complete registration and join {invite.tenant_name}</p>
						</div>

						<div class="space-y-3">
							<div class="space-y-1">
								<label for="invite-email-readonly" class="text-xs font-semibold text-slate-400">Account Email</label>
								<div class="relative">
									<Mail class="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
									<input
										id="invite-email-readonly"
										type="email"
										value={invite.email}
										readonly
										class="w-full rounded-xl border border-slate-800 bg-slate-950/80 py-2 pl-9 pr-3 text-xs text-slate-400 focus:outline-none cursor-not-allowed"
									/>
								</div>
							</div>

							<div class="space-y-1">
								<label for="onboarding-password" class="text-xs font-semibold text-slate-300">Create Password</label>
								<div class="relative">
									<KeyRound class="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
									<input
										id="onboarding-password"
										type="password"
										bind:value={password}
										placeholder="Choose password (min 6 characters)"
										class="w-full rounded-xl border border-slate-700 bg-slate-950 py-2 pl-9 pr-3 text-xs text-white placeholder-slate-500 focus:border-emerald-500 focus:outline-none"
									/>
								</div>
							</div>

							<div class="space-y-1">
								<label for="onboarding-confirm-password" class="text-xs font-semibold text-slate-300">Confirm Password</label>
								<div class="relative">
									<KeyRound class="absolute left-3 top-2.5 h-4 w-4 text-slate-500" />
									<input
										id="onboarding-confirm-password"
										type="password"
										bind:value={confirmPassword}
										placeholder="Confirm password"
										class="w-full rounded-xl border border-slate-700 bg-slate-950 py-2 pl-9 pr-3 text-xs text-white placeholder-slate-500 focus:border-emerald-500 focus:outline-none"
									/>
								</div>
							</div>
						</div>

						<button
							onclick={handleAccept}
							disabled={accepting}
							class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
						>
							{#if accepting}
								<Loader2 class="h-4 w-4 animate-spin" />
								Creating Account & Joining...
							{:else}
								<UserCheck class="h-4 w-4" />
								Accept & Join Organization
							{/if}
						</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

