<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import type { Snippet } from 'svelte';
	import type { Todo } from './todo';
	import CompleteButton from './CompleteButton.svelte';
	import TodoDescription from './TodoDescription.svelte';
	import RankingGraph from './RankingGraph.svelte';

	let {
		todo,
		todos,
		controls,
		forms,
	}: { todo: Todo | null; todos: Todo[]; controls: Snippet; forms: Snippet } = $props();

	let selectedId = $state<number | null>(null);
	let hoveredId = $state<number | null>(null);
	type RankingSort = 'time' | 'value' | 'mix';
	let sortBy = $state<RankingSort>('mix');
	const sortOptions: { value: RankingSort; label: string }[] = [
		{ value: 'mix', label: 'Mix' },
		{ value: 'value', label: 'Value' },
		{ value: 'time', label: 'Time' },
	];
	function rankScore(task: Todo) {
		if (!task.ranked || task.timeRank === null || task.valueRank === null) return Infinity;
		return sortBy === 'time'
			? task.timeRank
			: sortBy === 'value'
				? task.valueRank
				: task.timeRank + task.valueRank;
	}
	let sortedTodos = $derived(
		todos.toSorted(
			(a, b) =>
				Number(b.id === selectedId) - Number(a.id === selectedId) ||
				Number(a.completed) - Number(b.completed) ||
				rankScore(a) - rankScore(b) ||
				0,
		),
	);
</script>

<div class="mb-4 flex flex-col gap-4 sm:flex-row sm:flex-wrap sm:items-start sm:justify-between">
	<div class="w-full sm:order-2 sm:w-auto">{@render controls()}</div>
	<h1
		class={[
			'min-w-0 break-words text-3xl font-semibold tracking-tight text-primary',
			todo?.completed && 'line-through',
		]}
	>
		{todo?.name ?? 'Todos'}
	</h1>
</div>
{#if todo}
	<TodoDescription descrption={todo.description} />
	{#if todo.completed}<Badge variant="secondary" class="mt-2">Completed</Badge>{/if}
{/if}

{@render forms()}

<RankingGraph {todos} {hoveredId} onselect={(id) => (selectedId = id)} />

<div class="mt-6 mb-4 flex flex-wrap items-center justify-between gap-3">
	<h2 class="text-xl font-semibold text-primary">{todo ? 'Subtasks:' : 'Tasks:'}</h2>
	<div
		role="group"
		aria-label="Sort tasks by ranking"
		class="inline-flex items-center rounded-lg bg-muted p-1 text-muted-foreground"
	>
		{#each sortOptions as option (option.value)}
			<Button
				variant="ghost"
				size="sm"
				class={[
					'h-7 rounded-md px-3 hover:text-foreground',
					sortBy === option.value
						? 'bg-background text-foreground shadow-sm hover:bg-background dark:bg-input/30 dark:hover:bg-input/30'
						: 'hover:bg-transparent dark:hover:bg-transparent',
				]}
				aria-pressed={sortBy === option.value}
				onclick={() => {
					sortBy = option.value;
					selectedId = null;
				}}>{option.label}</Button
			>
		{/each}
	</div>
</div>
{#if todos.length > 0}
	<ol class="flex list-none flex-col gap-4 p-0">
		{#each sortedTodos as child (child.id)}
			<li onpointerenter={() => (hoveredId = child.id)} onpointerleave={() => (hoveredId = null)}>
				<Card.Root
					class={[
						'gap-3 rounded-lg py-0',
						child.completed && 'bg-muted opacity-60',
						child.id === selectedId &&
							'border-purple-500 ring-1 ring-purple-500 dark:border-purple-400 dark:ring-purple-400',
					]}
				>
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
										'min-w-0 max-w-full text-xl font-semibold text-primary [overflow-wrap:anywhere]',
										child.completed && 'line-through',
									]}>{child.name}</span
								>
								{#if child.completed}<Badge variant="secondary">Completed</Badge>{/if}
							</div>
							<TodoDescription descrption={child.description} lineClamp={2} />
						</a>
						{#if !child.completed && !child.ranked}
							<div class="flex shrink-0 items-center border-l px-3 sm:w-40 sm:justify-center">
								<Button
									variant="outline"
									href={`/${child.id}/rank`}
									aria-label={`Rank ${child.name}`}>Rank task</Button
								>
							</div>
						{:else if child.ranked}
							<a
								href={`/${child.id}/rank`}
								aria-label={`Re-rank ${child.name}`}
								title="Re-rank task"
								class="shrink-0 rounded-r-lg transition-opacity hover:opacity-80 focus-visible:outline-2 focus-visible:outline-ring"
							>
								<dl class="grid h-full shrink-0 grid-cols-1 border-l text-center sm:grid-cols-2">
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
							</a>
						{/if}
					</Card.Content>
				</Card.Root>
			</li>
		{/each}
	</ol>
{:else}
	<p class="text-sm text-muted-foreground">No {todo ? 'subtasks' : 'todos'} yet.</p>
{/if}
