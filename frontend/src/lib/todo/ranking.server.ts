import { error } from '@sveltejs/kit';
import { getTodo, listTodos } from './api.server';

export async function getRankingContext(id: number, token: string) {
	const [todo, todos] = await Promise.all([getTodo(id, token), listTodos(token)]);
	if (todo.completed) {
		error(400, 'Reopen this todo before ranking it');
	}
	const siblings = todos.filter(
		(sibling) =>
			sibling.id !== id &&
			sibling.parentId === todo.parentId &&
			!sibling.completed &&
			sibling.ranked,
	);
	const snapshot = JSON.stringify({
		id: todo.id,
		parentId: todo.parentId,
		valueRank: todo.valueRank,
		timeRank: todo.timeRank,
		siblings: siblings
			.toSorted((a, b) => a.id - b.id)
			.map(({ id, valueRank, timeRank }) => ({ id, valueRank, timeRank })),
	});
	return { todo, siblings, snapshot };
}
