<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { onMount } from "svelte";
  import {
    AlertCircle,
    Hash,
    LoaderCircle,
    MessageSquare,
    Plus,
    Radio,
    RefreshCcw,
    Server,
    Users,
    X,
  } from "lucide-svelte";

  import { createRoom, listRooms, type Room } from "$lib/api/rooms";
  import {
    hasCreateRoomFormErrors,
    validateCreateRoomForm,
    type CreateRoomFormErrors,
  } from "$lib/rooms/room-validation";
  import { auth } from "$lib/stores/auth.svelte";

  let rooms = $state<Room[]>([]);
  let isLoadingRooms = $state(true);
  let roomError = $state<string | null>(null);

  let isCreateModalOpen = $state(false);
  let isCreatingRoom = $state(false);
  let createError = $state<string | null>(null);
  let createForm = $state({
    name: "",
    description: "",
  });
  let createErrors = $state<CreateRoomFormErrors>({});

  let roomCount = $derived(rooms.length);
  let statusCards = $derived([
    {
      label: "Rooms",
      value: String(roomCount),
      helper: roomCount === 1 ? "Joined room" : "Joined rooms",
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
  ]);

  onMount(() => {
    void loadRooms();
  });

  $effect(() => {
    if (page.url.searchParams.get("create") === "room" && !isCreateModalOpen) {
      openCreateModal();
    }
  });

  async function loadRooms(): Promise<void> {
    if (!auth.token) {
      isLoadingRooms = false;
      return;
    }

    isLoadingRooms = true;
    roomError = null;

    try {
      rooms = await listRooms(auth.token);
    } catch (error) {
      roomError =
        error instanceof Error ? error.message : "Unable to load rooms.";
    } finally {
      isLoadingRooms = false;
    }
  }

  function openCreateModal(): void {
    createForm = {
      name: "",
      description: "",
    };
    createErrors = {};
    createError = null;
    isCreateModalOpen = true;
  }

  async function closeCreateModal(): Promise<void> {
    isCreateModalOpen = false;
    createError = null;
    createErrors = {};

    if (page.url.searchParams.get("create") === "room") {
      await goto("/", {
        replaceState: true,
        noScroll: true,
      });
    }
  }

  async function handleCreateRoom(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    createError = null;
    createErrors = validateCreateRoomForm(createForm);

    if (hasCreateRoomFormErrors(createErrors)) {
      return;
    }

    if (!auth.token) {
      createError = "Authentication token is missing. Please sign in again.";
      return;
    }

    isCreatingRoom = true;

    try {
      const room = await createRoom(
        {
          name: createForm.name,
          description: createForm.description,
          type: "group",
        },
        auth.token,
      );

      const description = createForm.description.trim();
      const nextRoom = description === "" ? room : { ...room, description };

      rooms = [nextRoom, ...rooms.filter((item) => item.id !== room.id)];
      await closeCreateModal();
    } catch (error) {
      createError =
        error instanceof Error ? error.message : "Unable to create room.";
    } finally {
      isCreatingRoom = false;
    }
  }

  function clearCreateFieldError(field: keyof typeof createForm): void {
    if (createErrors[field] === undefined) {
      return;
    }

    createErrors = {
      ...createErrors,
      [field]: undefined,
    };
  }

  function formatRoomType(type: Room["type"]): string {
    switch (type) {
      case "direct":
        return "Direct";
      case "group":
        return "Group";
      case "channel":
        return "Channel";
      default:
        return "Room";
    }
  }

  function formatCreatedAt(value: string): string {
    if (value.trim() === "") {
      return "Unknown time";
    }

    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return "Unknown time";
    }

    return new Intl.DateTimeFormat("en", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  }
</script>

<svelte:head>
  <title>Pylon Rooms</title>
  <meta
    name="description"
    content="Authenticated Pylon room list with create room modal."
  />
</svelte:head>

<section class="px-5 py-6">
  <div class="border border-zinc-800 bg-zinc-950 p-6 md:p-8">
    <div
      class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between"
    >
      <div>
        <p class="text-xs font-black uppercase tracking-[0.28em] text-lime-300">
          Room deck
        </p>

        <h2
          class="mt-4 max-w-3xl text-5xl font-black leading-[0.95] tracking-[-0.06em] text-white md:text-7xl"
        >
          Rooms you can join now.
        </h2>

        <p class="mt-6 max-w-xl text-base leading-7 text-zinc-400">
          Load your joined rooms from the API Gateway, create a new room, then
          attach the message stream in the next issue.
        </p>
      </div>

      <button
        class="inline-flex items-center justify-center gap-3 bg-lime-300 px-5 py-3 text-sm font-black uppercase tracking-[0.18em] text-zinc-950 transition hover:-translate-y-0.5 hover:shadow-[6px_6px_0_0_#3f3f46]"
        type="button"
        onclick={openCreateModal}
      >
        <Plus class="size-4" aria-hidden="true" />
        Create room
      </button>
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

  <section class="mt-6 border border-zinc-800">
    <div
      class="flex flex-col gap-4 border-b border-zinc-800 p-5 sm:flex-row sm:items-center sm:justify-between"
    >
      <div>
        <p class="text-xs font-black uppercase tracking-[0.24em] text-zinc-500">
          Joined rooms
        </p>
        <h3 class="mt-1 text-2xl font-black tracking-[-0.04em] text-white">
          Room list
        </h3>
      </div>

      <button
        class="inline-flex items-center justify-center gap-2 border border-zinc-800 px-3 py-2 text-sm font-bold text-zinc-300 transition hover:border-lime-300 hover:text-lime-300"
        type="button"
        onclick={loadRooms}
        disabled={isLoadingRooms}
      >
        <RefreshCcw
          class={isLoadingRooms ? "size-4 animate-spin" : "size-4"}
          aria-hidden="true"
        />
        Refresh
      </button>
    </div>

    {#if roomError}
      <div
        class="m-5 flex gap-3 border border-red-500/70 bg-red-950/40 p-4 text-sm text-red-100"
        role="alert"
      >
        <AlertCircle class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
        <p>{roomError}</p>
      </div>
    {/if}

    {#if isLoadingRooms}
      <div class="grid min-h-64 place-items-center p-8">
        <div class="flex items-center gap-3 text-zinc-400">
          <LoaderCircle
            class="size-5 animate-spin text-lime-300"
            aria-hidden="true"
          />
          <p class="text-sm font-bold">Loading rooms</p>
        </div>
      </div>
    {:else if rooms.length === 0}
      <div class="grid min-h-64 place-items-center p-8 text-center">
        <div>
          <MessageSquare
            class="mx-auto size-9 text-zinc-600"
            aria-hidden="true"
          />
          <h3 class="mt-4 text-2xl font-black tracking-[-0.04em] text-white">
            No rooms yet
          </h3>
          <p class="mx-auto mt-3 max-w-md text-sm leading-6 text-zinc-500">
            Create your first Pylon room to start shaping the workspace.
          </p>
          <button
            class="mt-6 inline-flex items-center justify-center gap-3 bg-lime-300 px-5 py-3 text-sm font-black uppercase tracking-[0.18em] text-zinc-950 transition hover:-translate-y-0.5 hover:shadow-[6px_6px_0_0_#3f3f46]"
            type="button"
            onclick={openCreateModal}
          >
            <Plus class="size-4" aria-hidden="true" />
            Create room
          </button>
        </div>
      </div>
    {:else}
      <div class="grid gap-px bg-zinc-800 md:grid-cols-2 xl:grid-cols-3">
        {#each rooms as room}
          <article class="bg-zinc-950 p-5">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <p
                  class="truncate text-xl font-black tracking-[-0.04em] text-white"
                >
                  {room.name}
                </p>
                <p
                  class="mt-2 text-xs font-black uppercase tracking-[0.2em] text-lime-300"
                >
                  {formatRoomType(room.type)}
                </p>
              </div>

              <div
                class="grid size-10 shrink-0 place-items-center border border-zinc-800 bg-zinc-900"
              >
                <Hash class="size-5 text-zinc-500" aria-hidden="true" />
              </div>
            </div>

            {#if room.description}
              <p class="mt-4 text-sm leading-6 text-zinc-400">
                {room.description}
              </p>
            {:else}
              <p class="mt-4 text-sm leading-6 text-zinc-600">
                No description provided.
              </p>
            {/if}

            <div class="mt-5 border-t border-zinc-800 pt-4">
              <p class="text-xs text-zinc-500">
                Created {formatCreatedAt(room.created_at)}
              </p>
            </div>
          </article>
        {/each}
      </div>
    {/if}
  </section>
</section>

{#if isCreateModalOpen}
  <div
    class="fixed inset-0 z-50 grid place-items-center bg-black/75 px-5 py-8 backdrop-blur-sm"
    role="presentation"
  >
    <div
      class="w-full max-w-lg border border-zinc-800 bg-zinc-950 p-5 shadow-[12px_12px_0_0_#bef264] sm:p-7"
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-room-title"
    >
      <div
        class="mb-6 flex items-start justify-between gap-4 border-b border-zinc-800 pb-5"
      >
        <div>
          <p
            class="text-xs font-black uppercase tracking-[0.28em] text-lime-300"
          >
            New room
          </p>
          <h3
            id="create-room-title"
            class="mt-2 text-3xl font-black tracking-[-0.04em] text-white"
          >
            Create room
          </h3>
        </div>

        <button
          class="grid size-9 place-items-center border border-zinc-800 text-zinc-400 transition hover:border-red-400 hover:text-red-300"
          type="button"
          aria-label="Close create room modal"
          onclick={closeCreateModal}
          disabled={isCreatingRoom}
        >
          <X class="size-4" aria-hidden="true" />
        </button>
      </div>

      {#if createError}
        <div
          class="mb-5 flex gap-3 border border-red-500/70 bg-red-950/40 p-4 text-sm text-red-100"
          role="alert"
        >
          <AlertCircle class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
          <p>{createError}</p>
        </div>
      {/if}

      <form class="grid gap-5" novalidate onsubmit={handleCreateRoom}>
        <label class="grid gap-2" for="room-name">
          <span class="text-sm font-bold text-zinc-200">Room name</span>
          <input
            id="room-name"
            class="w-full border border-zinc-700 bg-zinc-900 px-4 py-3 text-base text-white outline-none transition focus:border-lime-300"
            class:border-red-400={createErrors.name}
            name="name"
            type="text"
            autocomplete="off"
            placeholder="Backend Guild"
            bind:value={createForm.name}
            aria-invalid={createErrors.name ? "true" : "false"}
            aria-describedby={createErrors.name ? "room-name-error" : undefined}
            oninput={() => clearCreateFieldError("name")}
          />
          {#if createErrors.name}
            <span
              id="room-name-error"
              class="text-sm font-semibold text-red-300"
            >
              {createErrors.name}
            </span>
          {/if}
        </label>

        <label class="grid gap-2" for="room-description">
          <span class="text-sm font-bold text-zinc-200">Description</span>
          <textarea
            id="room-description"
            class="min-h-28 w-full resize-y border border-zinc-700 bg-zinc-900 px-4 py-3 text-base text-white outline-none transition focus:border-lime-300"
            class:border-red-400={createErrors.description}
            name="description"
            placeholder="What is this room for?"
            bind:value={createForm.description}
            aria-invalid={createErrors.description ? "true" : "false"}
            aria-describedby={createErrors.description
              ? "room-description-error"
              : "room-description-help"}
            oninput={() => clearCreateFieldError("description")}
          ></textarea>
          {#if createErrors.description}
            <span
              id="room-description-error"
              class="text-sm font-semibold text-red-300"
            >
              {createErrors.description}
            </span>
          {:else}
            <span id="room-description-help" class="text-xs text-zinc-500">
              Backend currently persists room name and type; description is kept
              in the current UI state after create.
            </span>
          {/if}
        </label>

        <div class="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <button
            class="border border-zinc-800 px-5 py-3 text-sm font-black uppercase tracking-[0.18em] text-zinc-300 transition hover:border-red-400 hover:text-red-300 disabled:cursor-not-allowed disabled:opacity-60"
            type="button"
            onclick={closeCreateModal}
            disabled={isCreatingRoom}
          >
            Cancel
          </button>

          <button
            class="inline-flex items-center justify-center gap-3 bg-lime-300 px-5 py-3 text-sm font-black uppercase tracking-[0.18em] text-zinc-950 transition hover:-translate-y-0.5 hover:shadow-[6px_6px_0_0_#3f3f46] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0 disabled:hover:shadow-none"
            type="submit"
            disabled={isCreatingRoom}
          >
            {#if isCreatingRoom}
              <span
                class="size-4 animate-spin rounded-full border-2 border-zinc-950 border-t-transparent"
                aria-hidden="true"
              ></span>
              Creating
            {:else}
              Create room
            {/if}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
