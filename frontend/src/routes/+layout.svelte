<script lang="ts">
	import Alerts from '$lib/alerts/Alerts.svelte';
	import { HOME_PAGE_PATH } from '$lib/constants';
	import type { LayoutProps } from './$types';
	import './global.css';

	let { children, data }: LayoutProps = $props();
</script>

<Alerts />

<header>
	<nav aria-label="Main navigation">
		<a class="home-link" href={HOME_PAGE_PATH}>Todos</a>
		{#if data.loggedIn}
			<form method="POST" action="/login?/logout">
				<button class="logout-button">Log out</button>
			</form>
		{/if}
	</nav>
</header>

<main>
	{@render children()}
</main>

<style>
	header {
		background-color: var(--primary-background-color);
		border-bottom: 1px solid #e5e0ef;
	}

	nav {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--size-3);
		max-width: 80rem;
		min-height: 4rem;
		margin: auto;
		padding: var(--size-2) var(--size-3);
	}

	.home-link {
		color: var(--primary-color);
		font-size: 1.25rem;
		font-weight: var(--font-weight-7);
		letter-spacing: -0.025em;
		text-decoration: none;
	}

	.home-link:hover {
		text-decoration: underline;
		text-underline-offset: 4px;
	}

	.logout-button {
		background-color: transparent;
		color: var(--primary-color);
		border: 1px solid #ddd5ef;
		border-radius: var(--radius-2);
		padding: 0.45rem 0.8rem;
		font-size: 0.875rem;
		font-weight: var(--font-weight-5);
		cursor: pointer;
	}

	.logout-button:hover {
		background-color: #eee9f8;
		border-color: var(--primary-color);
	}

	main {
		padding: var(--size-3);
		max-width: 80rem;
		margin: auto;
	}
</style>
