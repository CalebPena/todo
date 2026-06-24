<script lang="ts">
	import Todo from '$lib/todo/Todo.svelte';
	import type { PageProps } from './$types';

	const { data }: PageProps = $props();
</script>

<Todo todo={data.todo} />
