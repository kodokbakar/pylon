<script lang="ts">
  import { goto } from "$app/navigation";

  import { auth } from "$lib/stores/auth.svelte";

  let { children } = $props();

  $effect(() => {
    if (!auth.isLoading && auth.isAuthenticated) {
      void goto("/");
    }
  });
</script>

<div class="min-h-screen bg-zinc-950 text-zinc-100">
  {@render children()}
</div>
