<script lang="ts">
	import type { Todo } from './todo';
	import TodoDescription from './TodoDescription.svelte';

	let {
		todos,
		onselect,
		hoveredId = null,
	}: { todos: Todo[]; onselect: (id: number) => void; hoveredId?: number | null } = $props();
	const id = $props.id();
	let hovered = $state<Todo | null>(null);
	let focused = $state<Todo | null>(null);
	let ranked = $derived(
		todos.filter(
			(todo) => !todo.completed && todo.ranked && todo.valueRank !== null && todo.timeRank !== null,
		),
	);
	let count = $derived(ranked.length);
	let preview = $derived(
		hovered ?? focused ?? ranked.find((todo) => todo.id === hoveredId) ?? null,
	);
	const fraction = (rank: number, total: number) => (total <= 1 ? 0.5 : (rank - 1) / (total - 1));
	function point(todo: Todo) {
		const time = fraction(todo.timeRank!, count);
		const value = 1 - fraction(todo.valueRank!, count);
		return {
			x: 40 + time * 560,
			y: 278 - value * 250,
			color: `color-mix(in srgb, #e11d48 ${Math.round(((0.2 + value) / (0.4 + value + time)) * 100)}%, #2563eb)`,
		};
	}
</script>

{#if count > 0}
	<section class="relative mt-6 w-full" aria-label="Task value and time graph">
		<svg
			viewBox="0 0 620 330"
			class="graph w-full"
			role="img"
			aria-label="Tasks by time and value rank"
			aria-describedby={`${id}-description`}
		>
			<desc id={`${id}-description`}
				>Less time on the left, more time on the right. More valuable tasks are higher. Rose
				represents value and blue represents time. Select a dot to highlight its task at the top of
				the list.</desc
			>
			<rect x="28" y="16" width="290" height="139" fill="#a855f7" fill-opacity="0.07" />
			<path d="M28 155H608 M318 16V290" class="grid-line" />
			<path d="M28 16V290H608 M24 22L28 16L32 22 M602 286L608 290L602 294" class="axis-line" />
			<text x="318" y="316" text-anchor="middle" class="axis-label">Time →</text>
			<text x="14" y="155" transform="rotate(-90 14 155)" text-anchor="middle" class="axis-label"
				>Value →</text
			>
			{#each ranked as todo (todo.id)}
				{@const position = point(todo)}
				<a
					class:highlighted={todo.id === hoveredId}
					href={`/${todo.id}`}
					onclick={(event) => {
						if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return;
						event.preventDefault();
						onselect(todo.id);
					}}
					aria-label={`${todo.name}: value rank ${todo.valueRank}, time rank ${todo.timeRank}`}
					aria-describedby={preview?.id === todo.id ? `${id}-preview` : undefined}
					onpointerenter={() => (hovered = todo)}
					onpointerleave={() => (hovered = null)}
					onfocus={() => (focused = todo)}
					onblur={() => (focused = null)}
					onkeydown={(event) => {
						if (event.key === 'Escape') {
							hovered = null;
							focused = null;
						}
					}}
				>
					<circle cx={position.x} cy={position.y} r="18" fill="transparent" />
					<circle class="point" cx={position.x} cy={position.y} r="7" fill={position.color} />
				</a>
			{/each}
		</svg>
		{#if preview}
			{@const position = point(preview)}
			<div
				id={`${id}-preview`}
				role="tooltip"
				class="preview pointer-events-none absolute z-10 rounded-lg border bg-popover p-3 text-popover-foreground shadow-lg"
				style:left={`clamp(8rem, ${(position.x / 620) * 100}%, calc(100% - 8rem))`}
				style:top={`${(position.y / 330) * 100}%`}
				class:above={position.y > 165}
			>
				<p class="mb-1 truncate text-sm font-semibold">{preview.name}</p>
				<TodoDescription descrption={preview.description} lineClamp={2} />
			</div>
		{/if}
	</section>
{/if}

<style>
	.preview {
		width: min(16rem, calc(100% - 16px));
		transform: translate(-50%, 16px);
	}
	.preview.above {
		transform: translate(-50%, calc(-100% - 16px));
	}
	.graph {
		display: block;
		width: 100%;
		height: auto;
		overflow: visible;
		background: white;
		border-radius: 12px;
	}
	.grid-line {
		stroke: #64748b;
		stroke-opacity: 0.2;
		stroke-dasharray: 4 6;
		fill: none;
	}
	.axis-line {
		stroke: #64748b;
		stroke-width: 1.5;
		fill: none;
	}
	.axis-label {
		fill: #475569;
		font-size: 11px;
	}
	.point {
		stroke: white;
		stroke-width: 2;
		transition: r 150ms ease;
	}
	a.highlighted .point,
	a:hover .point,
	a:focus-visible .point {
		r: 10;
	}
	a:focus-visible {
		outline: none;
	}
	a:focus-visible .point {
		stroke: var(--ring);
		stroke-width: 3px;
	}
	@media (prefers-reduced-motion: reduce) {
		.point {
			transition: none;
		}
	}
</style>
