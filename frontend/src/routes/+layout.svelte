<script lang="ts">
	import Alerts from '$lib/alerts/Alerts.svelte';
	import { Button } from '$lib/components/ui/button';
	import { DropdownMenu } from 'bits-ui';
	import { HOME_PAGE_PATH } from '$lib/constants';
	import type { LayoutProps } from './$types';
	import './global.css';

	let { children, data }: LayoutProps = $props();
</script>

<Alerts />
<header class="border-b bg-muted/40">
	<nav
		aria-label="Main navigation"
		class="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-4 px-4 py-2"
	>
		<Button href={HOME_PAGE_PATH} variant="link" class="h-auto p-0 text-xl font-bold tracking-tight"
			>Todos</Button
		>
		{#if data.loggedIn}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button
							{...props}
							variant="outline"
							size="icon"
							class="rounded-full"
							aria-label="Profile menu"
						>
							<svg
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="1.75"
								class="size-5"
								aria-hidden="true"
								><circle cx="12" cy="8" r="4" /><path d="M4 21v-2a8 8 0 0 1 16 0v2" /></svg
							>
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Portal>
					<DropdownMenu.Content
						align="end"
						sideOffset={6}
						class="z-50 min-w-44 rounded-md border bg-popover p-1 text-popover-foreground shadow-md"
					>
						<DropdownMenu.Item>
							{#snippet child({ props })}<a
									{...props}
									href="/connect-mcp"
									class="block cursor-default rounded-sm px-3 py-2 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
									>Connect MCP</a
								>{/snippet}
						</DropdownMenu.Item>
						<DropdownMenu.Separator class="my-1 h-px bg-border" />
						<form method="POST" action="/login?/logout">
							<DropdownMenu.Item>
								{#snippet child({ props })}<button
										{...props}
										type="submit"
										class="block w-full cursor-default rounded-sm px-3 py-2 text-left text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
										>Log out</button
									>{/snippet}
							</DropdownMenu.Item>
						</form>
					</DropdownMenu.Content>
				</DropdownMenu.Portal>
			</DropdownMenu.Root>
		{/if}
	</nav>
</header>
<main class="mx-auto max-w-7xl p-4">{@render children()}</main>
