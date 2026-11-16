import { BACKEND_DOMAIN } from '$env/static/private';
import { requestApi } from '$lib/api/helpers';
import { responseSchemaWrapper } from '$lib/schemas';
import { z } from 'zod';
import { todoSchema } from './todo';

const todoResponseSchema = responseSchemaWrapper(z.object({ todo: todoSchema }));

export async function listTodos(token: string) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/todos`,
		responseSchemaWrapper(z.object({ todos: z.array(todoSchema) })),
		'GET',
		undefined,
		token,
	);
	return data.response.todos;
}

export async function getTodo(id: number, token: string) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/todos/${id}`,
		todoResponseSchema,
		'GET',
		undefined,
		token,
	);
	return data.response.todo;
}

export async function saveTodo(
	id: number | null,
	input: { name: string; description: string; parentId?: number | null },
	token: string,
) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/todos${id === null ? '' : `/${id}`}`,
		todoResponseSchema,
		id === null ? 'POST' : 'PATCH',
		input,
		token,
	);
	return data.response.todo;
}

export async function setCompleted(id: number, completed: boolean, token: string) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/todos/${id}/complete`,
		todoResponseSchema,
		completed ? 'PUT' : 'DELETE',
		undefined,
		token,
	);
	return data.response.todo;
}
