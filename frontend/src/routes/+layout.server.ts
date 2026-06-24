import { TOKEN_COOKIE_NAME } from '$lib/constants';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
	const loggedIn = cookies.get(TOKEN_COOKIE_NAME) !== undefined;

	return {
		loggedIn,
	};
};
