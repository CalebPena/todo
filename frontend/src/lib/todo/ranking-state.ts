import type { Todo } from './todo';

export type RankingCategory = 'value' | 'time';
export type RankingBounds = { low: number; high: number };
export type RankingState = {
	stage: RankingCategory | 'review';
	value: RankingBounds;
	time: RankingBounds;
	valueBeforeId: number | null;
	timeBeforeId: number | null;
	valuePosition: number;
	timePosition: number;
};
export type RankingOrders = Record<RankingCategory, Todo[]>;

export function rankingOrders(siblings: Todo[]): RankingOrders {
	const sort = (category: RankingCategory) =>
		siblings.toSorted(
			(a, b) =>
				(a[category === 'value' ? 'valueRank' : 'timeRank'] ?? Infinity) -
					(b[category === 'value' ? 'valueRank' : 'timeRank'] ?? Infinity) || a.id - b.id,
		);
	return { value: sort('value'), time: sort('time') };
}

export function initialRankingState(count: number): RankingState {
	return {
		stage: count === 0 ? 'review' : 'value',
		value: { low: 0, high: count },
		time: { low: 0, high: count },
		valueBeforeId: null,
		timeBeforeId: null,
		valuePosition: 1,
		timePosition: 1,
	};
}

export function comparisonIndex(state: RankingState, category: RankingCategory) {
	const { low, high } = state[category];
	return low < high ? Math.floor((low + high) / 2) : null;
}

export function chooseRanking(
	state: RankingState,
	orders: RankingOrders,
	before: boolean,
): RankingState {
	if (state.stage === 'review') return state;
	const category = state.stage;
	const middle = comparisonIndex(state, category);
	if (middle === null) return state;
	const bounds = before
		? { low: state[category].low, high: middle }
		: { low: middle + 1, high: state[category].high };
	const next = { ...state, [category]: bounds };
	if (bounds.low !== bounds.high) return next;
	const beforeId = orders[category][bounds.low]?.id ?? null;
	return category === 'value'
		? { ...next, valueBeforeId: beforeId, valuePosition: bounds.low + 1, stage: 'time' }
		: { ...next, timeBeforeId: beforeId, timePosition: bounds.low + 1, stage: 'review' };
}
