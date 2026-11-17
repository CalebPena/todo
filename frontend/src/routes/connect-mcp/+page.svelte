<script lang="ts">
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Alert from '$lib/components/ui/alert';
	import type { SubmitFunction } from '@sveltejs/kit';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	let pending = $state(false);
	let copied = $state('');
	let copyError = $state(false);
	let secret = $derived(form && 'secret' in form ? form.secret : undefined);
	let config = $derived(
		JSON.stringify(
			{
				mcpServers: {
					todo: {
						url: data.endpoint,
						headers: { Authorization: `Bearer ${secret ?? 'YOUR_MCP_TOKEN'}` },
					},
				},
			},
			null,
			2,
		),
	);
	const submit: SubmitFunction = () => {
		pending = true;
		copied = '';
		return async ({ update }) => {
			try {
				await update();
			} finally {
				pending = false;
			}
		};
	};
	async function copy(value: string, label: string) {
		try {
			await navigator.clipboard.writeText(value);
			copied = label;
			copyError = false;
		} catch {
			copyError = true;
		}
	}
</script>

<svelte:head><title>Connect MCP · Todos</title></svelte:head>

<div class="mx-auto max-w-2xl space-y-6">
	<Button variant="ghost" href="/">← Back to tasks</Button>
	<div class="space-y-2">
		<h1 class="text-2xl font-semibold">Connect MCP</h1>
		<p class="text-sm text-muted-foreground">
			Connect your AI assistant to your tasks using a personal MCP token.
		</p>
	</div>
	{#if form && 'message' in form}
		<Alert.Root variant="destructive"
			><Alert.Description>{form.message}</Alert.Description></Alert.Root
		>
	{/if}
	{#if secret}
		<section
			class="space-y-3 rounded-lg border border-purple-300 bg-purple-50 p-4 dark:border-purple-800 dark:bg-purple-950/40"
			aria-label="New MCP token"
		>
			<h2 class="font-semibold">
				Token for {form && 'name' in form ? form.name : 'your assistant'}
			</h2>
			<p class="text-sm text-muted-foreground">
				Copy it now. The full token cannot be retrieved again.
			</p>
			<div class="flex gap-2">
				<Input
					aria-label="New MCP token"
					value={secret}
					readonly
					class="min-w-0 font-mono text-xs"
				/>
				<Button variant="outline" onclick={() => copy(secret, 'token')}
					>{copied === 'token' ? 'Copied' : 'Copy'}</Button
				>
			</div>
		</section>
	{/if}
	<section class="space-y-3 rounded-lg border p-4">
		<h2 class="font-semibold">Create a token</h2>
		<p class="text-sm text-muted-foreground">
			Tokens can read and manage all your tasks, including deleting them. Use one token per
			assistant.
		</p>
		<form method="POST" action="?/create" use:enhance={submit} class="space-y-3">
			<Label for="token-name">Token name</Label>
			<div class="flex flex-wrap gap-2">
				<Input
					id="token-name"
					name="name"
					required
					maxlength={80}
					placeholder="e.g. Claude"
					class="min-w-0 flex-1"
					disabled={pending}
				/>
				<Button type="submit" disabled={pending}>Create token</Button>
			</div>
		</form>
	</section>
	<section class="space-y-3">
		<h2 class="font-semibold">Your tokens</h2>
		{#each data.tokens as token (token.id)}
			<div
				class="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-4"
				class:opacity-60={token.revokedAt !== null}
			>
				<div class="min-w-0 space-y-1">
					<p class="break-words text-sm font-medium">
						{token.name}{token.revokedAt ? ' · Revoked' : ''}
					</p>
					<p class="font-mono text-xs text-muted-foreground">{token.prefix}…</p>
					<p class="text-xs text-muted-foreground">
						Created {token.createdAt.slice(0, 10)} · {token.lastUsedAt
							? `Last used ${token.lastUsedAt.slice(0, 10)}`
							: 'Not used yet'}
					</p>
				</div>
				{#if !token.revokedAt}
					<div class="flex gap-2">
						<form method="POST" action="?/rotate" use:enhance={submit}>
							<input type="hidden" name="id" value={token.id} />
							<Button type="submit" variant="outline" size="sm" disabled={pending}>Rotate</Button>
						</form>
						<form method="POST" action="?/revoke" use:enhance={submit}>
							<input type="hidden" name="id" value={token.id} />
							<Button type="submit" variant="destructive" size="sm" disabled={pending}
								>Revoke</Button
							>
						</form>
					</div>
				{/if}
			</div>
		{:else}<p class="text-sm text-muted-foreground">No MCP tokens yet.</p>{/each}
		<p class="text-xs text-muted-foreground">
			Rotating replaces a token immediately. Revoking disconnects any assistant using it.
		</p>
	</section>
	<section class="space-y-3 rounded-lg border p-4">
		<h2 class="font-semibold">Connect your assistant</h2>
		<Label for="mcp-endpoint">MCP server URL · Streamable HTTP</Label>
		<Input id="mcp-endpoint" value={data.endpoint} readonly class="font-mono text-xs" />
		<p class="text-sm text-muted-foreground">
			Use your token as a Bearer token. For clients that support JSON configuration:
		</p>
		<pre class="overflow-x-auto rounded-md bg-muted p-3 text-xs"><code>{config}</code></pre>
		<Button variant="outline" onclick={() => copy(config, 'config')}
			>{copied === 'config' ? 'Copied' : 'Copy configuration'}</Button
		>
		{#if copyError}<p class="text-sm text-muted-foreground" role="status">
				Select and copy the text manually.
			</p>{/if}
	</section>
</div>
