import { BACKEND_DOMAIN } from '$env/static/private';
import { checkResponse, requestApi } from '$lib/api/helpers';
import { responseSchemaWrapper } from '$lib/schemas';
import { z } from 'zod';

const tokenSchema = z.object({
	id: z.number().int().positive(),
	name: z.string(),
	prefix: z.string(),
	createdAt: z.string(),
	lastUsedAt: z.string().nullable(),
	revokedAt: z.string().nullable(),
});
const issuedTokenSchema = responseSchemaWrapper(
	z.object({ token: tokenSchema, secret: z.string() }),
);

export async function listMCPTokens(session: string) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/mcp-tokens`,
		responseSchemaWrapper(z.object({ tokens: z.array(tokenSchema) })),
		'GET',
		undefined,
		session,
	);
	return data.response.tokens;
}

export async function createMCPToken(session: string, name: string) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/mcp-tokens`,
		issuedTokenSchema,
		'POST',
		{ name },
		session,
	);
	return data.response;
}

export async function rotateMCPToken(session: string, id: number) {
	const { data } = await requestApi(
		`${BACKEND_DOMAIN}/mcp-tokens/${id}/rotate`,
		issuedTokenSchema,
		'POST',
		undefined,
		session,
	);
	return data.response;
}

export async function revokeMCPToken(session: string, id: number) {
	const response = await fetch(`${BACKEND_DOMAIN}/mcp-tokens/${id}`, {
		method: 'DELETE',
		headers: { Authorization: `Bearer ${session}` },
	});
	await checkResponse(response);
}
