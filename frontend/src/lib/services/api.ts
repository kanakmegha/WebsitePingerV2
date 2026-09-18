import type { Monitor, MonitorCheck, AuthResponse, TenantSettings, TenantDTO, AlertEvent } from '$lib/types';
import { tenantStore } from '$lib/stores/tenant';
import { authStore } from '$lib/stores/auth';
import { get } from 'svelte/store';

const API_BASE = 'http://localhost:8080/api';

export function getAuthToken(): string {
	if (typeof window !== 'undefined') {
		return localStorage.getItem('pinger_token') || '';
	}
	return '';
}

export function getSelectedTenantId(): string {
	const store = get(tenantStore);
	if (store.selectedTenantId) {
		return store.selectedTenantId;
	}
	if (typeof window !== 'undefined') {
		return localStorage.getItem('pinger_selected_tenant_id') || '';
	}
	return '';
}

let isRefreshingTenants = false;

// API Wrapper: Automatically injects Authorization Bearer token & X-Tenant-ID header
export async function apiFetch(endpoint: string, options: RequestInit = {}): Promise<Response> {
	const token = getAuthToken();
	const isAuthEndpoint = endpoint.startsWith('/auth/');

	if (token && !isAuthEndpoint) {
		const tenantId = getSelectedTenantId();
		if (!tenantId) {
			console.error(`[API FETCH ERROR] Attempted request to ${endpoint} without a selected tenant.`);
			throw new Error('Tenant not selected. Please select an organization.');
		}
	}

	const tenantId = getSelectedTenantId();
	const headers = new Headers(options.headers || {});
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}
	if (tenantId) {
		headers.set('X-Tenant-ID', tenantId);
	}
	if (!headers.has('Content-Type') && options.body && typeof options.body === 'string') {
		headers.set('Content-Type', 'application/json');
	}

	const res = await fetch(`${API_BASE}${endpoint}`, {
		...options,
		headers
	});

	if (res.status === 401) {
		console.warn(`[API FETCH] 401 Unauthorized on ${endpoint}. Logging out user.`);
		authStore.logout();
		tenantStore.reset();
	} else if (res.status === 403 && !isAuthEndpoint) {
		console.warn(`[API FETCH] 403 Forbidden on ${endpoint}. Triggering tenant validation recovery.`);
		if (!isRefreshingTenants) {
			isRefreshingTenants = true;
			try {
				const tenants = await fetchUserTenants();
				if (!tenants || tenants.length === 0) {
					console.warn('[API FETCH RECOVERY] User has no active tenant memberships. Logging out.');
					authStore.logout();
					tenantStore.reset();
				}
			} finally {
				isRefreshingTenants = false;
			}
		}
	}

	return res;
}

export const MOCK_ALERTS: AlertEvent[] = [
	{
		id: 'alt-1',
		alert_id: 'a-1',
		monitor_id: 'mon-101',
		monitor_name: 'PCAPPA Main Domain',
		status: 'triggered',
		message: 'HTTP response time exceeded 1000ms threshold (1500ms)',
		created_at: new Date(Date.now() - 3600000).toISOString()
	}
];

export const MOCK_MONITORS: Monitor[] = [
	{
		id: 'mon-101',
		tenant_id: 'ten-1',
		name: 'PCAPPA Main Domain',
		url: 'https://www.pcappa.org',
		domain: 'www.pcappa.org',
		check_configs: [
			{ id: 'c1', check_type: 'http', interval_seconds: 30, timeout_seconds: 10, is_enabled: true },
			{ id: 'c2', check_type: 'ssl', interval_seconds: 3600, timeout_seconds: 10, is_enabled: true },
			{ id: 'c3', check_type: 'dns', interval_seconds: 300, timeout_seconds: 10, is_enabled: true },
			{ id: 'c4', check_type: 'domain', interval_seconds: 86400, timeout_seconds: 10, is_enabled: true }
		],
		is_active: true,
		created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
		status: 'up',
		last_checked_at: new Date(Date.now() - 45000).toISOString(),
		avg_latency_ms: 42,
		ssl_days_remaining: 184,
		uptime_pct_24h: 99.98
	}
];

export async function loginUser(email: string, password: string): Promise<AuthResponse> {
	const res = await fetch(`${API_BASE}/auth/login`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || 'Invalid credentials');
	}

	const data: AuthResponse = await res.json();
	if (data.tenants && data.tenants.length > 0) {
		tenantStore.setTenants(data.tenants);
	}
	return data;
}

export async function registerUser(email: string, password: string): Promise<AuthResponse> {
	const res = await fetch(`${API_BASE}/auth/register`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password })
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || 'Registration failed');
	}

	const data: AuthResponse = await res.json();
	if (data.tenants && data.tenants.length > 0) {
		tenantStore.setTenants(data.tenants);
	}
	return data;
}

export async function fetchUserTenants(): Promise<TenantDTO[]> {
	try {
		const res = await apiFetch('/me/tenants');
		if (res.ok) {
			const tenants: TenantDTO[] = await res.json();
			tenantStore.setTenants(tenants);
			return tenants;
		}
	} catch (e) {
		console.warn('Failed to fetch user tenants:', e);
	}
	return [];
}

export async function fetchMonitors(): Promise<Monitor[]> {
	const token = getAuthToken();
	if (!token) return MOCK_MONITORS;

	try {
		const res = await apiFetch('/monitors');
		if (res.ok) {
			const data = await res.json();
			return data.map((m: any) => ({
				...m,
				status: m.is_active ? 'up' : 'down',
				avg_latency_ms: Math.floor(Math.random() * 80) + 20,
				ssl_days_remaining: 120,
				uptime_pct_24h: 99.9
			}));
		}
	} catch (e) {
		console.warn('Backend connection issue, serving fallback monitors:', e);
	}
	return MOCK_MONITORS;
}

export async function fetchMonitorById(id: string): Promise<Monitor | null> {
	try {
		const res = await apiFetch(`/monitors/${id}`);
		if (res.ok) {
			const m = await res.json();
			return {
				...m,
				status: m.is_active ? 'up' : 'down',
				avg_latency_ms: 42,
				ssl_days_remaining: 120,
				uptime_pct_24h: 99.9
			};
		}
	} catch (e) {
		console.warn('Failed to fetch monitor details:', e);
	}
	return null;
}

export async function fetchTenantSettings(): Promise<TenantSettings> {
	const token = getAuthToken();
	if (!token) {
		return {
			http_interval_seconds: 60,
			dns_interval_seconds: 300,
			ssl_interval_seconds: 3600,
			domain_interval_seconds: 86400,
			email_auth_interval_seconds: 300
		};
	}

	try {
		const res = await apiFetch('/settings');
		if (res.ok) {
			return await res.json();
		}
	} catch (e) {
		console.warn('Backend unavailable, using default tenant settings');
	}

	return {
		http_interval_seconds: 60,
		dns_interval_seconds: 300,
		ssl_interval_seconds: 3600,
		domain_interval_seconds: 86400,
		email_auth_interval_seconds: 300
	};
}

export async function updateTenantSettings(data: TenantSettings): Promise<TenantSettings> {
	const res = await apiFetch('/settings', {
		method: 'PUT',
		body: JSON.stringify(data)
	});

	const responseText = await res.text();
	if (!res.ok) {
		let errMessage = 'Failed to update settings';
		try {
			const errData = JSON.parse(responseText);
			errMessage = errData.error || errMessage;
		} catch (_) {}
		throw new Error(errMessage);
	}

	return JSON.parse(responseText);
}

export async function createMonitor(data: {
	name: string;
	url: string;
	checks: string[];
}): Promise<Monitor> {
	const res = await apiFetch('/monitors', {
		method: 'POST',
		body: JSON.stringify(data)
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || 'Failed to create monitor');
	}

	const m = await res.json();
	return {
		...m,
		status: 'up',
		last_checked_at: new Date().toISOString(),
		avg_latency_ms: 42,
		ssl_days_remaining: 90,
		uptime_pct_24h: 100.0
	};
}

export async function fetchChecks(monitorId: string): Promise<MonitorCheck[]> {
	try {
		const res = await apiFetch(`/monitors/${monitorId}/checks`);
		if (res.ok) {
			return await res.json();
		}
	} catch (e) {
		console.warn('Backend unavailable, serving mock check series');
	}

	const checks: MonitorCheck[] = [];
	const now = Date.now();
	for (let i = 0; i < 50; i++) {
		const isFail = i === 12 || i === 13;
		checks.push({
			id: `chk-${i}`,
			monitor_id: monitorId,
			check_type: 'http',
			status: isFail ? 'failure' : 'success',
			error: isFail ? 'HTTP 503 Service Unavailable' : undefined,
			checked_at: new Date(now - i * 3 * 60000).toISOString(),
			http_result: {
				status_code: isFail ? 503 : 200,
				response_time_ms: isFail ? 1500 : Math.floor(Math.random() * 60) + 25,
				response_size_bytes: 14250,
				final_url: 'https://google.com'
			},
			ssl_result: {
				expiry_date: new Date(now + 180 * 86400000).toISOString(),
				issuer: 'Google Trust Services LLC',
				valid: true,
				days_remaining: 180
			}
		});
	}
	return checks;
}
