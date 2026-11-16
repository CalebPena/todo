<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { alerts } from '$lib/alerts/alerts.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Alert from '$lib/components/ui/alert';
	import * as Dialog from '$lib/components/ui/dialog';
	import CompleteButton from '$lib/todo/CompleteButton.svelte';
	import Todo from '$lib/todo/Todo.svelte';
	import TodoForm from '$lib/todo/TodoForm.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	let creating = $state(false);
	let editing = $state(false);

	$effect(() => {
		alerts.addAlertFromFormError(form);
	});
	afterNavigate(() => {
		creating = false;
		editing = false;
	});
</script>

<svelte:head><title>{data.todo?.name ?? 'Todos'}</title></svelte:head>

<Dialog.Root
	bind:open={creating}
	onOpenChange={(open) => {
		if (open) editing = false;
	}}
>
	<Todo todo={data.todo} todos={data.todos}>
		{#snippet controls()}
			<div
				class="ml-auto flex shrink-0 flex-wrap items-center justify-end gap-2"
				role="group"
				aria-label="Task actions"
			>
				{#if data.todo}
					<Button
						variant="outline"
						href={data.todo.parentId === null ? '/' : `/${data.todo.parentId}`}
					>
						← {data.todo.parentId === null ? 'All todos' : 'Parent todo'}
					</Button>
				{/if}
				<Dialog.Trigger>
					{#snippet child({ props })}
						<Button {...props}>{data.todo ? 'Add subtask' : 'Add todo'}</Button>
					{/snippet}
				</Dialog.Trigger>
				{#if data.todo}
					<Dialog.Root
						bind:open={editing}
						onOpenChange={(open) => {
							if (open) creating = false;
						}}
					>
						<Dialog.Trigger>
							{#snippet child({ props })}<Button variant="outline" {...props}>Edit todo</Button
								>{/snippet}
						</Dialog.Trigger>
						<Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-lg">
							<Dialog.Header>
								<Dialog.Title>Edit todo</Dialog.Title>
							</Dialog.Header>
							<TodoForm action="edit" todo={data.todo} onSaved={() => (editing = false)} />
							{#if form && 'message' in form}<Alert.Root variant="destructive"
									><Alert.Description>{form.message}</Alert.Description></Alert.Root
								>{/if}
						</Dialog.Content>
					</Dialog.Root>
					<div class="flex"><CompleteButton todo={data.todo} /></div>
				{/if}
			</div>
		{/snippet}
		{#snippet forms()}
			{#if !creating && !editing && form && 'message' in form}<Alert.Root
					variant="destructive"
					class="mt-4"><Alert.Description>{form.message}</Alert.Description></Alert.Root
				>{/if}
		{/snippet}
	</Todo>
	<Dialog.Content class="max-h-[85vh] overflow-y-auto sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>{data.todo ? 'Add subtask' : 'Add todo'}</Dialog.Title>
			{#if data.todo}
				<Dialog.Description>Create a subtask for {data.todo.name}.</Dialog.Description>
			{/if}
		</Dialog.Header>
		<TodoForm action="create" onSaved={() => (creating = false)} />
		{#if form && 'message' in form}<Alert.Root variant="destructive"
				><Alert.Description>{form.message}</Alert.Description></Alert.Root
			>{/if}
	</Dialog.Content>
</Dialog.Root>
