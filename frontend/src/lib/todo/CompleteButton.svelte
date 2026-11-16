<script lang="ts">
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import Check from '$lib/icons/Check.svelte';
	import type { Todo } from './todo';

	let { todo, iconOnly = false }: { todo: Todo; iconOnly?: boolean } = $props();
	let pending = $state(false);
</script>

<form
	class="flex flex-1"
	method="POST"
	action="?/complete"
	use:enhance={() => {
		pending = true;
		return async ({ update }) => {
			try {
				await update();
			} finally {
				pending = false;
			}
		};
	}}
>
	<input type="hidden" name="id" value={todo.id} />
	<input type="hidden" name="completed" value={String(!todo.completed)} />
	<Button
		type="submit"
		disabled={pending}
		variant={iconOnly ? 'ghost' : todo.completed ? 'secondary' : 'outline'}
		size={iconOnly ? 'icon' : 'default'}
		class={iconOnly ? 'h-auto w-full rounded-none' : 'flex-1'}
		role={iconOnly ? 'checkbox' : undefined}
		aria-checked={iconOnly ? todo.completed : undefined}
		aria-pressed={iconOnly ? undefined : todo.completed}
		aria-busy={pending}
		aria-label={`${todo.completed ? 'Reopen' : 'Complete'} ${todo.name}`}
		title={todo.completed ? 'Reopen todo' : 'Complete todo'}
	>
		{#if iconOnly}
			<span
				aria-hidden="true"
				class={[
					'flex size-5 items-center justify-center rounded-sm border transition-colors',
					todo.completed
						? 'border-primary bg-primary text-primary-foreground'
						: 'border-muted-foreground bg-background',
				]}
			>
				{#if todo.completed}
					<svg
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="3"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="size-3.5"
					>
						<path d="m5 12 4 4L19 6" />
					</svg>
				{/if}
			</span>
		{:else if todo.completed}<span aria-hidden="true">↶</span>{:else}<Check
				--fill="currentColor"
			/>{/if}
		{#if !iconOnly}<span>{pending ? 'Saving…' : todo.completed ? 'Reopen' : 'Complete'}</span>{/if}
	</Button>
</form>
