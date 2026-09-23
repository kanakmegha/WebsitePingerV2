import { apiFetch } from './api';

export interface PushDiagnosticReport {
	isSecureContext: boolean;
	hasServiceWorker: boolean;
	hasPushManager: boolean;
	hasNotification: boolean;
	isIOS: boolean;
	isIOSStandalone: boolean;
	protocol: string;
	hostname: string;
	status: 'supported' | 'insecure_context' | 'ios_home_screen_required' | 'api_unsupported';
	reason: string;
	recommendation: string;
}

export function getPushDiagnostics(): PushDiagnosticReport {
	if (typeof window === 'undefined') {
		return {
			isSecureContext: false,
			hasServiceWorker: false,
			hasPushManager: false,
			hasNotification: false,
			isIOS: false,
			isIOSStandalone: false,
			protocol: '',
			hostname: '',
			status: 'api_unsupported',
			reason: 'Server-side rendering context',
			recommendation: 'Run in browser environment'
		};
	}

	const isSecureContext = window.isSecureContext ?? false;
	const hasServiceWorker = 'serviceWorker' in navigator;
	const hasPushManager = 'PushManager' in window;
	const hasNotification = 'Notification' in window;
	const protocol = window.location.protocol;
	const hostname = window.location.hostname;

	const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream;
	const isIOSStandalone = isIOS && ('standalone' in window.navigator) && (window.navigator as any).standalone === true;

	console.log('[PUSH DIAGNOSTICS]', {
		isSecureContext,
		hasServiceWorker,
		hasPushManager,
		hasNotification,
		protocol,
		hostname,
		isIOS,
		isIOSStandalone
	});

	if (!isSecureContext) {
		console.error(`[PUSH ERROR] Insecure context! window.isSecureContext is false on ${window.location.href}. Web Push requires HTTPS or localhost.`);
		return {
			isSecureContext,
			hasServiceWorker,
			hasPushManager,
			hasNotification,
			isIOS,
			isIOSStandalone,
			protocol,
			hostname,
			status: 'insecure_context',
			reason: `Accessed via HTTP on IP address (${hostname}). Browsers disable Web Push on non-secure origins.`,
			recommendation: `Access via http://localhost:4001 or set up HTTPS for ${hostname}.`
		};
	}

	if (isIOS && !isIOSStandalone && !hasPushManager) {
		return {
			isSecureContext,
			hasServiceWorker,
			hasPushManager,
			hasNotification,
			isIOS,
			isIOSStandalone,
			protocol,
			hostname,
			status: 'ios_home_screen_required',
			reason: 'iOS Safari requires adding Pinger to Home Screen to enable Web Push notifications.',
			recommendation: 'Tap Share icon in Safari → select "Add to Home Screen" → launch Pinger from Home Screen.'
		};
	}

	if (!hasServiceWorker || !hasPushManager || !hasNotification) {
		const missing: string[] = [];
		if (!hasServiceWorker) missing.push('Service Worker API');
		if (!hasPushManager) missing.push('PushManager API');
		if (!hasNotification) missing.push('Notification API');

		return {
			isSecureContext,
			hasServiceWorker,
			hasPushManager,
			hasNotification,
			isIOS,
			isIOSStandalone,
			protocol,
			hostname,
			status: 'api_unsupported',
			reason: `Missing browser APIs: ${missing.join(', ')}`,
			recommendation: 'Use a modern browser (Chrome, Edge, Firefox, or iOS 16.4+ PWA).'
		};
	}

	return {
		isSecureContext,
		hasServiceWorker,
		hasPushManager,
		hasNotification,
		isIOS,
		isIOSStandalone,
		protocol,
		hostname,
		status: 'supported',
		reason: 'Web Push API is fully supported in this environment.',
		recommendation: 'Ready for push notification subscription.'
	};
}

export function isPushSupported(): boolean {
	const diag = getPushDiagnostics();
	return diag.status === 'supported';
}

export function getNotificationPermission(): NotificationPermission | 'unsupported' {
	if (typeof window === 'undefined' || !('Notification' in window)) return 'unsupported';
	return Notification.permission;
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
	const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
	const rawData = window.atob(base64);
	const outputArray = new Uint8Array(rawData.length);
	for (let i = 0; i < rawData.length; ++i) {
		outputArray[i] = rawData.charCodeAt(i);
	}
	return outputArray;
}

export async function fetchVAPIDPublicKey(): Promise<string> {
	const res = await apiFetch('/push/vapid-key');
	if (!res.ok) {
		throw new Error('Failed to fetch VAPID public key from backend server.');
	}
	const data = await res.json();
	return data.public_key || '';
}

export async function registerServiceWorker(): Promise<ServiceWorkerRegistration> {
	try {
		const reg = await navigator.serviceWorker.register('/sw.js');
		console.log('[SW REGISTERED] Service Worker registered successfully, scope:', reg.scope);
		await navigator.serviceWorker.ready;
		return reg;
	} catch (err) {
		console.error('[SW REGISTRATION FAILED] Failed to register /sw.js:', err);
		throw err;
	}
}

export async function checkCurrentSubscription(): Promise<PushSubscription | null> {
	const diag = getPushDiagnostics();
	if (diag.status !== 'supported') return null;

	try {
		const reg = await registerServiceWorker();
		return await reg.pushManager.getSubscription();
	} catch (e) {
		console.warn('Failed checking push subscription:', e);
		return null;
	}
}

export async function subscribeUserToPush(): Promise<boolean> {
	const diag = getPushDiagnostics();
	if (diag.status !== 'supported') {
		throw new Error(diag.reason + ' ' + diag.recommendation);
	}

	const permission = await Notification.requestPermission();
	if (permission !== 'granted') {
		throw new Error('Notification permission denied by user.');
	}

	const reg = await registerServiceWorker();

	const publicKey = await fetchVAPIDPublicKey();
	if (!publicKey) {
		throw new Error('VAPID public key unavailable from server.');
	}

	const applicationServerKey = urlBase64ToUint8Array(publicKey);
	const subscription = await reg.pushManager.subscribe({
		userVisibleOnly: true,
		applicationServerKey: applicationServerKey as unknown as BufferSource
	});

	const jsonSub = subscription.toJSON();
	const endpoint = jsonSub.endpoint || '';
	const p256dh = jsonSub.keys?.p256dh || '';
	const auth = jsonSub.keys?.auth || '';

	const res = await apiFetch('/push/subscribe', {
		method: 'POST',
		body: JSON.stringify({ endpoint, p256dh, auth })
	});

	if (!res.ok) {
		const err = await res.json().catch(() => ({}));
		throw new Error(err.error || 'Failed to save push subscription on server.');
	}

	return true;
}

export async function unsubscribeUserFromPush(): Promise<boolean> {
	const diag = getPushDiagnostics();
	if (diag.status !== 'supported') return false;

	try {
		const reg = await registerServiceWorker();
		const subscription = await reg.pushManager.getSubscription();
		if (subscription) {
			const endpoint = subscription.endpoint;
			await subscription.unsubscribe();
			await apiFetch('/push/unsubscribe', {
				method: 'POST',
				body: JSON.stringify({ endpoint })
			});
		}
		return true;
	} catch (e) {
		console.error('Failed unsubscribing push notification:', e);
		return false;
	}
}
