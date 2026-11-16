<script lang="ts">
	import Alert from './Alert.svelte';
	import { alerts } from './alerts.svelte';
</script>

<!-- svelte-ignore a11y_mouse_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="fixed top-4 right-4 z-50 flex max-w-[calc(100vw-2rem)] flex-col gap-3"
	onmouseover={() => {
		alerts.pauseAll();
	}}
	onmouseleave={() => {
		alerts.unpauseAll();
	}}
>
	{#each alerts.alerts as alert (alert.id)}
		<Alert {alert} />
	{/each}
</div>
