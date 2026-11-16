import { dev } from '$app/environment';
import { loginEndpoint } from '$lib/api/endpoints';
import { actionWrapper, requestApi } from '$lib/api/helpers';
import { HOME_PAGE_PATH, REDIRECT_PARAM_NAME, TOKEN_COOKIE_NAME } from '$lib/constants';
import {
	emailSchema,
	loginPasswordSchema,
	responseSchemaWrapper,
	validateApiInput,
} from '$lib/schemas';
import { redirect, type Actions } from '@sveltejs/kit';
import { z } from 'zod';

export const actions: Actions = {
	login: actionWrapper(async ({ request, cookies, url }) => {
		const formData = await request.formData();

		const email = validateApiInput(emailSchema, formData.get('email'));
		const password = validateApiInput(loginPasswordSchema, formData.get('password'));

		const { data } = await requestApi(
			loginEndpoint(),
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

		const redirectUrl = url.searchParams.get(REDIRECT_PARAM_NAME);

		// Only follow same-site paths; "//host" and "/\host" would leave the site.
		if (redirectUrl !== null && /^\/(?![/\\])/.test(redirectUrl)) {
			redirect(303, redirectUrl);
		}

		redirect(303, HOME_PAGE_PATH);
	}),
	logout: actionWrapper(async ({ cookies }) => {
		cookies.delete(TOKEN_COOKIE_NAME, { path: '/' });

		redirect(303, '/login');
	}),
};
