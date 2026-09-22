export type CheckType = 'http' | 'ssl' | 'dns' | 'whois' | 'domain' | 'email_auth';
export type MonitorStatus = 'up' | 'down' | 'warning' | 'pending';
export type Role = 'owner' | 'admin' | 'member';

export interface TenantSettings {
	id?: string;
	tenant_id?: string;
	http_interval_seconds: number;
	dns_interval_seconds: number;
	ssl_interval_seconds: number;
	domain_interval_seconds: number;
	email_auth_interval_seconds: number;
	ssl_min_expiry_days?: number;
	domain_min_expiry_days?: number;
	updated_at?: string;
}

export interface CheckItemPayload {
	type: 'http' | 'dns' | 'ssl' | 'domain' | 'email_auth';
	interval_seconds: number;
	timeout_seconds?: number;
}

export interface MonitorCheckConfig {
	id: string;
	check_type: CheckType;
	interval_seconds: number;
	timeout_seconds: number;
	is_enabled: boolean;
	last_checked_at?: string;
}

export interface HTTPCheckResult {
	status_code: number;
	response_time_ms: number;
	response_size_bytes: number;
	final_url: string;
}

export interface SSLCheckResult {
	expiry_date: string;
	issuer: string;
	valid: boolean;
	days_remaining: number;
}

export interface DomainCheckResult {
	expiry_date: string;
	registrar: string;
	days_remaining: number;
}

export interface DNSCheckResult {
	has_a_record: boolean;
	has_aaaa_record: boolean;
	a_records?: string[];
	aaaa_records?: string[];
	mx_records?: string[];
	txt_records?: string[];
	spf_valid: boolean;
	dmarc_valid: boolean;
}

export interface MonitorCheck {
	id: string;
	monitor_id: string;
	check_type: CheckType;
	status: 'success' | 'failure' | 'rate_limited';
	error?: string;
	checked_at: string;
	http_result?: HTTPCheckResult;
	ssl_result?: SSLCheckResult;
	domain_result?: DomainCheckResult;
	dns_result?: DNSCheckResult;
}

export interface Monitor {
	id: string;
	tenant_id: string;
	name: string;
	url: string;
	domain: string;
	type?: string;
	check_configs?: MonitorCheckConfig[];
	interval_seconds?: number;
	timeout_seconds?: number;
	is_active: boolean;
	created_at: string;
	status: MonitorStatus;
	last_checked_at?: string;
	avg_latency_ms?: number;
	ssl_days_remaining?: number;
	uptime_pct_24h?: number;
	latest_ssl?: SSLCheckResult;
	latest_dns?: DNSCheckResult;
	latest_domain?: DomainCheckResult;
	latest_http?: HTTPCheckResult;
}

export type AlertType = 'incident' | 'recovery' | 'invite_sent' | 'invite_accepted';

export interface AlertEvent {
	id: string;
	tenant_id: string;
	monitor_id?: string;
	alert_id?: string;
	type: AlertType;
	message: string;
	status: 'unread' | 'read';
	created_at: string;
}

export interface NotificationChannel {
	id: string;
	tenant_id: string;
	type: 'email' | 'webhook' | 'slack' | 'pagerduty';
	config: Record<string, any>;
	created_at: string;
}

export interface User {
	id: string;
	email: string;
}

export interface TenantDTO {
	id: string;
	name: string;
	role: string;
}

export interface AuthResponse {
	token: string;
	user_id: string;
	tenant_id?: string;
	tenants?: TenantDTO[];
}

export interface InviteDetails {
	id: string;
	email: string;
	tenant_id: string;
	tenant_name: string;
	role: string;
	status: string;
	expires_at: string;
}

export interface TenantMember {
	id: string;
	user_id: string;
	email: string;
	role: string;
	created_at: string;
}

export interface PendingInvite {
	id: string;
	tenant_id: string;
	email: string;
	role: string;
	status: string;
	expires_at: string;
	created_at: string;
}
