<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import {
    Hash,
    LogOut,
    MessageSquare,
    Plus,
    Radio,
    ShieldCheck,
    UserCircle,
  } from "lucide-svelte";

  import { auth } from "$lib/stores/auth.svelte";

  let { children } = $props();

  let displayName = $derived(
    auth.user?.display_name || auth.user?.username || "Pylon User",
  );
  let username = $derived(auth.user?.username || "user");
  let userInitial = $derived(displayName.slice(0, 1).toUpperCase());
  let isHomeActive = $derived(page.url.pathname === "/");

  $effect(() => {
    if (!auth.isLoading && !auth.isAuthenticated) {
      void goto("/login");
    }
  });

  async function handleLogout(): Promise<void> {
    auth.logout();
    await goto("/login");
  }
</script>

{#if auth.isLoading}
  <main
    class="grid min-h-screen place-items-center bg-zinc-950 px-6 text-zinc-100"
  >
    <section
      class="border border-zinc-800 bg-zinc-950 p-6 shadow-[10px_10px_0_0_#bef264]"
    >
      <p class="text-xs font-black uppercase tracking-[0.28em] text-lime-300">
        Pylon Web
      </p>
      <div class="mt-5 flex items-center gap-3">
        <span
          class="size-4 animate-spin rounded-full border-2 border-lime-300 border-t-transparent"
          aria-hidden="true"
        ></span>
        <p class="text-sm font-bold text-zinc-300">Restoring session</p>
      </div>
    </section>
  </main>
{:else if auth.isAuthenticated}
  <div class="min-h-screen bg-zinc-950 text-zinc-100 lg:flex">
    <aside
      class="border-b border-zinc-800 bg-zinc-950 lg:sticky lg:top-0 lg:h-screen lg:w-72 lg:shrink-0 lg:overflow-y-auto lg:border-b-0 lg:border-r"
    >
      <div class="border-b border-zinc-800 p-5">
        <a class="group block" href="/" aria-label="Pylon home">
          <p
            class="text-xs font-black uppercase tracking-[0.32em] text-lime-300"
          >
            Pylon
          </p>
          <div class="mt-3 flex items-center gap-3">
            <div
              class="grid size-10 place-items-center border border-lime-300 bg-lime-300 text-zinc-950"
            >
              <Radio class="size-5" aria-hidden="true" />
            </div>
            <div>
              <p class="text-lg font-black tracking-[-0.04em] text-white">
                Web Console
              </p>
              <p class="text-xs uppercase tracking-[0.18em] text-zinc-500">
                Realtime chat
              </p>
            </div>
          </div>
        </a>
      </div>

      <nav class="p-4" aria-label="Main navigation">
        <a
          class="flex items-center gap-3 border px-3 py-3 text-sm font-bold transition hover:border-lime-300 hover:text-lime-300"
          class:border-lime-300={isHomeActive}
          class:bg-lime-300={isHomeActive}
          class:text-zinc-950={isHomeActive}
          class:border-zinc-800={!isHomeActive}
          class:text-zinc-300={!isHomeActive}
          href="/"
        >
          <Hash class="size-4" aria-hidden="true" />
          Home
        </a>
      </nav>

      <section class="border-t border-zinc-800 p-4">
        <div class="mb-4 flex items-center justify-between">
          <div>
            <p
              class="text-xs font-black uppercase tracking-[0.24em] text-zinc-500"
            >
              Rooms
            </p>
            <p class="mt-1 text-sm font-bold text-zinc-300">
              Channel directory
            </p>
          </div>

          <a
            class="grid size-9 place-items-center border border-zinc-800 text-zinc-400 transition hover:border-lime-300 hover:text-lime-300"
            href="/?create=room"
            aria-label="Create room"
          >
            <Plus class="size-4" aria-hidden="true" />
          </a>
        </div>

        <div class="border border-dashed border-zinc-800 p-5 text-center">
          <MessageSquare
            class="mx-auto size-7 text-zinc-600"
            aria-hidden="true"
          />
          <p class="mt-3 text-sm font-bold text-zinc-300">No rooms yet</p>
          <p class="mt-2 text-xs leading-5 text-zinc-500">
            Room list lands in the next sprint. This shell is ready for it.
          </p>
        </div>
      </section>
    </aside>

    <div class="min-w-0 flex-1">
      <header
        class="sticky top-0 z-10 flex min-h-16 items-center justify-between border-b border-zinc-800 bg-zinc-950/95 px-5 backdrop-blur"
      >
        <div>
          <p
            class="text-xs font-black uppercase tracking-[0.24em] text-lime-300"
          >
            Pylon App
          </p>
          <h1 class="text-lg font-black tracking-[-0.03em] text-white">
            Chat workspace
          </h1>
        </div>

        <div class="flex items-center gap-3">
          <div
            class="hidden items-center gap-3 border border-zinc-800 px-3 py-2 sm:flex"
          >
            <div
              class="grid size-8 place-items-center bg-zinc-800 text-sm font-black text-lime-300"
            >
              {userInitial}
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-bold text-white">{displayName}</p>
              <p class="truncate text-xs text-zinc-500">@{username}</p>
            </div>
          </div>

          <button
            class="inline-flex items-center gap-2 border border-zinc-800 px-3 py-2 text-sm font-bold text-zinc-300 transition hover:border-red-400 hover:text-red-300"
            type="button"
            onclick={handleLogout}
          >
            <LogOut class="size-4" aria-hidden="true" />
            <span class="hidden sm:inline">Logout</span>
          </button>
        </div>
      </header>

      <main class="min-h-[calc(100vh-4rem)] bg-zinc-950">
        {@render children()}
      </main>
    </div>
  </div>
{:else}
  <main
    class="grid min-h-screen place-items-center bg-zinc-950 px-6 text-zinc-100"
  >
    <section
      class="border border-zinc-800 bg-zinc-950 p-6 shadow-[10px_10px_0_0_#bef264]"
    >
      <ShieldCheck class="size-8 text-lime-300" aria-hidden="true" />
      <p class="mt-4 text-sm font-bold text-zinc-300">Redirecting to login</p>
    </section>
  </main>
{/if}
