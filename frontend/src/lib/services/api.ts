import type {
	AlertEvent,
	AuthResponse,
	InviteDetails,
	Monitor,
	MonitorCheck,
	MonitorStatus,
	PendingInvite,
	TenantDTO,
	TenantMember,
	TenantSettings,
} from "$lib/types";
import { tenantStore } from "$lib/stores/tenant";
import { authStore } from "$lib/stores/auth";
import { get } from "svelte/store";

const API_BASE = (import.meta.env.VITE_API_URL || "http://localhost:8080") + "/api";

export function getAuthToken(): string {
	if (typeof window !== "undefined") {
		return localStorage.getItem("pinger_token") || "";
	}
	return "";
}

export function getSelectedTenantId(): string {
	const store = get(tenantStore);
	if (store.selectedTenantId) {
		return store.selectedTenantId;
	}
	if (typeof window !== "undefined") {
		return localStorage.getItem("pinger_selected_tenant_id") || "";
	}
	return "";
}

let isRefreshingTenants = false;

// API Wrapper: Automatically injects Authorization Bearer token & X-Tenant-ID header
export async function apiFetch(
	endpoint: string,
	options: RequestInit = {},
): Promise<Response> {
	const token = getAuthToken();
	const isAuthEndpoint = endpoint.startsWith("/auth/");
	const isOrgOnboardingEndpoint = endpoint.startsWith("/orgs/") ||
		endpoint === "/me/tenants";

	if (token && !isAuthEndpoint && !isOrgOnboardingEndpoint) {
		const tenantId = getSelectedTenantId();
		if (!tenantId) {
			console.error(
				`[API FETCH ERROR] Attempted request to ${endpoint} without a selected tenant.`,
			);
			throw new Error(
				"Tenant not selected. Please select an organization.",
			);
		}
	}

	const tenantId = getSelectedTenantId();
	const headers = new Headers(options.headers || {});
	if (token) {
		headers.set("Authorization", `Bearer ${token}`);
	}
	if (tenantId) {
		headers.set("X-Tenant-ID", tenantId);
	}
	if (
		!headers.has("Content-Type") && options.body &&
		typeof options.body === "string"
	) {
		headers.set("Content-Type", "application/json");
	}

	const res = await fetch(`${API_BASE}${endpoint}`, {
		...options,
		headers,
	});

	if (res.status === 401 && !endpoint.startsWith("/orgs/accept")) {
		console.warn(
			`[API FETCH] 401 Unauthorized on ${endpoint}. Logging out user.`,
		);
		authStore.logout();
		tenantStore.reset();
	} else if (res.status === 403 && !isAuthEndpoint) {
		console.warn(
			`[API FETCH] 403 Forbidden on ${endpoint}. Triggering tenant validation recovery.`,
		);
		if (!isRefreshingTenants) {
			isRefreshingTenants = true;
			try {
				const tenants = await fetchUserTenants();
				if (!tenants || tenants.length === 0) {
					console.warn(
						"[API FETCH RECOVERY] User has no active tenant memberships. Logging out.",
					);
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
		id: "alt-1",
		alert_id: "a-1",
		monitor_id: "mon-101",
		monitor_name: "PCAPPA Main Domain",
		status: "triggered",
		message: "HTTP response time exceeded 1000ms threshold (1500ms)",
		created_at: new Date(Date.now() - 3600000).toISOString(),
	},
];

export const MOCK_MONITORS: Monitor[] = [
	{
		id: "mon-101",
		tenant_id: "ten-1",
		name: "PCAPPA Main Domain",
		url: "https://www.pcappa.org",
		domain: "www.pcappa.org",
		check_configs: [
			{
				id: "c1",
				check_type: "http",
				interval_seconds: 30,
				timeout_seconds: 10,
				is_enabled: true,
			},
			{
				id: "c2",
				check_type: "ssl",
				interval_seconds: 3600,
				timeout_seconds: 10,
				is_enabled: true,
			},
			{
				id: "c3",
				check_type: "dns",
				interval_seconds: 300,
				timeout_seconds: 10,
				is_enabled: true,
			},
			{
				id: "c4",
				check_type: "domain",
				interval_seconds: 86400,
				timeout_seconds: 10,
				is_enabled: true,
			},
		],
		is_active: true,
		created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
		status: "up",
		last_checked_at: new Date(Date.now() - 45000).toISOString(),
		avg_latency_ms: 42,
		ssl_days_remaining: 184,
		uptime_pct_24h: 99.98,
	},
];

export async function loginUser(
	email: string,
	password: string,
): Promise<AuthResponse> {
	const res = await fetch(`${API_BASE}/auth/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password }),
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Invalid credentials");
	}

	const data: AuthResponse = await res.json();
	if (data.tenants && data.tenants.length > 0) {
		tenantStore.setTenants(data.tenants);
	}
	return data;
}

export async function registerUser(
	email: string,
	password: string,
	tenantName?: string,
): Promise<AuthResponse> {
	const res = await fetch(`${API_BASE}/auth/register`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ email, password, tenant_name: tenantName }),
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Registration failed");
	}

	const data: AuthResponse = await res.json();
	if (data.tenants && data.tenants.length > 0) {
		tenantStore.setTenants(data.tenants);
	}
	return data;
}

export async function createOrganization(name: string): Promise<TenantDTO> {
	const res = await apiFetch("/orgs/create", {
		method: "POST",
		body: JSON.stringify({ name }),
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({}));
		throw new Error(err.error || "Failed to create organization");
	}
	const tenant: TenantDTO = await res.json();
	await fetchUserTenants();
	tenantStore.selectTenant(tenant.id);
	return tenant;
}

export async function joinOrganization(tenantId: string): Promise<TenantDTO> {
	const res = await apiFetch("/orgs/join", {
		method: "POST",
		body: JSON.stringify({ tenant_id: tenantId }),
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({}));
		throw new Error(err.error || "Failed to join organization");
	}
	const tenant: TenantDTO = await res.json();
	await fetchUserTenants();
	tenantStore.selectTenant(tenant.id);
	return tenant;
}

export async function fetchUserTenants(): Promise<TenantDTO[]> {
	try {
		const res = await apiFetch("/me/tenants");
		if (res.ok) {
			const tenants: TenantDTO[] = await res.json();
			tenantStore.setTenants(tenants);
			return tenants;
		}
	} catch (e) {
		console.warn("Failed to fetch user tenants:", e);
	}
	return [];
}

function mapMonitorFromBackend(m: any): Monitor {
	const checks: any[] = m.edges?.checks || m.edges?.Checks || [];
	const check_configs: any[] = m.check_configs || m.edges?.check_configs ||
		m.edges?.CheckConfigs || [];

	// Find latest HTTP check result
	const latestHttpCheck = checks.find((c) =>
		c.check_type === "http" &&
		(c.http_result || c.edges?.http_result || c.edges?.HTTPResult)
	);
	const httpRes = latestHttpCheck?.http_result ||
		latestHttpCheck?.edges?.http_result ||
		latestHttpCheck?.edges?.HTTPResult;

	// Find latest SSL check result
	const latestSslCheck = checks.find((c) =>
		c.check_type === "ssl" &&
		(c.ssl_result || c.edges?.ssl_result || c.edges?.SslResult)
	);
	const sslRes = latestSslCheck?.ssl_result ||
		latestSslCheck?.edges?.ssl_result || latestSslCheck?.edges?.SslResult;

	// Find latest DNS check result
	const latestDnsCheck = checks.find((c) =>
		c.check_type === "dns" &&
		(c.dns_result || c.edges?.dns_result || c.edges?.DNSResult)
	);
	const dnsRes = latestDnsCheck?.dns_result ||
		latestDnsCheck?.edges?.dns_result || latestDnsCheck?.edges?.DNSResult;

	// Find latest Domain check result
	const latestDomainCheck = checks.find((c) =>
		c.check_type === "domain" &&
		(c.domain_result || c.edges?.domain_result || c.edges?.DomainResult)
	);
	const domainRes = latestDomainCheck?.domain_result ||
		latestDomainCheck?.edges?.domain_result ||
		latestDomainCheck?.edges?.DomainResult;

	// Find most recent check timestamp across all checks
	const mostRecentCheck = checks.length > 0 ? checks[0] : null;

	const status: MonitorStatus = checks.length > 0
		? (checks[0].status === "success" ? "up" : "down")
		: (m.is_active ? "up" : "down");

	return {
		...m,
		check_configs,
		status,
		avg_latency_ms: httpRes?.response_time_ms ?? m.avg_latency_ms,
		ssl_days_remaining: sslRes?.days_remaining ?? m.ssl_days_remaining,
		last_checked_at: mostRecentCheck?.checked_at ?? m.last_checked_at,
		latest_ssl: sslRes,
		latest_dns: dnsRes,
		latest_domain: domainRes,
		latest_http: httpRes,
	};
}

export async function fetchMonitors(): Promise<Monitor[]> {
	try {
		const res = await apiFetch("/monitors");
		if (res.ok) {
			const data = await res.json();
			return data.map(mapMonitorFromBackend);
		}
	} catch (e) {
		console.warn("Backend connection issue fetching monitors:", e);
	}
	return [];
}

export async function fetchMonitorById(id: string): Promise<Monitor | null> {
	try {
		const res = await apiFetch(`/monitors/${id}`);
		if (res.ok) {
			const m = await res.json();
			return mapMonitorFromBackend(m);
		}
	} catch (e) {
		console.warn("Failed to fetch monitor details:", e);
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
			email_auth_interval_seconds: 300,
		};
	}

	try {
		const res = await apiFetch("/settings");
		if (res.ok) {
			return await res.json();
		}
	} catch (e) {
		console.warn("Backend unavailable, using default tenant settings");
	}

	return {
		http_interval_seconds: 60,
		dns_interval_seconds: 300,
		ssl_interval_seconds: 3600,
		domain_interval_seconds: 86400,
		email_auth_interval_seconds: 300,
	};
}

export async function updateTenantSettings(
	data: TenantSettings,
): Promise<TenantSettings> {
	const res = await apiFetch("/settings", {
		method: "PUT",
		body: JSON.stringify(data),
	});

	const responseText = await res.text();
	if (!res.ok) {
		let errMessage = "Failed to update settings";
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
	const res = await apiFetch("/monitors", {
		method: "POST",
		body: JSON.stringify(data),
	});

	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Failed to create monitor");
	}

	const m = await res.json();
	return {
		...m,
		status: "up",
		last_checked_at: new Date().toISOString(),
		avg_latency_ms: 42,
		ssl_days_remaining: 90,
		uptime_pct_24h: 100.0,
	};
}

export async function fetchChecks(
	monitorId: string,
	checkType?: string,
): Promise<MonitorCheck[]> {
	try {
		const url = checkType
			? `/monitors/${monitorId}/checks?type=${checkType}`
			: `/monitors/${monitorId}/checks`;
		const res = await apiFetch(url);
		if (res.ok) {
			const rawChecks = await res.json();
			return rawChecks.map((chk: any) => ({
				...chk,
				http_result: chk.http_result || chk.edges?.http_result ||
					chk.edges?.HTTPResult,
				ssl_result: chk.ssl_result || chk.edges?.ssl_result ||
					chk.edges?.SslResult,
				dns_result: chk.dns_result || chk.edges?.dns_result ||
					chk.edges?.DNSResult,
				domain_result: chk.domain_result || chk.edges?.domain_result ||
					chk.edges?.DomainResult,
			}));
		}
	} catch (e) {
		console.warn("Failed to fetch monitor checks from backend:", e);
	}
	return [];
}

export async function deleteMonitor(monitorId: string): Promise<void> {
	const res = await apiFetch(`/monitors/${monitorId}`, {
		method: "DELETE",
	});
	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Failed to delete monitor");
	}
}

export async function fetchInviteDetails(token: string): Promise<InviteDetails> {
	const res = await fetch(`${API_BASE}/orgs/invite/${token}`);
	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Invalid or expired invitation token");
	}
	return await res.json();
}

export async function acceptInvite(token: string, password?: string): Promise<{ token?: string; tenant_id: string; role: string; user_id?: string }> {
	const res = await apiFetch("/orgs/accept-invite", {
		method: "POST",
		body: JSON.stringify({ token, password }),
	});
	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Failed to accept invitation");
	}
	const data = await res.json();
	if (data.token) {
		authStore.login(data.token, data.user_id || "", data.email || "");
	}
	if (data.tenants && data.tenants.length > 0) {
		tenantStore.setTenants(data.tenants);
	}
	await fetchUserTenants();
	if (data.tenant_id) {
		tenantStore.selectTenant(data.tenant_id);
	}
	return data;
}

export async function inviteUser(email: string, role: string): Promise<PendingInvite> {
	const res = await apiFetch("/orgs/invite", {
		method: "POST",
		body: JSON.stringify({ email, role }),
	});
	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Failed to send invitation");
	}
	return await res.json();
}

export async function fetchTenantMembers(): Promise<TenantMember[]> {
	try {
		const res = await apiFetch("/orgs/members");
		if (res.ok) {
			return await res.json();
		}
	} catch (e) {
		console.warn("Failed to fetch tenant members:", e);
	}
	return [];
}

export async function fetchPendingInvites(): Promise<PendingInvite[]> {
	try {
		const res = await apiFetch("/orgs/invites");
		if (res.ok) {
			return await res.json();
		}
	} catch (e) {
		console.warn("Failed to fetch pending invites:", e);
	}
	return [];
}

export async function cancelInvite(id: string): Promise<void> {
	const res = await apiFetch(`/orgs/invites/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) {
		const errData = await res.json().catch(() => ({}));
		throw new Error(errData.error || "Failed to cancel invitation");
	}
}
