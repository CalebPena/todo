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
<TodoDescription descrption={todo.description} --font-size="1.3em" />

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
					<a class="view-button" href={`/${child.id}`}><Eye --fill="var(--primary-light)" /></a>
					<button class="done-button"><Check --fill="var(--primary-light)" /></button>
				</div>
			</li>
		{/each}
	</ol>
{/if}

<style>
	h1 {
		color: var(--primary-color);
		margin-bottom: var(--size-3);
	}

	h2 {
		color: var(--primary-color);
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
		border: var(--border-size-3) solid var(--primary-color);
		border-radius: var(--radius-3);
		min-width: 100%;
		padding: 0;
		overflow: hidden;
	}

	.child-container {
		padding: var(--size-2);
	}

	.child-name {
		color: var(--secondary-color);
		font-size: var(--font-size-3);
	}

	.line {
		height: var(--border-size-3);
		background-color: var(--primary-color);
		border-radius: var(--radius-round);
	}

	.button-container {
		display: flex;

		& a,
		button {
			flex: 1;
			display: flex;
			justify-content: center;
			padding: 0.2em;
			font-size: 1.5em;
		}
	}

	.view-button {
		background: #00509d;
	}

	.done-button {
		background: #4bb543;
	}
</style>
