<script lang="ts">
	import { alerts } from '$lib/alerts/alerts.svelte';
	import LoginForm from '$lib/auth/LoginForm.svelte';
	import type { PageProps } from './$types';
	import { enhance } from '$app/forms';
	import { loginPasswordSchema } from '$lib/schemas';
	import { page } from '$app/state';
	import { REDIRECT_PARAM_NAME } from '$lib/constants';
	import { Button } from '$lib/components/ui/button';

	let { form }: PageProps = $props();

	$effect(() => {
		alerts.addAlertFromFormError(form);
	});

	let submitted = $state(false);
	let valid = $state(false);

	// A bare "?/login" would drop the redirect param from the current URL.
	let action = $derived.by(() => {
		const redirectUrl = page.url.searchParams.get(REDIRECT_PARAM_NAME);

		if (redirectUrl === null) {
			return '?/login';
		}

		return `?/login&${new URLSearchParams({ [REDIRECT_PARAM_NAME]: redirectUrl })}`;
	});
</script>

<form
	method="POST"
	{action}
	use:enhance={({ controller }) => {
		submitted = true;

		if (!valid) {
			controller.abort();
		}
	}}
>
	<LoginForm title="Login" {submitted} bind:valid passwordSchema={loginPasswordSchema}>
		{#snippet link()}<Button href="/register" variant="link" class="h-auto p-0"
				>Click here to create an account</Button
			>{/snippet}
	</LoginForm>
</form>
