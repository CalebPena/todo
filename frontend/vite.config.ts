import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	envDir: '../',
	server: {
		watch: {
			ignored: ['**/.svelte-kit-check/**', '**/.svelte-kit-build/**'],
		},
	},
	ssr: {
		noExternal: ['bits-ui'],
	},
	optimizeDeps: {
		exclude: ['bits-ui'],
	},
});
