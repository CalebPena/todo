<script lang="ts">
	import { applyAction, enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import * as Alert from '$lib/components/ui/alert';
	import { onDestroy, onMount, tick, untrack } from 'svelte';
	import { prefersReducedMotion } from 'svelte/motion';
	import type { Todo } from './todo';
	import RankingCard from './RankingCard.svelte';
	import TodoDescription from './TodoDescription.svelte';
	import {
		chooseRanking,
		comparisonIndex,
		initialRankingState,
		rankingOrders,
		type RankingState,
	} from './ranking-state';
	import {
		layoutRankingScene,
		rankingCards,
		rankingEntrance,
		rankingTransition,
		type RankingClip,
		type Viewport,
	} from './ranking-scene';
	import { createRankingTimeline, type RankingCardElements } from './ranking-timeline';
	import { rankingMotion } from './ranking-motion';

	let {
		todo,
		siblings,
		snapshot,
		message,
		restartRequired = false,
	}: {
		todo: Todo;
		siblings: Todo[];
		snapshot: string;
		message?: string;
		restartRequired?: boolean;
	} = $props();

	const orders = untrack(() => rankingOrders(siblings));
	const cards = untrack(() => rankingCards(todo, orders));
	const initial = untrack(() => initialRankingState(siblings.length));
	const elements = new Map<string, RankingCardElements>();
	const timeline = createRankingTimeline(elements, () => experience);
	let ranking = $state<RankingState>(initial);
	let renderState = $state<RankingState>(initial);
	let viewport = $state<Viewport>({ width: 0, height: 0 });
	let experience = $state<HTMLDivElement>();
	let phase = $state<'entering' | 'comparing' | 'moving' | 'changing' | 'undoing' | 'review'>(
		'entering',
	);
	let saving = $state(false);
	let reloading = $state(false);
	let reviewCard = $state<HTMLElement>();
	let reviewForm = $state<HTMLFormElement>();
	let saveDirection: { x: number; y: number } | null = null;
	let generation = 0;
	type Checkpoint = { before: RankingState; after: RankingState; clip: RankingClip };
	let history = $state<Checkpoint[]>([]);
	type RankingAction = { kind: 'choose'; side: 'left' | 'right' } | { kind: 'undo' };
	let actions = $state<RankingAction[]>([]);
	let processing = $state(false);
	let queueGeneration = 0;
	$effect(() => {
		timeline.setSpeed(1 + actions.length * 0.5);
	});
	let busy = $derived(phase !== 'comparing' && phase !== 'review');
	let middle = $derived(
		ranking.stage === 'review' ? null : comparisonIndex(ranking, ranking.stage),
	);
	let comparison = $derived(
		ranking.stage !== 'review' && middle !== null ? orders[ranking.stage][middle] : null,
	);
	let focusedWidth = $derived(Math.max(120, Math.min(340, viewport.width - 116)));
	let focusedHeight = $derived(
		Math.max(140, Math.min(360, viewport.height * 0.36, (viewport.height - 56 - 104) / 2)),
	);
	let swipeActive = $state(false);
	let swipeX = $state(0);
	let swipeY = $state(0);
	let swipeThreshold = $derived(Math.min(100, Math.max(56, viewport.width * 0.16)));
	let swipeDistance = $derived(
		ranking.stage === 'review' ? Math.hypot(swipeX, swipeY) : Math.abs(swipeX),
	);
	let swipeProgress = $derived(Math.min(1, swipeDistance / swipeThreshold));
	let swipeSide = $derived(swipeActive && swipeX !== 0 ? (swipeX < 0 ? 'right' : 'left') : null);
	let swipeLabel = $derived(
		ranking.stage === 'review'
			? 'Save and close'
			: swipeSide === 'left'
				? ranking.stage === 'value'
					? 'Less value'
					: 'Less time'
				: ranking.stage === 'value'
					? 'More value'
					: 'More time',
	);
	type Swipe = { id: number; x: number; y: number; claimed: boolean };
	let swipe: Swipe | null = null;
	let swipeClick: { id: number; until: number } | null = null;

	function resetSwipe() {
		const previous = swipe;
		swipe = null;
		swipeActive = false;
		swipeX = 0;
		swipeY = 0;
		if (previous && experience?.hasPointerCapture(previous.id))
			experience.releasePointerCapture(previous.id);
	}

	function startSwipe(event: PointerEvent) {
		if (!event.isPrimary || event.button !== 0 || saving || reloading) return;
		if (
			!(event.target instanceof Element) ||
			event.target.closest(
				'a, input, textarea, select, [contenteditable="true"], button:not([data-swipe-choice])',
			)
		)
			return;
		swipe = { id: event.pointerId, x: event.clientX, y: event.clientY, claimed: false };
	}

	function moveSwipe(event: PointerEvent) {
		if (!swipe || event.pointerId !== swipe.id) return;
		const x = event.clientX - swipe.x;
		const y = event.clientY - swipe.y;
		if (!swipe.claimed) {
			if (ranking.stage === 'review') {
				if (Math.hypot(x, y) < 8) return;
			} else {
				if (Math.abs(y) > 8 && Math.abs(y) > Math.abs(x)) {
					resetSwipe();
					return;
				}
				if (Math.abs(x) < 8 || Math.abs(x) < Math.abs(y) * 1.2) return;
			}
			swipe.claimed = true;
			swipeActive = true;
			experience?.setPointerCapture(event.pointerId);
		}
		event.preventDefault();
		swipeX = x;
		swipeY = y;
	}

	function finishSwipe(event: PointerEvent) {
		if (!swipe || event.pointerId !== swipe.id) return;
		const claimed = swipe.claimed;
		const distance = event.clientX - swipe.x;
		const vertical = event.clientY - swipe.y;
		const magnitude =
			ranking.stage === 'review' ? Math.hypot(distance, vertical) : Math.abs(distance);
		if (claimed) swipeClick = { id: swipe.id, until: performance.now() + 250 };
		resetSwipe();
		if (claimed && magnitude >= swipeThreshold) {
			if (ranking.stage === 'review') {
				if (phase !== 'review' || restartRequired || processing || actions.length > 0) return;
				saveDirection = { x: distance / magnitude, y: vertical / magnitude };
				reviewForm?.requestSubmit();
			} else enqueue({ kind: 'choose', side: distance < 0 ? 'right' : 'left' });
		}
	}

	function cancelSwipe(event: PointerEvent) {
		// Touch capture transfers from the touched child to the experience when a swipe starts.
		if (event.type === 'lostpointercapture' && event.target !== experience) return;
		if (swipe?.id === event.pointerId) resetSwipe();
	}

	function suppressSwipeClick(event: MouseEvent) {
		if (!swipeClick || event.detail === 0) return;
		if (performance.now() > swipeClick.until) {
			swipeClick = null;
			return;
		}
		if (event instanceof PointerEvent && event.pointerId !== swipeClick.id) return;
		swipeClick = null;
		event.preventDefault();
		event.stopPropagation();
	}

	function register(key: string, nodes: RankingCardElements) {
		elements.set(key, nodes);
		return () => {
			elements.delete(key);
		};
	}

	function idle() {
		phase = ranking.stage === 'review' ? 'review' : 'comparing';
		void drainQueue();
	}

	function enqueue(action: RankingAction) {
		if (saving || reloading) return;
		actions.push(action);
		void drainQueue();
	}

	async function drainQueue() {
		if (processing || busy || saving || reloading) return;
		processing = true;
		const current = queueGeneration;
		try {
			while (current === queueGeneration && actions.length > 0 && !saving && !reloading) {
				const action = actions.shift()!;
				if (action.kind === 'undo') await undo();
				else if (ranking.stage !== 'review') {
					await choose(
						action.side === 'left' ? ranking.stage === 'time' : ranking.stage === 'value',
					);
				}
			}
		} finally {
			if (current === queueGeneration) processing = false;
		}
	}

	function clearQueue() {
		queueGeneration++;
		actions = [];
		processing = false;
	}

	async function enter() {
		const current = ++generation;
		phase = 'entering';
		await tick();
		if (current !== generation) return;
		if (ranking.stage === 'review') {
			timeline.set(layoutRankingScene(ranking, cards, orders, viewport));
			idle();
			return;
		}
		const complete = await timeline.play(
			rankingEntrance(ranking, cards, orders, viewport),
			false,
			prefersReducedMotion.current,
		);
		if (current === generation && complete) idle();
	}

	async function choose(before: boolean) {
		if (busy || saving || reloading || ranking.stage === 'review') return;
		const previous = ranking;
		const next = chooseRanking(previous, orders, before);
		const clip = rankingTransition(previous, next, cards, orders, viewport);
		const current = ++generation;
		phase = previous.stage === next.stage ? 'moving' : 'changing';
		renderState = next;
		const complete = await timeline.play(clip, false, prefersReducedMotion.current);
		if (current !== generation || !complete) return;
		history.push({ before: previous, after: next, clip });
		ranking = next;
		idle();
	}

	async function undo() {
		if (busy || saving || reloading) return;
		const previous = history.at(-1);
		if (!previous) return;
		const current = ++generation;
		phase = 'undoing';
		renderState = previous.before;
		const complete = await timeline.play(previous.clip, true, prefersReducedMotion.current);
		if (current !== generation || !complete) return;
		history.pop();
		ranking = previous.before;
		idle();
	}

	function restart() {
		resetSwipe();
		clearQueue();
		generation++;
		timeline.cancel();
		ranking = initialRankingState(orders.value.length);
		renderState = ranking;
		history = [];
		void enter();
	}

	function resize() {
		if (!experience) return;
		const next = {
			width: Math.round(experience.clientWidth),
			height: Math.round(experience.clientHeight),
		};
		if (next.width === viewport.width && next.height === viewport.height) return;
		resetSwipe();
		clearQueue();
		generation++;
		timeline.cancel();
		viewport = next;
		renderState = ranking;
		history = history.map((checkpoint) => ({
			...checkpoint,
			clip: rankingTransition(checkpoint.before, checkpoint.after, cards, orders, next),
		}));
		timeline.set(layoutRankingScene(ranking, cards, orders, viewport));
		idle();
	}

	function handleArrow(event: KeyboardEvent) {
		if (
			(event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') ||
			event.defaultPrevented ||
			event.altKey ||
			event.ctrlKey ||
			event.metaKey ||
			event.shiftKey ||
			saving ||
			reloading ||
			ranking.stage === 'review' ||
			(event.target instanceof Element &&
				event.target.closest(
					'input, textarea, select, [contenteditable]:not([contenteditable="false"])',
				))
		)
			return;
		event.preventDefault();
		enqueue({ kind: 'choose', side: event.key === 'ArrowLeft' ? 'left' : 'right' });
	}

	function fitReviewText(node: HTMLElement) {
		const observer = new ResizeObserver(() => {
			const title = node.querySelector('h1');
			const available = node.clientHeight - (title?.getBoundingClientRect().height ?? 0) - 10;
			node.style.setProperty(
				'--description-lines',
				String(Math.max(1, Math.floor(available / 24))),
			);
		});
		observer.observe(node);
		return { destroy: () => observer.disconnect() };
	}

	onMount(() => {
		if (!experience) return;
		const events = new AbortController();
		window.addEventListener('keydown', handleArrow, { signal: events.signal });
		experience.addEventListener('pointerdown', startSwipe, { signal: events.signal });
		experience.addEventListener('pointermove', moveSwipe, {
			signal: events.signal,
			passive: false,
		});
		experience.addEventListener('pointerup', finishSwipe, { signal: events.signal });
		experience.addEventListener('pointercancel', cancelSwipe, { signal: events.signal });
		experience.addEventListener('lostpointercapture', cancelSwipe, { signal: events.signal });
		experience.addEventListener('click', suppressSwipeClick, {
			signal: events.signal,
			capture: true,
		});
		viewport = {
			width: Math.round(experience.clientWidth),
			height: Math.round(experience.clientHeight),
		};
		void enter();
		const observer = new ResizeObserver(resize);
		observer.observe(experience);
		return () => {
			observer.disconnect();
			events.abort();
			resetSwipe();
		};
	});
	onDestroy(() => {
		clearQueue();
		generation++;
		timeline.cancel();
	});
</script>

<div
	bind:this={experience}
	class="experience"
	class:swiping={swipeActive}
	class:reviewing={phase === 'review'}
	data-category={renderState.stage === 'value' ? 'value' : 'time'}
	style:--focused-width={`${focusedWidth}px`}
	style:--focused-height={`${focusedHeight}px`}
	aria-busy={busy}
>
	<div
		class="absolute inset-x-0 top-0 z-[200] flex items-center justify-between gap-3 px-4 py-3 sm:px-6"
	>
		<Button variant="outline" href={todo.parentId === null ? '/' : `/${todo.parentId}`}
			>← Back to list</Button
		>
		<Button variant="ghost" disabled={saving || reloading} onclick={restart}>Start over</Button>
	</div>
	<div class="absolute top-3 left-1/2 z-[201] flex -translate-x-1/2 flex-col items-center">
		<Button
			variant="ghost"
			size="sm"
			disabled={(history.length === 0 && !processing && actions.length === 0) ||
				saving ||
				reloading}
			onclick={() => enqueue({ kind: 'undo' })}>Undo</Button
		>
	</div>

	{#each cards as card (card.key)}
		{@const bounds = renderState[card.category]}
		{@const active = ranking.stage === card.category && (card.selected || card.index === middle)}
		<RankingCard
			{card}
			{busy}
			{active}
			{register}
			swipeOffset={active && !card.selected ? Math.sign(swipeX) * swipeProgress * 32 : 0}
			backLabel={card.selected
				? 'Your task'
				: card.category === 'value'
					? card.index < bounds.low
						? 'More valuable'
						: 'Less valuable'
					: card.index < bounds.low
						? 'Less time'
						: 'More time'}
		/>
	{/each}

	{#if ranking.stage !== 'review' && viewport.width > 0}
		<div
			class="choice-zone choice-left"
			class:locked={saving || reloading}
			style:--swipe-glow={swipeSide === 'left' ? swipeProgress : 0}
		>
			<Button
				variant="ghost"
				data-swipe-choice
				style={swipeSide === 'left' ? 'color: var(--ranking-choice-color)' : undefined}
				class="absolute inset-0 h-auto w-full flex-col gap-3 rounded-none px-1 text-muted-foreground hover:bg-transparent hover:text-[var(--ranking-choice-color)] focus-visible:text-[var(--ranking-choice-color)] dark:hover:bg-transparent"
				disabled={saving || reloading}
				onclick={() => enqueue({ kind: 'choose', side: 'left' })}
				aria-label={ranking.stage === 'value'
					? 'This task has less value'
					: 'This task takes less time'}
			>
				<span aria-hidden="true" class="text-3xl">←</span><span class="text-xs sm:text-base"
					>{ranking.stage === 'value' ? 'Less value' : 'Less time'}</span
				>
			</Button>
		</div>
		<div
			class="choice-zone choice-right"
			class:locked={saving || reloading}
			style:--swipe-glow={swipeSide === 'right' ? swipeProgress : 0}
		>
			<Button
				variant="ghost"
				data-swipe-choice
				style={swipeSide === 'right' ? 'color: var(--ranking-choice-color)' : undefined}
				class="absolute inset-0 h-auto w-full flex-col gap-3 rounded-none px-1 text-muted-foreground hover:bg-transparent hover:text-[var(--ranking-choice-color)] focus-visible:text-[var(--ranking-choice-color)] dark:hover:bg-transparent"
				disabled={saving || reloading}
				onclick={() => enqueue({ kind: 'choose', side: 'right' })}
				aria-label={ranking.stage === 'value'
					? 'This task has more value'
					: 'This task takes more time'}
			>
				<span aria-hidden="true" class="text-3xl">→</span><span class="text-xs sm:text-base"
					>{ranking.stage === 'value' ? 'More value' : 'More time'}</span
				>
			</Button>
		</div>
	{/if}

	{#if phase === 'review'}
		<section
			bind:this={reviewCard}
			aria-label="Review rankings"
			class="review-card"
			class:dragging={swipeActive}
			style:translate={`${swipeDistance ? (swipeX / swipeDistance) * swipeProgress * 32 : 0}px ${swipeDistance ? (swipeY / swipeDistance) * swipeProgress * 32 : 0}px`}
		>
			<div class="review-copy min-h-0 flex-1 overflow-hidden" use:fitReviewText>
				<h1 class="break-words text-xl font-semibold">{todo.name}</h1>
				<div class="review-description"><TodoDescription descrption={todo.description} /></div>
			</div>
			<dl class="grid grid-cols-2 gap-3">
				<div
					class="rounded-lg border border-rose-200 bg-rose-50 p-2 text-rose-700 dark:border-rose-900 dark:bg-rose-950 dark:text-rose-300"
				>
					<dt class="text-xs font-medium">Value</dt>
					<dd class="text-xl font-bold">#{ranking.valuePosition}</dd>
				</div>
				<div
					class="rounded-lg border border-blue-200 bg-blue-50 p-2 text-blue-700 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-300"
				>
					<dt class="text-xs font-medium">Time</dt>
					<dd class="text-xl font-bold">#{ranking.timePosition}</dd>
				</div>
			</dl>
			<form
				bind:this={reviewForm}
				method="POST"
				action="?/save"
				use:enhance={({ cancel }) => {
					if (saving || restartRequired || processing || actions.length > 0) {
						cancel();
						return;
					}
					const angle = Math.random() * Math.PI * 2;
					const direction = saveDirection ?? { x: Math.cos(angle), y: Math.sin(angle) };
					saveDirection = null;
					saving = true;
					return async ({ result, update }) => {
						try {
							if (
								(result.type === 'redirect' || result.type === 'success') &&
								reviewCard &&
								!prefersReducedMotion.current
							) {
								const distance =
									Math.hypot(viewport.width, viewport.height) +
									Math.hypot(reviewCard.offsetWidth, reviewCard.offsetHeight);
								await reviewCard
									.animate(
										[
											{ transform: 'translateX(0) rotate(0deg)', opacity: 1 },
											{
												transform: `translate(${distance * direction.x}px, ${distance * direction.y}px) rotate(${direction.x * 12}deg)`,
												opacity: 0,
											},
										],
										{
											duration: rankingMotion.duration,
											easing: rankingMotion.easing,
											fill: 'forwards',
										},
									)
									.finished.catch(() => {});
							}
							await update();
						} finally {
							saving = false;
						}
					};
				}}
			>
				<input type="hidden" name="valueBeforeId" value={ranking.valueBeforeId ?? ''} />
				<input type="hidden" name="timeBeforeId" value={ranking.timeBeforeId ?? ''} />
				<input type="hidden" name="snapshot" value={snapshot} />
				<Button
					type="submit"
					class="w-full"
					size="lg"
					disabled={saving || restartRequired || processing || actions.length > 0}
					>{saving ? 'Saving…' : 'Save rankings'}</Button
				>
			</form>
		</section>
	{/if}

	{#if message}
		<div class="absolute inset-x-4 bottom-4 z-[210] mx-auto max-w-lg space-y-2">
			<Alert.Root variant="destructive"><Alert.Description>{message}</Alert.Description></Alert.Root
			>
			{#if restartRequired}<Button
					variant="outline"
					disabled={reloading || busy}
					onclick={async () => {
						reloading = true;
						try {
							await applyAction({ type: 'success', status: 200, data: {} });
							await invalidateAll();
							restart();
						} finally {
							reloading = false;
						}
					}}>Reload current rankings</Button
				>{/if}
		</div>
	{/if}
	<p class="sr-only" aria-live="polite" aria-atomic="true">
		{swipeSide || (swipeActive && ranking.stage === 'review')
			? `${swipeLabel}. ${swipeProgress >= 1 ? 'Release to choose.' : 'Keep swiping.'}`
			: busy
				? 'Moving cards.'
				: comparison
					? `Compare ${todo.name} with ${comparison.name} for ${ranking.stage}.`
					: 'Review and save your rankings.'}
	</p>
</div>

<style>
	.review-card {
		position: absolute;
		top: calc((100% - var(--focused-height)) / 2);
		left: 0;
		right: 0;
		z-index: 160;
		margin-inline: auto;
		width: var(--focused-width);
		height: var(--focused-height);
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 22px;
		border: 1px solid #7e22ce;
		border-radius: 12px;
		background: #faf5ff;
		box-shadow: 0 20px 48px -14px #00000030;
		transition: translate 180ms ease-out;
	}
	.review-card h1 {
		color: #7e22ce;
		margin-bottom: 10px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.review-description {
		display: -webkit-box;
		-webkit-box-orient: vertical;
		-webkit-line-clamp: var(--description-lines, 2);
		line-clamp: var(--description-lines, 2);
		line-height: 24px;
		overflow: hidden;
	}
	.review-card dl,
	.review-card form {
		flex-shrink: 0;
	}
	.review-card.dragging {
		transition: none;
	}
	:global(.dark) .review-card {
		background: #2e1f40;
		border-color: #d8b4fe;
	}
	:global(.dark) .review-card h1 {
		color: #d8b4fe;
	}
	@media (prefers-reduced-motion: reduce) {
		.review-card {
			transition: none;
		}
	}
	@property --ranking-choice-color {
		syntax: '<color>';
		inherits: true;
		initial-value: #be123c;
	}
	.experience {
		touch-action: pan-y pinch-zoom;
		--ranking-value-color: #be123c;
		--ranking-time-color: #1d4ed8;
		--ranking-choice-color: var(--ranking-value-color);
		position: fixed;
		inset: 0;
		z-index: 40;
		height: 100dvh;
		overflow: hidden;
		background: var(--background);
		isolation: isolate;
	}
	.experience.reviewing {
		touch-action: none;
	}
	.experience[data-category='time'] {
		--ranking-choice-color: var(--ranking-time-color);
	}
	:global(.dark) .experience {
		--ranking-value-color: #fda4af;
		--ranking-time-color: #93c5fd;
	}
	.choice-zone {
		--glow-origin: 0%;
		position: absolute;
		top: 56px;
		bottom: 0;
		z-index: 150;
		width: calc((100% - var(--focused-width)) / 2);
		min-width: 58px;
	}
	.choice-zone.locked {
		pointer-events: none;
	}
	.choice-zone::before {
		content: '';
		position: absolute;
		top: -56px;
		bottom: 0;
		left: 0;
		width: 100vw;
		pointer-events: none;
		background: radial-gradient(
			ellipse at var(--glow-origin) 50%,
			color-mix(in srgb, var(--ranking-choice-color) 22%, transparent) 0%,
			color-mix(in srgb, var(--ranking-choice-color) 9%, transparent) 38%,
			transparent 72%
		);
		opacity: 0;
		transition: opacity 220ms ease;
	}
	.choice-zone:has(:focus-visible)::before {
		opacity: 1;
	}
	.experience.swiping {
		user-select: none;
	}
	.experience.swiping .choice-zone::before {
		opacity: var(--swipe-glow, 0);
		transition: none;
	}
	@media (hover: hover) {
		.choice-zone:hover::before {
			opacity: 1;
		}
	}
	.choice-left {
		left: 0;
	}
	.choice-right {
		--glow-origin: 100%;
		right: 0;
	}
	.choice-right::before {
		left: auto;
		right: 0;
	}
	@media (prefers-reduced-motion: reduce) {
		.choice-zone::before {
			transition: none;
		}
	}
</style>
