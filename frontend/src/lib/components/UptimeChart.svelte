<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';
	import type { MonitorCheck } from '$lib/types';

	let { period = '24h', checks = [] }: { period?: '24h' | '7d' | '30d'; checks?: MonitorCheck[] } = $props();

	let canvas: HTMLCanvasElement;
	let chartInstance: Chart | null = null;

	function extractChartData() {
		// Filter HTTP checks that contain http_result
		const httpChecks = checks
			.filter((c) => c.check_type === 'http' && c.http_result)
			.slice()
			.reverse(); // Chronological order

		if (httpChecks.length === 0) {
			return { labels: ['No Check Data'], latencies: [0] };
		}

		const labels = httpChecks.map((c) => {
			const d = new Date(c.checked_at);
			return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		});

		const latencies = httpChecks.map((c) => c.http_result?.response_time_ms || 0);

		return { labels, latencies };
	}

	function renderChart() {
		if (!canvas) return;
		if (chartInstance) chartInstance.destroy();

		const { labels, latencies } = extractChartData();
		const ctx = canvas.getContext('2d');
		if (!ctx) return;

		const gradient = ctx.createLinearGradient(0, 0, 0, 300);
		gradient.addColorStop(0, 'rgba(16, 185, 129, 0.35)');
		gradient.addColorStop(1, 'rgba(16, 185, 129, 0.0)');

		chartInstance = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets: [
					{
						label: 'Latency (ms)',
						data: latencies,
						borderColor: '#10b981',
						borderWidth: 2,
						backgroundColor: gradient,
						fill: true,
						tension: 0.35,
						pointRadius: 3,
						pointHoverRadius: 6,
						pointBackgroundColor: '#10b981'
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						mode: 'index',
						intersect: false,
						backgroundColor: '#0f172a',
						titleColor: '#f8fafc',
						bodyColor: '#34d399',
						borderColor: '#1e293b',
						borderWidth: 1,
						padding: 12,
						displayColors: false
					}
				},
				scales: {
					x: {
						grid: { display: false },
						ticks: { color: '#64748b', font: { size: 11 } }
					},
					y: {
						grid: { color: 'rgba(30, 41, 59, 0.5)' },
						ticks: { color: '#64748b', font: { size: 11 } },
						suggestedMin: 0
					}
				}
			}
		});
	}

	$effect(() => {
		if ((period || checks) && canvas) {
			renderChart();
		}
	});

	onMount(() => {
		renderChart();
	});

	onDestroy(() => {
		if (chartInstance) chartInstance.destroy();
	});
</script>

<div class="relative h-64 w-full">
	<canvas bind:this={canvas}></canvas>
</div>
