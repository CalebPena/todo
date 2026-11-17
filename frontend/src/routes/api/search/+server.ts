import { TOKEN_COOKIE_NAME } from '$lib/constants';
import { searchTodos, listFavoriteTodos } from '$lib/todo/api.server';
import { isHttpError, json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ cookies, url }) => {
	const headers = { 'Cache-Control': 'private, no-store' };
	const token = cookies.get(TOKEN_COOKIE_NAME);
	if (!token) return json({ message: 'Sign in to search your tasks' }, { status: 401, headers });
	const query = (url.searchParams.get('q') ?? '').trim();
	if (query.length > 200)
		return json({ message: 'Search query is too long' }, { status: 400, headers });
	try {
		return json(
			{ todos: query ? await searchTodos(query, token) : await listFavoriteTodos(token) },
			{ headers },
		);
	} catch (error) {
		if (isHttpError(error))
			return json({ message: error.body.message }, { status: error.status, headers });
		throw error;
	}
};
