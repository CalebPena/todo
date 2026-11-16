<script lang="ts">
	let { descrption, lineClamp }: { descrption: string; lineClamp?: number } = $props();
</script>

<p
	class="whitespace-pre-wrap break-words text-sm leading-6 text-muted-foreground"
	class:line-clamp={lineClamp !== undefined}
	style:--line-clamp={lineClamp}
>
	{descrption}
</p>

<style>
	.line-clamp {
		display: -webkit-box;
		line-clamp: var(--line-clamp);
		-webkit-line-clamp: var(--line-clamp);
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
