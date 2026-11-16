<script lang="ts">
	import { fly } from 'svelte/transition';
	import type { Alert } from './alerts.svelte';
	import X from '$lib/icons/X.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as AlertBox from '$lib/components/ui/alert';

	const { alert }: { alert: Alert } = $props();

	let message = $derived.by(() => {
		let formattedAmount: string;
		if (alert.count <= 1) {
			formattedAmount = '';
		} else if (alert.count <= 10) {
			formattedAmount = ' (' + alert.count + ')';
		} else {
			formattedAmount = ' (10+)';
		}

		return alert.message + formattedAmount;
	});
</script>

<div transition:fly={{ opacity: 1, x: '120%' }} class="w-96 max-w-full">
	<AlertBox.Root
		variant={alert.type === 'error' ? 'destructive' : 'default'}
		class="pr-12 shadow-sm"
	>
		<Button
			variant="ghost"
			size="icon-sm"
			class="absolute top-1 right-1"
			aria-label="Dismiss notification"
			onfocusin={() => {
				alert.pause();
			}}
			onfocusout={() => {
				alert.unpause();
			}}
			onclick={() => {
				alert.remove();
			}}
		>
			<X --fill="currentColor" />
		</Button>
		<AlertBox.Description>{message}</AlertBox.Description>
	</AlertBox.Root>
</div>
