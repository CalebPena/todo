import { actionWrapper } from '$lib/api/helpers';
import { TOKEN_COOKIE_NAME } from '$lib/constants';
import { validateApiInput } from '$lib/schemas';
import { rankTodo } from '$lib/todo/api.server';
import { getRankingContext } from '$lib/todo/ranking.server';
import { error, fail, isHttpError, redirect, type Cookies } from '@sveltejs/kit';
import { z } from 'zod';
import type { Actions, PageServerLoad } from './$types';

const idSchema = z.coerce.number().int().positive().max(2147483647);
const beforeIdSchema = z.number().int().positive().max(2147483647).nullable();

function beforeId(value: FormDataEntryValue | null) {
	return value === '' ? null : typeof value === 'string' ? Number(value) : NaN;
}

function requireToken(cookies: Cookies, url: URL) {
	const token = cookies.get(TOKEN_COOKIE_NAME);
	if (!token) redirect(303, `/login?redirect=${encodeURIComponent(url.pathname)}`);
	return token;
}

function routeId(value: string | undefined) {
	const result = idSchema.safeParse(value);
	if (!result.success) error(404, 'Todo not found');
	return result.data;
}

export const load: PageServerLoad = async ({ cookies, params, url }) => {
	const token = requireToken(cookies, url);
	try {
		return await getRankingContext(routeId(params.todoId), token);
	} catch (err) {
		if (isHttpError(err) && (err.status === 401 || err.status === 403)) {
			cookies.delete(TOKEN_COOKIE_NAME, { path: '/' });
			redirect(303, `/login?redirect=${encodeURIComponent(url.pathname)}`);
		}
		throw err;
	}
};

export const actions: Actions = {
	save: actionWrapper(async ({ cookies, params, request, url }) => {
		const token = requireToken(cookies, url);
		const id = routeId(params.todoId);
		const form = await request.formData();
		const input = validateApiInput(
			z.object({
				valueBeforeId: beforeIdSchema,
				timeBeforeId: beforeIdSchema,
				snapshot: z.string(),
			}),
			{
				valueBeforeId: beforeId(form.get('valueBeforeId')),
				timeBeforeId: beforeId(form.get('timeBeforeId')),
				snapshot: form.get('snapshot'),
			},
		);
		const current = await getRankingContext(id, token);
		if (input.snapshot !== current.snapshot) {
			return fail(409, {
				message: 'Sibling rankings changed while you were comparing. Reload and try again.',
				restartRequired: true,
			});
		}
		await rankTodo(
			id,
			{ valueBeforeId: input.valueBeforeId, timeBeforeId: input.timeBeforeId },
			token,
		);
		redirect(303, current.todo.parentId === null ? '/' : `/${current.todo.parentId}`);
	}),
};
