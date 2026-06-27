<script lang="ts">
  import { MessageSquare, Radio, Server, Users } from "lucide-svelte";

  const statusCards = [
    {
      label: "Rooms",
      value: "0",
      helper: "Room list placeholder",
      icon: MessageSquare,
    },
    {
      label: "Presence",
      value: "Ready",
      helper: "Realtime status shell",
      icon: Radio,
    },
    {
      label: "Members",
      value: "1",
      helper: "Current authenticated user",
      icon: Users,
    },
  ];
</script>

<svelte:head>
  <title>Pylon Workspace</title>
  <meta
    name="description"
    content="Authenticated Pylon chat workspace shell with sidebar and room placeholder."
  />
</svelte:head>

<section class="px-5 py-6">
  <div class="border border-zinc-800 bg-zinc-950 p-6 md:p-8">
    <p class="text-xs font-black uppercase tracking-[0.28em] text-lime-300">
      Main app
    </p>

    <div class="mt-5 grid gap-6 lg:grid-cols-[1.2fr_0.8fr]">
      <div>
        <h2
          class="max-w-3xl text-5xl font-black leading-[0.95] tracking-[-0.06em] text-white md:text-7xl"
        >
          Welcome to the room deck.
        </h2>
        <p class="mt-6 max-w-xl text-base leading-7 text-zinc-400">
          This is the protected Pylon workspace. Room creation, room list, and
          realtime message panels will attach here in the next sprint.
        </p>
      </div>

      <div
        class="border border-lime-300 bg-lime-300 p-5 text-zinc-950 lg:self-end"
      >
        <Server class="size-7" aria-hidden="true" />
        <p class="mt-6 text-4xl font-black tracking-[-0.05em]">
          Auth guard active
        </p>
        <p class="mt-3 text-sm font-bold leading-6">
          Unauthenticated users are routed back to login before seeing the app
          shell.
        </p>
      </div>
    </div>
  </div>

  <div
    class="mt-6 grid gap-px border border-zinc-800 bg-zinc-800 md:grid-cols-3"
  >
    {#each statusCards as card}
      {@const Icon = card.icon}

      <article class="bg-zinc-950 p-5">
        <Icon class="size-6 text-lime-300" aria-hidden="true" />
        <p
          class="mt-5 text-xs font-black uppercase tracking-[0.22em] text-zinc-500"
        >
          {card.label}
        </p>
        <p class="mt-2 text-2xl font-black text-white">{card.value}</p>
        <p class="mt-2 text-sm text-zinc-400">{card.helper}</p>
      </article>
    {/each}
  </div>

  <section class="mt-6 border border-dashed border-zinc-800 p-8 text-center">
    <MessageSquare class="mx-auto size-8 text-zinc-600" aria-hidden="true" />
    <h3 class="mt-4 text-2xl font-black tracking-[-0.04em] text-white">
      No room selected
    </h3>
    <p class="mx-auto mt-3 max-w-md text-sm leading-6 text-zinc-500">
      The chat area is reserved for the upcoming room list and message stream.
    </p>
  </section>
</section>
