<script lang="ts">
	import Alerts from '$lib/alerts/Alerts.svelte';
	import { HOME_PAGE_PATH } from '$lib/constants';
	import type { LayoutProps } from './$types';
	import './global.css';

	let { children, data }: LayoutProps = $props();
</script>

<Alerts />

<nav>
	<span>
		<a href={HOME_PAGE_PATH}>Todos</a>
	</span>
	<span>
		{#if data.loggedIn}
			<form method="POST" action="/login?/logout"><button class="link">logout</button></form>
		{/if}
	</span>
</nav>

<main>
	{@render children()}
</main>

<style>
	nav {
		display: flex;
		justify-content: space-between;
		background-color: var(--primary-background-color);
		color: var(--primary-color);
		border-bottom: var(--border-size-1) solid var(--primary-dark);
		font-weight: var(--font-weight-7);
		padding: var(--size-3);
	}

	button {
		background-color: transparent;
	}

	main {
		padding: var(--size-3);
		max-width: 80rem;
		margin: auto;
	}
</style>
