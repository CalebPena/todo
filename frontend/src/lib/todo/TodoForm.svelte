<script lang="ts">
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import type { Todo } from './todo';

	let {
		action,
		todo,
		onSaved,
	}: {
		action: 'create' | 'edit';
		todo?: Todo;
		onSaved: () => void;
	} = $props();
	let pending = $state(false);
</script>

<form
	class="my-4 max-w-xl"
	method="POST"
	action={`?/${action}`}
	use:enhance={() => {
		pending = true;
		return async ({ result, update }) => {
			try {
				await update();
				if (result.type === 'success') onSaved();
			} finally {
				pending = false;
			}
		};
	}}
>
	<fieldset disabled={pending} class="grid min-w-0 gap-3">
		<Label for={`${action}-name`}>Title</Label>
		<Input id={`${action}-name`} name="name" value={todo?.name ?? ''} required maxlength={255} />
		<Label for={`${action}-description`}>Description</Label>
		<Textarea
			id={`${action}-description`}
			name="description"
			rows={3}
			value={todo?.description ?? ''}
		/>
		<div class="mt-2 flex flex-wrap items-center gap-2">
			<Button type="submit" disabled={pending}
				>{pending ? 'Saving…' : action === 'create' ? 'Create todo' : 'Save changes'}</Button
			>
			{#if action === 'create'}
				<Button type="submit" name="intent" value="rank" variant="outline" disabled={pending}
					>Create and rank</Button
				>
			{/if}
			<Button variant="ghost" type="button" onclick={onSaved} disabled={pending}>Cancel</Button>
		</div>
	</fieldset>
</form>
