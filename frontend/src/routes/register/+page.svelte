<script lang="ts">
	import { alerts } from '$lib/alerts/alerts.svelte';
	import LoginForm from '$lib/auth/LoginForm.svelte';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';
	import { passwordSchema } from '$lib/schemas';
	import { Button } from '$lib/components/ui/button';

	let { form }: PageProps = $props();

	$effect(() => {
		alerts.addAlertFromFormError(form);
	});

	let submitted = $state(false);
	let valid = $state(false);
</script>

<form
	method="POST"
	use:enhance={({ controller }) => {
		submitted = true;

		if (!valid) {
			controller.abort();
		}
	}}
>
	<LoginForm
		title="Create an account"
		submitLabel="Create account"
		{submitted}
		bind:valid
		{passwordSchema}
	>
		{#snippet link()}<Button href="/login" variant="link" class="h-auto p-0"
				>Click here to login</Button
			>{/snippet}
	</LoginForm>
</form>
