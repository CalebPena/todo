import { dev } from '$app/environment';
import { registerEndpoint } from '$lib/api/endpoints';
import { actionWrapper, requestApi } from '$lib/api/helpers';
import { HOME_PAGE_PATH, TOKEN_COOKIE_NAME } from '$lib/constants';
import { emailSchema, passwordSchema, responseSchemaWrapper, validateApiInput } from '$lib/schemas';
import { redirect, type Actions } from '@sveltejs/kit';
import { z } from 'zod';

export const actions: Actions = {
	default: actionWrapper(async ({ request, cookies }) => {
		const formData = await request.formData();

		const email = validateApiInput(emailSchema, formData.get('email'));
		const password = validateApiInput(passwordSchema, formData.get('password'));

		const { data } = await requestApi(
			registerEndpoint(),
			responseSchemaWrapper(z.object({ token: z.string() })),
			'POST',
			{ email: email, password: password },
		);

		cookies.set(TOKEN_COOKIE_NAME, data.response.token, {
			httpOnly: true,
			secure: !dev,
			sameSite: 'strict',
			path: '/',
		});

		throw redirect(303, HOME_PAGE_PATH);
	}),
};
