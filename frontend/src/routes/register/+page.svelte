<script lang="ts">
	import { goto } from '$app/navigation';
	import { registerUser } from '$lib/services/api';
	import { authStore } from '$lib/stores/auth';
	import { toastStore } from '$lib/stores/toast';
	import { Activity, UserPlus, AlertCircle, Lock, Mail } from '@lucide/svelte';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!email || !password) {
			errorMessage = 'Please enter both email and password';
			return;
		}

		if (password !== confirmPassword) {
			errorMessage = 'Passwords do not match';
			return;
		}

		loading = true;
		errorMessage = '';

		try {
			const res = await registerUser(email, password);
			authStore.login(res.token, res.user_id, email);
			toastStore.show('Account created successfully!', 'success');
			goto('/');
		} catch (err: any) {
			errorMessage = err.message || 'Registration failed';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Register Account | Pinger</title>
</svelte:head>

<div class="mx-auto max-w-md space-y-8 py-12">
	<div class="text-center space-y-2">
		<div class="inline-flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-tr from-emerald-500 to-cyan-500 shadow-xl shadow-emerald-500/20">
			<Activity class="h-6 w-6 text-slate-950" />
		</div>
		<h1 class="text-2xl font-extrabold text-white">Create Tenant Account</h1>
		<p class="text-xs text-slate-400">Start monitoring HTTP, SSL, DNS, & WHOIS for free</p>
	</div>

	<form onsubmit={handleSubmit} class="rounded-xl border border-slate-800 bg-slate-900/60 p-6 backdrop-blur shadow-xl space-y-5">
		{#if errorMessage}
			<div class="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-red-400">
				<AlertCircle class="h-4 w-4 shrink-0" />
				<span>{errorMessage}</span>
			</div>
		{/if}

		<div>
			<label for="reg-email" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
				Email Address
			</label>
			<div class="relative mt-2">
				<Mail class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
				<input
					id="reg-email"
					type="email"
					required
					bind:value={email}
					placeholder="admin@company.com"
					class="w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
				/>
			</div>
		</div>

		<div>
			<label for="reg-pass" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
				Password
			</label>
			<div class="relative mt-2">
				<Lock class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
				<input
					id="reg-pass"
					type="password"
					required
					bind:value={password}
					placeholder="••••••••••••"
					class="w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
				/>
			</div>
		</div>

		<div>
			<label for="reg-pass-confirm" class="block text-xs font-bold uppercase tracking-wider text-slate-300">
				Confirm Password
			</label>
			<div class="relative mt-2">
				<Lock class="absolute left-3.5 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
				<input
					id="reg-pass-confirm"
					type="password"
					required
					bind:value={confirmPassword}
					placeholder="••••••••••••"
					class="w-full rounded-xl border border-slate-800 bg-slate-950 py-2.5 pl-10 pr-4 text-xs font-medium text-slate-100 placeholder-slate-600 focus:border-emerald-500 focus:outline-none"
				/>
			</div>
		</div>

		<button
			type="submit"
			disabled={loading}
			class="flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 py-3 text-xs font-bold text-slate-950 shadow-lg shadow-emerald-500/20 transition hover:bg-emerald-400 active:scale-95 disabled:opacity-50"
		>
			<UserPlus class="h-4 w-4" />
			{loading ? 'Creating Account...' : 'Register Organization'}
		</button>

		<div class="text-center pt-2">
			<p class="text-xs text-slate-400">
				Already have an account?
				<a href="/login" class="font-bold text-emerald-400 hover:underline ml-1">Sign In</a>
			</p>
		</div>
	</form>
</div>
