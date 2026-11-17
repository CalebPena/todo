<script lang="ts">
	import RankingFlow from '$lib/todo/RankingFlow.svelte';
	import type { PageProps } from './$types';

	let { data, form }: PageProps = $props();
</script>

<svelte:head><title>Rank {data.todo.name} · Todos</title></svelte:head>

{#key data.snapshot}
	<RankingFlow
		{...data}
		message={form && 'message' in form ? form.message : undefined}
		restartRequired={Boolean(form && 'restartRequired' in form && form.restartRequired)}
	/>
{/key}
