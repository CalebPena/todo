<script lang="ts">
	let { text, query }: { text: string; query: string } = $props();
	let parts = $derived.by(() => {
		const needle = query.trim().toLocaleLowerCase('en-US');
		if (!needle) return [{ text, match: false }];
		const normalized = text.toLocaleLowerCase('en-US');
		const result: { text: string; match: boolean }[] = [];
		let cursor = 0;
		let index = normalized.indexOf(needle);
		while (index !== -1) {
			result.push({ text: text.slice(cursor, index), match: false });
			result.push({ text: text.slice(index, index + needle.length), match: true });
			cursor = index + needle.length;
			index = normalized.indexOf(needle, cursor);
		}
		result.push({ text: text.slice(cursor), match: false });
		return result;
	});
</script>

{#each parts as part, index (index)}
	{#if part.match}<mark class="rounded-sm bg-purple-500/15 font-semibold text-foreground"
			>{part.text}</mark
		>{:else}{part.text}{/if}
{/each}
