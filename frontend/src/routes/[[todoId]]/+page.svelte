<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import { enhance } from '$app/forms';
	import { DropdownMenu } from 'bits-ui';
	import { alerts } from '$lib/alerts/alerts.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Alert from '$lib/components/ui/alert';
	import * as Dialog from '$lib/components/ui/dialog';
	import Todo from '$lib/todo/Todo.svelte';
	import TodoForm from '$lib/todo/TodoForm.svelte';
	import FavoriteButton from '$lib/todo/FavoriteButton.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
	let creating = $state(false);
	let editing = $state(false);
	let completing = $state(false);
	let completionForm = $state<HTMLFormElement>();

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
				class="ml-auto flex w-full min-w-0 flex-wrap items-center justify-end gap-2 sm:w-auto sm:shrink-0"
				role="group"
				aria-label="Task actions"
			>
				{#if data.todo}
					<Button
						variant="outline"
						class="mr-auto"
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
					<FavoriteButton todo={data.todo} />
					<Dialog.Root
						bind:open={editing}
						onOpenChange={(open) => {
							if (open) creating = false;
						}}
					>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								{#snippet child({ props })}<Button
										{...props}
										variant="outline"
										size="icon"
										aria-label="Task options">⋯</Button
									>{/snippet}
							</DropdownMenu.Trigger>
							<DropdownMenu.Portal>
								<DropdownMenu.Content
									align="end"
									sideOffset={6}
									class="z-50 min-w-40 rounded-md border bg-popover p-1 text-popover-foreground shadow-md"
								>
									<DropdownMenu.Item
										onclick={() => (editing = true)}
										class="cursor-default rounded-sm px-3 py-2 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
										>Edit</DropdownMenu.Item
									>
									{#if !data.todo.completed}
										<DropdownMenu.Item>
											{#snippet child({ props })}
												<a
													{...props}
													href={`/${data.todo.id}/rank`}
													class="block cursor-default rounded-sm px-3 py-2 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
													>{data.todo.ranked ? 'Re-rank' : 'Rank task'}</a
												>
											{/snippet}
										</DropdownMenu.Item>
									{/if}
									<DropdownMenu.Item
										disabled={completing}
										onclick={() => completionForm?.requestSubmit()}
										class="cursor-default rounded-sm px-3 py-2 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:opacity-50"
										>{completing
											? 'Saving…'
											: data.todo.completed
												? 'Reopen'
												: 'Complete'}</DropdownMenu.Item
									>
								</DropdownMenu.Content>
							</DropdownMenu.Portal>
						</DropdownMenu.Root>
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
					<form
						bind:this={completionForm}
						method="POST"
						action="?/complete"
						class="hidden"
						use:enhance={() => {
							completing = true;
							return async ({ update }) => {
								try {
									await update();
								} finally {
									completing = false;
								}
							};
						}}
					>
						<input type="hidden" name="id" value={data.todo.id} />
						<input type="hidden" name="completed" value={String(!data.todo.completed)} />
					</form>
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
