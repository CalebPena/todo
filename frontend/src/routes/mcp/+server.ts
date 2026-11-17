import { BACKEND_DOMAIN } from '$env/static/private';
import type { RequestHandler } from './$types';

const proxy: RequestHandler = async ({ request, url }) => {
	const origin = request.headers.get('origin');
	if (origin && origin !== url.origin) return new Response('Forbidden origin', { status: 403 });
	const headers = new Headers();
	for (const name of ['authorization', 'content-type', 'accept', 'mcp-protocol-version']) {
		const value = request.headers.get(name);
		if (value) headers.set(name, value);
	}
	let body: Uint8Array<ArrayBuffer> | undefined;
	if (request.method === 'POST' && request.body) {
		const reader = request.body.getReader();
		const chunks: Uint8Array[] = [];
		let size = 0;
		try {
			while (true) {
				const { done, value } = await reader.read();
				if (done) break;
				size += value.byteLength;
				if (size > 1_048_576) {
					await reader.cancel();
					return new Response('Request too large', { status: 413 });
				}
				chunks.push(value);
			}
		} finally {
			reader.releaseLock();
		}
		body = new Uint8Array(size);
		let offset = 0;
		for (const chunk of chunks) {
			body.set(chunk, offset);
			offset += chunk.byteLength;
		}
	}
	const response = await fetch(`${BACKEND_DOMAIN}/mcp`, {
		method: request.method,
		headers,
		body,
		signal: request.signal,
		redirect: 'manual',
	});
	const responseHeaders = new Headers({ 'Cache-Control': 'no-store' });
	for (const name of ['content-type', 'www-authenticate', 'allow', 'mcp-protocol-version']) {
		const value = response.headers.get(name);
		if (value) responseHeaders.set(name, value);
	}
	return new Response(response.body, { status: response.status, headers: responseHeaders });
};

export const POST = proxy;
export const GET = proxy;
export const DELETE = proxy;
