<script lang="ts">
	import Alerts from '$lib/alerts/Alerts.svelte';
	import { Button } from '$lib/components/ui/button';
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
			<form method="POST" action="/login?/logout">
				<Button type="submit" variant="outline">Log out</Button>
			</form>
		{/if}
	</nav>
</header>
<main class="mx-auto max-w-7xl p-4">{@render children()}</main>
