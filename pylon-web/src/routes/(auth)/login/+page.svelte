<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { AlertCircle, ArrowRight, Lock, Mail } from "lucide-svelte";

  import {
    hasLoginFormErrors,
    validateLoginForm,
    type LoginFormErrors,
  } from "$lib/auth/login-validation";
  import { auth } from "$lib/stores/auth.svelte";

  let formData = $state({
    email: "",
    password: "",
  });

  let errors = $state<LoginFormErrors>({});
  let apiError = $state<string | null>(null);
  let isSubmitting = $state(false);

  onMount(() => {
    if (auth.isAuthenticated) {
      void goto("/");
    }
  });

  async function handleSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    apiError = null;
    errors = validateLoginForm(formData);

    if (hasLoginFormErrors(errors)) {
      return;
    }

    isSubmitting = true;

    try {
      await auth.login({
        email: formData.email.trim(),
        password: formData.password,
      });

      await goto("/");
    } catch (error) {
      apiError =
        error instanceof Error
          ? error.message
          : "Unable to sign in to your Pylon account.";
    } finally {
      isSubmitting = false;
    }
  }

  function clearFieldError(field: keyof typeof formData): void {
    if (errors[field] === undefined) {
      return;
    }

    errors = {
      ...errors,
      [field]: undefined,
    };
  }
</script>

<svelte:head>
  <title>Sign in · Pylon</title>
  <meta
    name="description"
    content="Sign in to the Pylon distributed real-time chat control surface."
  />
</svelte:head>

<main class="min-h-screen bg-zinc-950 px-5 py-8 text-zinc-100 sm:px-8">
  <section
    class="mx-auto flex min-h-[calc(100vh-4rem)] max-w-5xl items-center justify-center"
  >
    <div
      class="w-full max-w-md border border-zinc-800 bg-zinc-950 p-5 shadow-[12px_12px_0_0_#bef264] sm:p-8"
    >
      <div class="mb-8 border-b border-zinc-800 pb-6">
        <p class="text-xs font-black uppercase tracking-[0.28em] text-lime-300">
          Login
        </p>
        <h1 class="mt-3 text-4xl font-black tracking-[-0.04em] text-white">
          Back to the grid
        </h1>
        <p class="mt-3 text-sm leading-6 text-zinc-400">
          Sign in with your Pylon credentials and continue to the realtime chat
          control surface.
        </p>
      </div>

      {#if apiError}
        <div
          class="mb-5 flex gap-3 border border-red-500/70 bg-red-950/40 p-4 text-sm text-red-100"
          role="alert"
        >
          <AlertCircle class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
          <p>{apiError}</p>
        </div>
      {/if}

      <form class="grid gap-5" novalidate onsubmit={handleSubmit}>
        <label class="grid gap-2" for="email">
          <span class="text-sm font-bold text-zinc-200">Email</span>
          <span class="relative">
            <Mail
              class="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-zinc-500"
              aria-hidden="true"
            />
            <input
              id="email"
              class="w-full border border-zinc-700 bg-zinc-900 px-11 py-3 text-base text-white outline-none transition focus:border-lime-300"
              class:border-red-400={errors.email}
              name="email"
              type="email"
              autocomplete="email"
              placeholder="alice@example.com"
              bind:value={formData.email}
              aria-invalid={errors.email ? "true" : "false"}
              aria-describedby={errors.email ? "email-error" : undefined}
              oninput={() => clearFieldError("email")}
            />
          </span>
          {#if errors.email}
            <span id="email-error" class="text-sm font-semibold text-red-300">
              {errors.email}
            </span>
          {/if}
        </label>

        <label class="grid gap-2" for="password">
          <span class="text-sm font-bold text-zinc-200">Password</span>
          <span class="relative">
            <Lock
              class="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-zinc-500"
              aria-hidden="true"
            />
            <input
              id="password"
              class="w-full border border-zinc-700 bg-zinc-900 px-11 py-3 text-base text-white outline-none transition focus:border-lime-300"
              class:border-red-400={errors.password}
              name="password"
              type="password"
              autocomplete="current-password"
              placeholder="Your password"
              bind:value={formData.password}
              aria-invalid={errors.password ? "true" : "false"}
              aria-describedby={errors.password ? "password-error" : undefined}
              oninput={() => clearFieldError("password")}
            />
          </span>
          {#if errors.password}
            <span
              id="password-error"
              class="text-sm font-semibold text-red-300"
            >
              {errors.password}
            </span>
          {/if}
        </label>

        <button
          class="mt-2 inline-flex items-center justify-center gap-3 bg-lime-300 px-5 py-3 text-sm font-black uppercase tracking-[0.18em] text-zinc-950 transition hover:-translate-y-0.5 hover:shadow-[6px_6px_0_0_#3f3f46] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0 disabled:hover:shadow-none"
          type="submit"
          disabled={isSubmitting}
        >
          {#if isSubmitting}
            <span
              class="size-4 animate-spin rounded-full border-2 border-zinc-950 border-t-transparent"
              aria-hidden="true"
            ></span>
            Signing in
          {:else}
            Sign in
            <ArrowRight class="size-4" aria-hidden="true" />
          {/if}
        </button>
      </form>

      <p class="mt-7 text-center text-sm text-zinc-400">
        Need an account?
        <a
          class="font-bold text-lime-300 underline decoration-zinc-700 underline-offset-4 hover:text-lime-200"
          href="/register"
        >
          Create one
        </a>
      </p>
    </div>
  </section>
</main>
