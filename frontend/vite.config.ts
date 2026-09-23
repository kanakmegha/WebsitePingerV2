import tailwindcss from "@tailwindcss/vite";
import adapter from "@sveltejs/adapter-auto";
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		VitePWA({
			registerType: "autoUpdate",
			injectRegister: "auto",
			includeAssets: [
				"favicon.png",
				"icons/icon-192.png",
				"icons/icon-512.png",
			],
			manifest: {
				name: "Pinger LIVE",
				short_name: "Pinger LIVE",
				description:
					"Real-time Website Monitoring & State-Driven Alert System",
				start_url: "/",
				scope: "/",
				display: "standalone",
				background_color: "#0f172a",
				theme_color: "#0f172a",
				icons: [
					{
						src: "/icons/icon-192.png",
						sizes: "192x192",
						type: "image/png",
						purpose: "any maskable",
					},
					{
						src: "/icons/icon-512.png",
						sizes: "512x512",
						type: "image/png",
						purpose: "any maskable",
					},
				],
			},
			devOptions: {
				enabled: true,
			},
		}),
	],
	server: {
		host: "0.0.0.0",
		port: 4001,
		allowedHosts: [".trycloudflare.com", "192.168.1.58", "localhost"],
	},
	preview: {
		host: "0.0.0.0",
		port: 4001,
		allowedHosts: [".trycloudflare.com", "192.168.1.58", "localhost"],
	},
});
