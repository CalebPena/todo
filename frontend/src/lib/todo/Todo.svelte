<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import * as Card from '$lib/components/ui/card';
	import type { Snippet } from 'svelte';
	import type { Todo } from './todo';
	import CompleteButton from './CompleteButton.svelte';
	import TodoDescription from './TodoDescription.svelte';

	let {
		todo,
		todos,
		controls,
		forms,
	}: { todo: Todo | null; todos: Todo[]; controls: Snippet; forms: Snippet } = $props();

	let sortedTodos = $derived(todos.toSorted((a, b) => Number(a.completed) - Number(b.completed)));
</script>

<div class="mb-4 flex flex-wrap items-start justify-between gap-4">
	<h1
		class={[
			'min-w-0 break-words text-3xl font-semibold tracking-tight text-primary',
			todo?.completed && 'line-through',
		]}
	>
		{todo?.name ?? 'Todos'}
	</h1>
	{@render controls()}
</div>
{#if todo}
	<TodoDescription descrption={todo.description} />
	{#if todo.completed}<Badge variant="secondary" class="mt-2">Completed</Badge>{/if}
{/if}

{@render forms()}

<h2 class="mt-6 mb-4 text-xl font-semibold text-primary">{todo ? 'Subtasks:' : 'Tasks:'}</h2>
{#if todos.length > 0}
	<ol class="flex list-none flex-col gap-4 p-0">
		{#each sortedTodos as child (child.id)}
			<li>
				<Card.Root class={['gap-3 rounded-lg py-0', child.completed && 'bg-muted opacity-60']}>
					<Card.Content class="flex items-stretch px-0">
						<div class="flex w-12 shrink-0 items-stretch border-r border-border">
							<CompleteButton todo={child} iconOnly />
						</div>
						<a
							href={`/${child.id}`}
							aria-label={`View ${child.name}`}
							class="min-w-0 flex-1 px-4 py-3 transition-colors hover:bg-muted/50 focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-ring"
						>
							<div class="mb-2 flex min-w-0 flex-wrap items-center gap-2">
								<span
									class={[
										'break-words text-xl font-semibold text-primary',
										child.completed && 'line-through',
									]}>{child.name}</span
								>
								{#if child.completed}<Badge variant="secondary">Completed</Badge>{/if}
							</div>
							<TodoDescription descrption={child.description} lineClamp={2} />
						</a>
						<dl class="grid shrink-0 grid-cols-1 border-l text-center sm:grid-cols-2">
							<div
								class="flex min-w-16 flex-col justify-center gap-1 bg-rose-50 px-3 py-3 text-rose-700 dark:bg-rose-950 dark:text-rose-300 sm:min-w-20"
								title="Value rank among siblings. 1 is most valuable."
							>
								<dt class="text-[10px] font-semibold tracking-wider uppercase">Value</dt>
								<dd class="text-2xl leading-none font-bold tabular-nums">
									{child.valueRank === null ? '—' : `#${child.valueRank}`}
								</dd>
							</div>
							<div
								class="flex min-w-16 flex-col justify-center gap-1 border-t bg-blue-50 px-3 py-3 text-blue-700 dark:bg-blue-950 dark:text-blue-300 sm:min-w-20 sm:border-t-0 sm:border-l"
								title="Time rank among siblings. 1 takes the least time."
							>
								<dt class="text-[10px] font-semibold tracking-wider uppercase">Time</dt>
								<dd class="text-2xl leading-none font-bold tabular-nums">
									{child.timeRank === null ? '—' : `#${child.timeRank}`}
								</dd>
							</div>
						</dl>
					</Card.Content>
				</Card.Root>
			</li>
		{/each}
	</ol>
{:else}
	<p class="text-sm text-muted-foreground">No {todo ? 'subtasks' : 'todos'} yet.</p>
{/if}
