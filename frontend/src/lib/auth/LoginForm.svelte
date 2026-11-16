<script lang="ts">
	import { formInputError } from '$lib/utils/forms';
	import { emailSchema } from '$lib/schemas';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';
	import type { Snippet } from 'svelte';
	import type { Schema } from 'zod';

	let {
		title,
		link,
		submitted,
		passwordSchema,
		valid = $bindable(),
		submitLabel = 'Login',
	}: {
		title: string;
		link: Snippet;
		submitted: boolean;
		passwordSchema: Schema;
		valid: boolean;
		submitLabel?: string;
	} = $props();
	let email = $state('');
	let password = $state('');
	let emailErrorMessage = $derived(formInputError(emailSchema, email));
	let passwordErrorMessage = $derived(formInputError(passwordSchema, password));
	$effect(() => {
		valid = emailErrorMessage === '' && passwordErrorMessage === '';
	});
</script>

<div class="flex min-h-[75vh] items-center justify-center py-8">
	<Card.Root class="w-full max-w-sm">
		<Card.Header><Card.Title class="text-xl">{title}</Card.Title></Card.Header>
		<Card.Content class="space-y-5">
			<div class="space-y-2">
				<Label for="email">Email</Label>
				<Input
					type="email"
					id="email"
					name="email"
					bind:value={email}
					autocomplete="email"
					aria-invalid={submitted && emailErrorMessage !== ''}
					aria-describedby={submitted && emailErrorMessage ? 'email-error' : undefined}
				/>
				{#if emailErrorMessage !== '' && submitted}<p
						id="email-error"
						class="text-sm text-destructive"
					>
						{emailErrorMessage}
					</p>{/if}
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					type="password"
					id="password"
					name="password"
					bind:value={password}
					autocomplete={submitLabel === 'Login' ? 'current-password' : 'new-password'}
					aria-invalid={submitted && passwordErrorMessage !== ''}
					aria-describedby={submitted && passwordErrorMessage ? 'password-error' : undefined}
				/>
				{#if passwordErrorMessage !== '' && submitted}<p
						id="password-error"
						class="text-sm text-destructive"
					>
						{passwordErrorMessage}
					</p>{/if}
			</div>
			<div class="text-sm">{@render link()}</div>
			<Button type="submit" class="w-full">{submitLabel}</Button>
		</Card.Content>
	</Card.Root>
</div>
