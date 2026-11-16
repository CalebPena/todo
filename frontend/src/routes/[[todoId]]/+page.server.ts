import { actionWrapper } from '$lib/api/helpers';
import { TOKEN_COOKIE_NAME } from '$lib/constants';
import { validateApiInput } from '$lib/schemas';
import { getTodo, listTodos, saveTodo, setCompleted } from '$lib/todo/api.server';
import { todoInputSchema } from '$lib/todo/todo';
import { error, isHttpError, redirect, type Cookies } from '@sveltejs/kit';
import { z } from 'zod';
import type { Actions, PageServerLoad } from './$types';

const idSchema = z.coerce.number().int().positive().max(2147483647);

function requireToken(cookies: Cookies, url: URL) {
	const token = cookies.get(TOKEN_COOKIE_NAME);
	if (!token) {
		redirect(303, `/login?redirect=${encodeURIComponent(url.pathname)}`);
	}
	return token;
}

function routeId(value: string | undefined) {
	if (value === undefined) {
		return null;
	}
	const result = idSchema.safeParse(value);
	if (!result.success) {
		error(404, 'Todo not found');
	}
	return result.data;
}

export const load: PageServerLoad = async ({ cookies, params, url }) => {
	const token = requireToken(cookies, url);
	const id = routeId(params.todoId);
	try {
		if (id === null) {
			const todos = await listTodos(token);
			return { todo: null, todos: todos.filter((todo) => todo.parentId === null) };
		}
		const todo = await getTodo(id, token);
		return { todo, todos: todo.children };
	} catch (err) {
		if (isHttpError(err) && (err.status === 401 || err.status === 403)) {
			cookies.delete(TOKEN_COOKIE_NAME, { path: '/' });
			redirect(303, `/login?redirect=${encodeURIComponent(url.pathname)}`);
		}
		throw err;
	}
};

export const actions: Actions = {
	create: actionWrapper(async ({ request, cookies, params, url }) => {
		const token = requireToken(cookies, url);
		const form = await request.formData();
		const input = validateApiInput(todoInputSchema, {
			name: form.get('name'),
			description: form.get('description'),
		});
		await saveTodo(null, { ...input, parentId: routeId(params.todoId) }, token);
		return { success: true, action: 'create' };
	}),
	edit: actionWrapper(async ({ request, cookies, params, url }) => {
		const token = requireToken(cookies, url);
		const id = routeId(params.todoId);
		if (id === null) {
			error(400, 'Choose a todo to edit');
		}
		const form = await request.formData();
		const input = validateApiInput(todoInputSchema, {
			name: form.get('name'),
			description: form.get('description'),
		});
		await saveTodo(id, input, token);
		return { success: true, action: 'edit' };
	}),
	complete: actionWrapper(async ({ request, cookies, url }) => {
		const token = requireToken(cookies, url);
		const form = await request.formData();
		const id = validateApiInput(idSchema, form.get('id'));
		const completed = validateApiInput(z.enum(['true', 'false']), form.get('completed'));
		await setCompleted(id, completed === 'true', token);
		return { success: true, action: 'complete' };
	}),
};
