import { actionWrapper } from '$lib/api/helpers';
import { TOKEN_COOKIE_NAME } from '$lib/constants';
import { createMCPToken, listMCPTokens, revokeMCPToken, rotateMCPToken } from '$lib/mcp/api.server';
import { validateApiInput } from '$lib/schemas';
import { redirect, type Action, type Cookies } from '@sveltejs/kit';
import { z } from 'zod';
import type { Actions, PageServerLoad } from './$types';

function session(cookies: Cookies) {
	const token = cookies.get(TOKEN_COOKIE_NAME);
	if (!token) redirect(303, '/login?redirect=%2Fconnect-mcp');
	return token;
}

const tokenId = z.coerce.number().int().positive().max(2147483647);

function tokenAction(handler: Action): Action {
	return actionWrapper(async (event) => {
		// Enhanced submissions return JSON without running the page load.
		if (event.request.headers.get('x-sveltekit-action') === 'true') {
			event.setHeaders({ 'Cache-Control': 'private, no-store' });
		}
		return handler(event);
	});
}

export const load: PageServerLoad = async ({ cookies, url, setHeaders }) => {
	setHeaders({ 'Cache-Control': 'private, no-store', 'Referrer-Policy': 'no-referrer' });
	return { tokens: await listMCPTokens(session(cookies)), endpoint: `${url.origin}/mcp` };
};

export const actions: Actions = {
	create: tokenAction(async ({ cookies, request }) => {
		const token = session(cookies);
		const form = await request.formData();
		const name = validateApiInput(
			z
				.string()
				.trim()
				.min(1, 'Enter a token name')
				.refine((name) => new TextEncoder().encode(name).length <= 80, 'Use a shorter name'),
			form.get('name'),
		);
		const issued = await createMCPToken(token, name);
		return { secret: issued.secret, name: issued.token.name };
	}),
	rotate: tokenAction(async ({ cookies, request }) => {
		const token = session(cookies);
		const id = validateApiInput(tokenId, (await request.formData()).get('id'));
		const issued = await rotateMCPToken(token, id);
		return { secret: issued.secret, name: issued.token.name };
	}),
	revoke: tokenAction(async ({ cookies, request }) => {
		const token = session(cookies);
		const id = validateApiInput(tokenId, (await request.formData()).get('id'));
		await revokeMCPToken(token, id);
		return { revoked: true };
	}),
};
