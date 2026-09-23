// Service Worker for Pinger PWA + Web Push Notifications & App Badging API

self.addEventListener('install', (event) => {
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil(clients.claim());
});

// Push Notification & App Badging Handler
self.addEventListener('push', (event) => {
	let data = {
		title: 'Pinger Alert',
		body: 'New notification received',
		url: '/alerts',
		unread_count: 1
	};

	if (event.data) {
		try {
			data = event.data.json();
		} catch (e) {
			data.body = event.data.text();
		}
	}

	// Update App Badge on home screen icon if supported
	const count = typeof data.unread_count === 'number' ? data.unread_count : 1;
	if ('setAppBadge' in navigator) {
		navigator.setAppBadge(count).catch((err) => {
			console.warn('Failed setting app badge from SW:', err);
		});
	} else if ('setAppBadge' in self.navigator) {
		// @ts-ignore
		self.navigator.setAppBadge(count).catch(() => {});
	}

	const options = {
		body: data.body,
		icon: '/icons/icon-192.png',
		badge: '/icons/icon-192.png',
		tag: 'pinger-alert-notification',
		renotify: true,
		requireInteraction: true,
		vibrate: [200, 100, 200, 100, 200],
		data: {
			url: data.url || '/alerts'
		}
	};

	event.waitUntil(
		self.registration.showNotification(data.title, options)
	);
});

// Notification Click Handler
self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	const targetUrl = event.notification.data?.url || '/alerts';

	// Clear App Badge when user opens notification
	if ('clearAppBadge' in navigator) {
		navigator.clearAppBadge().catch(() => {});
	}

	event.waitUntil(
		clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
			for (const client of clientList) {
				if (client.url.includes(targetUrl) && 'focus' in client) {
					return client.focus();
				}
			}
			if (clients.openWindow) {
				return clients.openWindow(targetUrl);
			}
		})
	);
});
