<script lang="ts">
	let { descrption, lineClamp }: { descrption: string; lineClamp?: number } = $props();
</script>

<p class:line-clamp={lineClamp !== undefined} style:--line-clamp={lineClamp}>{descrption}</p>

<style>
	p {
		white-space: pre-wrap;
		font-size: var(--font-size);
	}

	.line-clamp {
		display: -webkit-box;
		line-clamp: var(--line-clamp);
		-webkit-line-clamp: var(--line-clamp);
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
