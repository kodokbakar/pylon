<script lang="ts">
	const services = ['API Gateway', 'Chat Service', 'Presence Service', 'Room Service', 'Notification Service'];
</script>

<svelte:head>
	<title>Pylon Web</title>
	<meta
		name="description"
		content="Frontend scaffold for the Pylon distributed real-time chat system."
	/>
</svelte:head>

<main class="min-h-screen bg-zinc-950 px-6 py-10 text-zinc-100">
	<section class="mx-auto grid max-w-5xl gap-10 border border-zinc-800 p-6 md:grid-cols-[1.2fr_0.8fr] md:p-10">
		<div>
			<p class="mb-4 text-sm font-semibold uppercase tracking-[0.28em] text-lime-300">
				Pylon Web
			</p>

			<h1 class="max-w-3xl text-5xl font-black tracking-tight text-white md:text-7xl">
				Realtime chat control surface.
			</h1>

			<p class="mt-6 max-w-xl text-lg leading-8 text-zinc-400">
				SvelteKit, Tailwind CSS v4, and Bun are wired for the frontend foundation.
			</p>
		</div>

		<div class="border-lime-300 bg-lime-300 p-5 text-zinc-950 md:self-end">
			<p class="text-sm font-bold uppercase tracking-[0.24em]">Scaffold status</p>
			<p class="mt-6 text-4xl font-black">Ready</p>
		</div>
	</section>

	<section class="mx-auto mt-8 grid max-w-5xl gap-px overflow-hidden border border-zinc-800 bg-zinc-800 md:grid-cols-5">
		{#each services as service}
			<article class="bg-zinc-950 p-5">
				<p class="text-sm font-semibold text-zinc-300">{service}</p>
				<p class="mt-3 text-xs uppercase tracking-[0.2em] text-zinc-600">Service</p>
			</article>
		{/each}
	</section>
</main>