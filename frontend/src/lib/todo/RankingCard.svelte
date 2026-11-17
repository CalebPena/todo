<script lang="ts">
	import { onMount } from 'svelte';
	import type { RankingCardModel } from './ranking-scene';
	import type { RankingCardElements } from './ranking-timeline';
	import TodoDescription from './TodoDescription.svelte';

	let {
		card,
		active,
		busy,
		backLabel,
		register,
		swipeOffset = 0,
	}: {
		card: RankingCardModel;
		active: boolean;
		busy: boolean;
		swipeOffset?: number;
		backLabel: string;
		register: (key: string, elements: RankingCardElements) => () => void;
	} = $props();
	let root: HTMLDivElement;
	let flipper: HTMLDivElement;
	let title: HTMLHeadingElement;
	onMount(() => register(card.key, { root, flipper, title }));
</script>

<div
	bind:this={root}
	class="ranking-card"
	class:active={active && !busy}
	class:swiping={swipeOffset !== 0}
	style:translate={`${swipeOffset}px 0`}
	data-category={card.category}
	data-ranking-card={card.key}
	aria-hidden={!active || busy}
	inert={!active || busy}
>
	<div bind:this={flipper} class="flipper">
		<div class="face front">
			<h2 bind:this={title}>{card.todo.name}</h2>
			<TodoDescription descrption={card.todo.description} />
		</div>
		<div class="face back" aria-hidden="true">
			<p>{backLabel}</p>
			<span>than your task</span>
		</div>
	</div>
</div>

<style>
	.ranking-card {
		--rank-color: #be123c;
		--rank-tint: #fff1f2;
		position: absolute;
		top: 0;
		left: 0;
		width: 176px;
		height: 190px;
		visibility: hidden;
		pointer-events: none;
		perspective: 1000px;
		transform-origin: center;
		transition: translate 180ms ease-out;
	}
	.ranking-card.swiping {
		transition: none;
	}
	@media (prefers-reduced-motion: reduce) {
		.ranking-card {
			transition: none;
		}
	}
	.ranking-card[data-category='time'] {
		--rank-color: #1d4ed8;
		--rank-tint: #eff6ff;
	}
	:global(.dark) .ranking-card {
		--rank-color: #fda4af;
		--rank-tint: #31212a;
	}
	:global(.dark) .ranking-card[data-category='time'] {
		--rank-color: #93c5fd;
		--rank-tint: #1f2b40;
	}
	.ranking-card.active {
		pointer-events: auto;
	}
	.flipper {
		position: relative;
		width: 100%;
		height: 100%;
		transform-style: preserve-3d;
	}
	.face {
		position: absolute;
		inset: 0;
		overflow: hidden;
		padding: 22px;
		border: 1px solid var(--rank-color);
		border-radius: 12px;
		background: var(--rank-tint);
		box-shadow: 0 20px 48px -14px #00000030;
		backface-visibility: hidden;
	}
	.active .front {
		overflow-y: auto;
	}
	h2 {
		color: var(--rank-color);
		margin-bottom: 10px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.back {
		transform: rotateY(180deg);
		border-color: color-mix(in srgb, var(--rank-color) 40%, var(--border));
		background: color-mix(in srgb, var(--rank-color) 10%, var(--rank-tint));
	}
	.back p {
		color: var(--rank-color);
		font-size: 14px;
		font-weight: 650;
	}
	.back span {
		display: block;
		margin-top: 4px;
		color: var(--muted-foreground);
		font-size: 12px;
	}
</style>
