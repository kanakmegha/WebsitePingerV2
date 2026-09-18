<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Chart from 'chart.js/auto';

	let { period = '24h' }: { period?: '24h' | '7d' | '30d' } = $props();

	let canvas: HTMLCanvasElement;
	let chartInstance: Chart | null = null;

	function generateChartData(p: string) {
		const points = p === '24h' ? 24 : p === '7d' ? 28 : 30;
		const labels: string[] = [];
		const latencies: number[] = [];

		const now = new Date();
		for (let i = points; i >= 0; i--) {
			if (p === '24h') {
				const d = new Date(now.getTime() - i * 3600000);
				labels.push(`${d.getHours()}:00`);
			} else {
				const d = new Date(now.getTime() - i * 86400000);
				labels.push(`${d.getMonth() + 1}/${d.getDate()}`);
			}
			const base = 35 + Math.floor(Math.random() * 20);
			const spike = i === 5 ? 180 : 0;
			latencies.push(base + spike);
		}

		return { labels, latencies };
	}

	function renderChart() {
		if (!canvas) return;
		if (chartInstance) chartInstance.destroy();

		const { labels, latencies } = generateChartData(period);
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
						pointRadius: 2,
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
		if (period && canvas) {
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
