<script lang="ts">
	import { goto, afterNavigate } from '$app/navigation';
	import { onMount } from 'svelte';
	import type { TodoSearchResult } from './todo';
	import SearchHighlight from './SearchHighlight.svelte';

	const id = $props.id();
	let root: HTMLDivElement;
	let input: HTMLInputElement;
	let query = $state('');
	let open = $state(false);
	let loading = $state(false);
	let message = $state('');
	let results = $state<TodoSearchResult[]>([]);
	let active = $state(-1);
	let needle = $derived(query.trim());

	$effect(() => {
		const search = needle;
		const visible = open;
		results = [];
		active = -1;
		message = '';
		loading = visible;
		if (!visible) return;
		const controller = new AbortController();
		const timer = window.setTimeout(
			async () => {
				try {
					const response = await fetch(`/api/search?q=${encodeURIComponent(search)}`, {
						signal: controller.signal,
					});
					if (!response.ok)
						throw new Error(
							response.status === 401 || response.status === 403
								? 'Sign in to search your tasks.'
								: 'Search is unavailable. Try again.',
						);
					const data: { todos: TodoSearchResult[] } = await response.json();
					if (!controller.signal.aborted) {
						results = data.todos;
						active = results.length ? 0 : -1;
					}
				} catch (error) {
					if (!controller.signal.aborted)
						message = error instanceof Error ? error.message : 'Search is unavailable. Try again.';
				} finally {
					if (!controller.signal.aborted) loading = false;
				}
			},
			search ? 180 : 0,
		);
		return () => {
			window.clearTimeout(timer);
			controller.abort();
		};
	});

	function close() {
		open = false;
		active = -1;
	}
	function clear() {
		query = '';
		close();
	}
	afterNavigate(clear);
	onMount(() => {
		function outside(event: PointerEvent) {
			if (event.target instanceof Node && !root.contains(event.target)) close();
		}
		window.addEventListener('pointerdown', outside);
		return () => window.removeEventListener('pointerdown', outside);
	});
	function keydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			close();
			return;
		}
		if (event.key === 'Tab') {
			close();
			return;
		}
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();
			open = true;
			if (!results.length) return;
			active = (active + (event.key === 'ArrowDown' ? 1 : -1) + results.length) % results.length;
			document.getElementById(`${id}-${active}`)?.scrollIntoView({ block: 'nearest' });
		} else if (event.key === 'Enter' && open && results[active]) {
			event.preventDefault();
			void goto(`/${results[active].id}`);
			close();
		}
	}
	function snippet(description: string) {
		const index = description.toLocaleLowerCase('en-US').indexOf(needle.toLocaleLowerCase('en-US'));
		const start = Math.max(0, index - 45);
		return `${start ? '…' : ''}${description.slice(start, start + 160)}${description.length > start + 160 ? '…' : ''}`;
	}
</script>

<div bind:this={root} class="relative min-w-0 flex-1 sm:max-w-lg">
	<div
		class="flex h-10 items-center gap-2 rounded-md border border-input bg-background px-3 shadow-xs focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50"
	>
		<svg
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			class="size-4 shrink-0 text-muted-foreground"
			aria-hidden="true"><circle cx="10.5" cy="10.5" r="6.5" /><path d="m16 16 4 4" /></svg
		>
		<input
			bind:this={input}
			bind:value={query}
			oninput={() => (open = true)}
			onfocus={() => (open = true)}
			onclick={() => (open = true)}
			onkeydown={keydown}
			role="combobox"
			aria-label="Search your tasks"
			aria-autocomplete="list"
			aria-expanded={open}
			aria-controls={`${id}-results`}
			aria-activedescendant={open && active >= 0 ? `${id}-${active}` : undefined}
			autocomplete="off"
			maxlength={200}
			placeholder="Search tasks…"
			class="min-w-0 flex-1 bg-transparent text-sm outline-none"
		/>
		{#if query}<button
				type="button"
				aria-label="Clear search"
				class="shrink-0 text-muted-foreground hover:text-foreground"
				onclick={() => {
					clear();
					open = true;
					input.focus();
				}}>×</button
			>{/if}
	</div>
	{#if open}
		<div
			class="results absolute inset-x-0 top-full z-[250] mt-2 overflow-hidden rounded-lg border bg-popover text-popover-foreground shadow-lg"
		>
			<div
				class="px-3 pt-3 pb-2 text-[10px] font-semibold tracking-wider text-muted-foreground uppercase"
			>
				{needle ? 'Your tasks' : 'Favorites'}
			</div>
			<div
				id={`${id}-results`}
				role="listbox"
				aria-label={needle ? 'Search results' : 'Favorite tasks'}
				aria-busy={loading}
				class="max-h-80 overflow-y-auto p-1"
			>
				{#each results as task, index (task.id)}
					<a
						id={`${id}-${index}`}
						href={`/${task.id}`}
						role="option"
						aria-selected={active === index}
						tabindex="-1"
						onpointermove={() => (active = index)}
						onclick={close}
						class={[
							'flex items-start gap-3 rounded-lg px-3 py-2.5 outline-none',
							active === index && 'bg-accent text-accent-foreground',
						]}
					>
						{#if task.favorite}
							<span
								class="mt-0.5 flex size-5 shrink-0 items-center justify-center rounded bg-purple-600 text-amber-300"
								aria-label="Favorite"
								><svg viewBox="0 0 24 24" fill="currentColor" class="size-3.5" aria-hidden="true"
									><path
										d="m12 3 2.78 5.63L21 9.54l-4.5 4.39 1.06 6.2L12 17.2l-5.56 2.93 1.06-6.2L3 9.54l6.22-.91L12 3Z"
									/></svg
								></span
							>
						{:else}<svg
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="1.75"
								class="mt-0.5 size-4 shrink-0 text-muted-foreground"
								aria-hidden="true"
								><rect x="4" y="3" width="16" height="18" rx="3" /><path
									d="M8 8h8M8 12h8M8 16h5"
								/></svg
							>
						{/if}
						<span class="min-w-0 flex-1">
							<span class="block truncate text-sm font-medium"
								><SearchHighlight text={task.name} query={needle} /></span
							>
							{#if task.description}<span class="mt-0.5 line-clamp-2 text-xs text-muted-foreground"
									><SearchHighlight text={snippet(task.description)} query={needle} /></span
								>{/if}
							<span class="mt-1 block truncate text-[10px] text-muted-foreground"
								>{#if task.parentName}In <SearchHighlight
										text={task.parentName}
										query={needle}
									/>{:else}Top-level task{/if}</span
							>
							{#if task.completed}<span class="mt-1 block text-[10px] text-muted-foreground"
									>Completed</span
								>{/if}
						</span>
					</a>
				{:else}<p class="px-3 py-5 text-center text-sm text-muted-foreground" role="status">
						{loading
							? needle
								? 'Searching…'
								: 'Loading favorites…'
							: message ||
								(needle
									? 'No matching tasks.'
									: 'No favorites yet. Favorite a task with its star button.')}
					</p>{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	@media (max-width: 639px) {
		.results {
			position: fixed;
			top: 56px;
			left: 16px;
			right: 16px;
		}
	}
</style>
