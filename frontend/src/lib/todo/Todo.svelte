<script lang="ts">
	import Check from '$lib/icons/Check.svelte';
	import Eye from '$lib/icons/Eye.svelte';
	import { todoPriority, type Todo } from './todo';
	import TodoDescription from './TodoDescription.svelte';

	let { todo }: { todo: Todo } = $props();

	let children = $derived(
		todo.children.toSorted(
			(a, b) => todoPriority(a, todo.children.length) - todoPriority(b, todo.children.length),
		),
	);
</script>

<h1>{todo.name}</h1>
<TodoDescription descrption={todo.description} --font-size="1em" />

{#if children.length > 0}
	<h2>Tasks:</h2>
	<ol>
		{#each children as child (child.name)}
			<li>
				<div class="child-container">
					<strong class="child-name">
						{child.name}
					</strong>
					<div class="line"></div>
					<TodoDescription descrption={child.description} lineClamp={2} />
				</div>
				<div class="button-container">
					<a class="view-button" href={`/${child.id}`} aria-label={`View ${child.name}`}
						><Eye --fill="var(--primary-light)" /></a
					>
					<button class="done-button" aria-label={`Complete ${child.name}`}
						><Check --fill="var(--primary-light)" /></button
					>
				</div>
			</li>
		{/each}
	</ol>
{/if}

<style>
	h1 {
		color: var(--primary-color);
		font-size: 2rem;
		margin-bottom: var(--size-3);
	}

	h2 {
		color: var(--primary-color);
		font-size: 1.25rem;
		margin-top: var(--size-5);
		margin-bottom: var(--size-3);
	}

	ol {
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: var(--size-3);
	}

	li {
		list-style-type: none;
		border: var(--border-size-1) solid var(--primary-color);
		border-radius: var(--radius-2);
		min-width: 100%;
		padding: 0;
		overflow: hidden;
	}

	.child-container {
		padding: var(--size-3);
	}

	.child-name {
		color: var(--secondary-color);
		font-size: var(--font-size-2);
	}

	.line {
		height: var(--border-size-1);
		background-color: var(--primary-color);
		margin-block: var(--size-2);
	}

	.button-container {
		display: flex;

		& a,
		button {
			flex: 1;
			display: flex;
			justify-content: center;
			padding: 0.4em;
			font-size: 1em;
		}
	}

	.view-button {
		background: #00509d;
	}

	.done-button {
		background: #4bb543;
	}
</style>
