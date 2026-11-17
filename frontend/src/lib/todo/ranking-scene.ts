import type { Todo } from './todo';
import {
	comparisonIndex,
	type RankingCategory,
	type RankingOrders,
	type RankingState,
} from './ranking-state';
import { rankingEase, rankingMotion } from './ranking-motion';

export type Viewport = { width: number; height: number };
export type RankingCardModel = {
	key: string;
	category: RankingCategory;
	todo: Todo;
	index: number;
	selected: boolean;
};
export type CardPose = {
	x: number;
	y: number;
	width: number;
	height: number;
	rotation: number;
	flip: number;
	fontSize: number;
	opacity: number;
	z: number;
};
export type RankingScene = Map<string, CardPose>;
export type RankingClip = {
	duration: number;
	accent: { from: RankingCategory; to: RankingCategory };
	start: RankingScene;
	end: RankingScene;
	tracks: {
		key: string;
		frames: { offset: number; pose: CardPose }[];
		reverseFrames: { offset: number; pose: CardPose }[];
	}[];
};

const clamp = (min: number, value: number, max: number) => Math.max(min, Math.min(value, max));

export function rankingCards(todo: Todo, orders: RankingOrders): RankingCardModel[] {
	return (['value', 'time'] as const).flatMap((category) => [
		{ key: `${category}:${todo.id}`, category, todo, index: -1, selected: true },
		...orders[category].map((sibling, index) => ({
			key: `${category}:${sibling.id}`,
			category,
			todo: sibling,
			index,
			selected: false,
		})),
	]);
}

function geometry(viewport: Viewport) {
	const width = Math.max(120, Math.min(340, viewport.width - 116));
	const height = Math.max(
		140,
		Math.min(360, viewport.height * 0.36, (viewport.height - 56 - 104) / 2),
	);
	return {
		width,
		height,
		topX: (viewport.width - width) / 2,
		topY: 56,
		focusY: Math.max(
			56 + height + 16,
			viewport.height - clamp(80, viewport.height * 0.16, 160) - height,
		),
		spacing: () => 130,
	};
}

function restingPose(
	card: RankingCardModel,
	position: number,
	focus: number,
	viewport: Viewport,
): CardPose {
	const offset = position - focus;
	const direction = card.category === 'value' ? -1 : 1;
	return {
		x: viewport.width / 2 - 88 + offset * geometry(viewport).spacing() * direction,
		y: viewport.height - 90,
		width: 176,
		height: 190,
		rotation: clamp(-7, offset * 1.8 * direction, 7),
		flip: 0,
		fontSize: 14,
		opacity: 1,
		z: Math.max(1, 40 - Math.round(Math.abs(offset))),
	};
}

function offscreen(pose: CardPose, side: 'left' | 'right', viewport: Viewport): CardPose {
	const distance = viewport.width + Math.abs(pose.x) + pose.width + 240;
	return { ...pose, x: pose.x + (side === 'right' ? distance : -distance), opacity: 0 };
}

export function layoutRankingScene(
	state: RankingState,
	cards: RankingCardModel[],
	orders: RankingOrders,
	viewport: Viewport,
): RankingScene {
	const size = geometry(viewport);
	return new Map(
		cards.map((card) => {
			const bounds = state[card.category];
			const middle = comparisonIndex(state, card.category);
			const placed = middle === null;
			const focus = middle ?? bounds.low;
			const position = card.selected
				? bounds.low
				: card.index + (placed && card.index >= bounds.low ? 1 : 0);
			let pose = restingPose(card, position, focus, viewport);
			if (!card.selected)
				pose.flip = card.index < bounds.low || card.index >= bounds.high ? 180 : 0;
			if (card.selected && !placed) {
				pose = {
					...pose,
					x: size.topX,
					y: size.topY,
					width: size.width,
					height: size.height,
					rotation: 0,
					fontSize: 20,
					z: 120,
				};
			} else if (!card.selected && card.index === middle) {
				pose = {
					...pose,
					x: size.topX,
					y: size.focusY,
					width: size.width,
					height: size.height,
					rotation: 0,
					fontSize: 20,
					z: 100,
				};
			}
			if (state.stage !== card.category) {
				const completed = card.category === 'value' || state.stage === 'review';
				pose = offscreen(pose, completed ? 'left' : 'right', viewport);
			}
			return [card.key, pose];
		}),
	);
}

function interpolate(from: CardPose, to: CardPose, progress: number): CardPose {
	const result = { ...from };
	for (const key of Object.keys(from) as (keyof CardPose)[])
		result[key] = from[key] + (to[key] - from[key]) * progress;
	return result;
}

type MotionPath = { key: string; from: CardPose; to: CardPose; via?: CardPose; delay: number };

function clip(
	paths: MotionPath[],
	start: RankingScene,
	end: RankingScene,
	accent: RankingClip['accent'],
): RankingClip {
	const lastDelay = Math.max(0, ...paths.map((path) => path.delay));
	const duration = rankingMotion.duration + lastDelay;
	const samples = Math.ceil(duration / (1000 / 60));
	function sample(path: MotionPath, reverse: boolean) {
		const delay = reverse ? lastDelay - path.delay : path.delay;
		return Array.from({ length: samples + 1 }, (_, index) => {
			const offset = index / samples;
			const eased = rankingEase(clamp(0, (offset * duration - delay) / rankingMotion.duration, 1));
			const progress = reverse ? 1 - eased : eased;
			const pose = path.via
				? progress < 0.8
					? interpolate(path.from, path.via, progress / 0.8)
					: interpolate(path.via, path.to, (progress - 0.8) / 0.2)
				: interpolate(path.from, path.to, progress);
			return { offset, pose: { ...pose, z: Math.max(path.from.z, path.to.z) } };
		});
	}
	return {
		duration,
		accent,
		start,
		end,
		tracks: paths
			.filter((path) => JSON.stringify(path.from) !== JSON.stringify(path.to))
			.map((path) => ({
				key: path.key,
				frames: sample(path, false),
				reverseFrames: sample(path, true),
			})),
	};
}

function entryDelay(card: RankingCardModel, count: number) {
	if (card.selected) return 0;
	return Math.min(card.category === 'value' ? count - 1 - card.index : card.index, 12) * 35;
}

export function rankingEntrance(
	state: RankingState,
	cards: RankingCardModel[],
	orders: RankingOrders,
	viewport: Viewport,
): RankingClip {
	const end = layoutRankingScene(state, cards, orders, viewport);
	const start = new Map(end);
	const paths = cards
		.filter((card) => card.category === state.stage)
		.map((card) => {
			const to = end.get(card.key)!;
			const from = { ...offscreen(to, 'right', viewport), opacity: 1 };
			start.set(card.key, from);
			return { key: card.key, from, to, delay: entryDelay(card, orders[card.category].length) };
		});
	const category = state.stage === 'value' ? 'value' : 'time';
	return clip(paths, start, end, { from: category, to: category });
}

export function rankingTransition(
	before: RankingState,
	after: RankingState,
	cards: RankingCardModel[],
	orders: RankingOrders,
	viewport: Viewport,
): RankingClip {
	const start = layoutRankingScene(before, cards, orders, viewport);
	const end = layoutRankingScene(after, cards, orders, viewport);
	const crossing = before.stage !== after.stage;
	const paths = cards.map((card): MotionPath => {
		const from = start.get(card.key)!;
		const to = end.get(card.key)!;
		if (crossing && card.category === before.stage) {
			const slot = after[card.category].low;
			const position = card.selected ? slot : card.index + Number(card.index >= slot);
			const total = orders[card.category].length + 1;
			const resting = restingPose(card, position, slot, viewport);
			const spread = Math.max(slot, total - 1 - slot) * geometry(viewport).spacing();
			const travel = viewport.width * 2 + spread + 520;
			const spatial = card.category === 'value' ? total - 1 - position : position;
			const flip = card.selected ? 0 : 180;
			const via = { ...resting, x: resting.x - travel * 0.4, flip };
			const outgoing = { ...resting, x: resting.x - travel, flip, opacity: 0 };
			end.set(card.key, outgoing);
			return { key: card.key, from, via, to: outgoing, delay: Math.min(spatial, 12) * 35 };
		}
		if (crossing && card.category === after.stage) {
			const incoming = { ...offscreen(to, 'right', viewport), opacity: 1 };
			start.set(card.key, incoming);
			return {
				key: card.key,
				from: incoming,
				to,
				delay: entryDelay(card, orders[card.category].length),
			};
		}
		return { key: card.key, from, to, delay: 0 };
	});
	return clip(paths, start, end, {
		from: before.stage === 'value' ? 'value' : 'time',
		to: after.stage === 'value' ? 'value' : 'time',
	});
}
