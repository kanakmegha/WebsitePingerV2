<script lang="ts">
	import type { AlertEvent } from '$lib/types';
	import { alertStore } from '$lib/stores/alerts';
	import { ShieldAlert, CheckCircle2, CheckCheck, Mail, UserCheck } from '@lucide/svelte';

	let { alerts = [] }: { alerts?: AlertEvent[] } = $props();

	function handleMarkRead(id: string) {
		alertStore.markRead(id);
	}

	function formatType(type: string): string {
		if (type === 'invite_sent') return 'Invite Sent';
		if (type === 'invite_accepted') return 'Invite Accepted';
		return type;
	}
</script>

<div class="space-y-3">
	{#if alerts.length === 0}
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 p-12 text-center backdrop-blur">
			<CheckCircle2 class="mx-auto h-10 w-10 text-emerald-400 opacity-60" />
			<h4 class="mt-3 text-base font-bold text-slate-200">No Alerts Found</h4>
			<p class="mt-1 text-xs text-slate-400">All monitored targets and org events are logged here.</p>
		</div>
	{:else}
		{#each alerts as alt (alt.id)}
			<div
				class="flex items-start justify-between gap-4 rounded-xl border p-4 backdrop-blur transition hover:border-slate-700 {alt.type === 'incident'
					? 'border-red-500/30 bg-red-950/20 text-red-200'
					: alt.type === 'invite_sent'
					? 'border-blue-500/30 bg-blue-950/20 text-blue-200'
					: alt.type === 'invite_accepted'
					? 'border-cyan-500/30 bg-cyan-950/20 text-cyan-200'
					: 'border-emerald-500/30 bg-emerald-950/20 text-emerald-200'}"
			>
				<div class="flex items-start gap-3.5">
					<div
						class="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border {alt.type === 'incident'
							? 'border-red-500/30 bg-red-500/10 text-red-400'
							: alt.type === 'invite_sent'
							? 'border-blue-500/30 bg-blue-500/10 text-blue-400'
							: alt.type === 'invite_accepted'
							? 'border-cyan-500/30 bg-cyan-500/10 text-cyan-400'
							: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-400'}"
					>
						{#if alt.type === 'incident'}
							<ShieldAlert class="h-5 w-5" />
						{:else if alt.type === 'invite_sent'}
							<Mail class="h-5 w-5" />
						{:else if alt.type === 'invite_accepted'}
							<UserCheck class="h-5 w-5" />
						{:else}
							<CheckCircle2 class="h-5 w-5" />
						{/if}
					</div>

					<div>
						<div class="flex items-center gap-2">
							<span
								class="rounded-full px-2.5 py-0.5 text-[10px] font-bold uppercase tracking-wider {alt.type === 'incident'
									? 'bg-red-500/20 text-red-400 border border-red-500/30'
									: alt.type === 'invite_sent'
									? 'bg-blue-500/20 text-blue-400 border border-blue-500/30'
									: alt.type === 'invite_accepted'
									? 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30'
									: 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'}"
							>
								{formatType(alt.type)}
							</span>
							{#if alt.status === 'unread'}
								<span class="inline-flex items-center gap-1 rounded-full bg-amber-500/20 px-2 py-0.5 text-[10px] font-bold text-amber-400 border border-amber-500/30">
									<span class="h-1.5 w-1.5 rounded-full bg-amber-400 animate-pulse"></span>
									UNREAD
								</span>
							{/if}
						</div>
						<p class="mt-1.5 font-mono text-xs text-slate-200">{alt.message}</p>
					</div>
				</div>

				<div class="flex flex-col items-end gap-2 shrink-0">
					<span class="font-mono text-xs text-slate-400">
						{new Date(alt.created_at).toLocaleString()}
					</span>
					{#if alt.status === 'unread'}
						<button
							onclick={() => handleMarkRead(alt.id)}
							class="inline-flex items-center gap-1 text-[11px] font-semibold text-slate-400 hover:text-emerald-400 transition"
							title="Mark as read"
						>
							<CheckCheck class="h-3.5 w-3.5" />
							Mark read
						</button>
					{/if}
				</div>
			</div>
		{/each}
	{/if}
</div>
