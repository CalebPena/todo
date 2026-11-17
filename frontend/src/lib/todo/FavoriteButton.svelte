<script lang="ts">
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import type { Todo } from './todo';

	let { todo }: { todo: Todo } = $props();
	let pending = $state(false);
</script>

<form
	method="POST"
	action="?/favorite"
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
	<input type="hidden" name="favorite" value={String(!todo.favorite)} />
	<Button
		type="submit"
		variant="outline"
		size="icon"
		disabled={pending}
		aria-pressed={todo.favorite}
		aria-label={`${todo.favorite ? 'Remove from favorites' : 'Favorite'}: ${todo.name}`}
		title={todo.favorite ? 'Remove from favorites' : 'Add to favorites'}
		class={[
			todo.favorite
				? 'border-purple-600 bg-purple-600 text-amber-300 hover:bg-purple-700 hover:text-amber-300 dark:border-purple-600 dark:bg-purple-600 dark:hover:bg-purple-700'
				: 'text-muted-foreground hover:text-purple-600',
		]}
	>
		<svg
			viewBox="0 0 24 24"
			fill={todo.favorite ? 'currentColor' : 'none'}
			stroke="currentColor"
			stroke-width="1.75"
			class="size-5"
			aria-hidden="true"
			><path
				d="m12 3 2.78 5.63L21 9.54l-4.5 4.39 1.06 6.2L12 17.2l-5.56 2.93 1.06-6.2L3 9.54l6.22-.91L12 3Z"
			/></svg
		>
	</Button>
</form>
