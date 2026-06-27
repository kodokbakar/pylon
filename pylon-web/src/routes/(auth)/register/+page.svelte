<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { AlertCircle, ArrowRight, Lock, Mail, User } from "lucide-svelte";

  import {
    hasRegisterFormErrors,
    validateRegisterForm,
    type RegisterFormErrors,
  } from "$lib/auth/register-validation";
  import { auth } from "$lib/stores/auth.svelte";

  let formData = $state({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  let errors = $state<RegisterFormErrors>({});
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
    errors = validateRegisterForm(formData);

    if (hasRegisterFormErrors(errors)) {
      return;
    }

    isSubmitting = true;

    try {
      await auth.register({
        username: formData.username.trim(),
        email: formData.email.trim(),
        password: formData.password,
      });

      await goto("/login");
    } catch (error) {
      apiError =
        error instanceof Error
          ? error.message
          : "Unable to create your Pylon account.";
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
  <title>Create account · Pylon</title>
  <meta
    name="description"
    content="Create a Pylon account for the distributed real-time chat control surface."
  />
</svelte:head>

<main class="min-h-screen bg-zinc-950 px-5 py-8 text-zinc-100 sm:px-8">
  <section
    class="mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl items-center gap-8 lg:grid-cols-[0.92fr_1.08fr]"
  >
    <aside class="hidden border border-zinc-800 bg-zinc-950 p-8 lg:block">
      <p class="text-sm font-black uppercase tracking-[0.34em] text-lime-300">
        Pylon Web
      </p>

      <h1
        class="mt-8 max-w-xl text-6xl font-black leading-[0.92] tracking-[-0.06em] text-white"
      >
        Claim your chat node.
      </h1>

      <p class="mt-6 max-w-md text-base leading-7 text-zinc-400">
        Register once, then plug into Pylon rooms, messages, and realtime
        presence from the same control surface.
      </p>

      <div
        class="mt-10 grid grid-cols-2 gap-px border border-zinc-800 bg-zinc-800"
      >
        <div class="bg-zinc-950 p-5">
          <p class="text-3xl font-black text-white">JWT</p>
          <p class="mt-2 text-xs uppercase tracking-[0.2em] text-zinc-500">
            Auth layer
          </p>
        </div>
        <div class="bg-lime-300 p-5 text-zinc-950">
          <p class="text-3xl font-black">Live</p>
          <p class="mt-2 text-xs font-bold uppercase tracking-[0.2em]">
            Gateway ready
          </p>
        </div>
      </div>
    </aside>

    <section
      class="mx-auto w-full max-w-xl border border-zinc-800 bg-zinc-950 p-5 shadow-[12px_12px_0_0_#bef264] sm:p-8"
    >
      <div class="mb-8 border-b border-zinc-800 pb-6">
        <p class="text-xs font-black uppercase tracking-[0.28em] text-lime-300">
          Register
        </p>
        <h2 class="mt-3 text-4xl font-black tracking-[-0.04em] text-white">
          Create account
        </h2>
        <p class="mt-3 text-sm leading-6 text-zinc-400">
          Start with a username, email, and strong password.
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
        <label class="grid gap-2" for="username">
          <span class="text-sm font-bold text-zinc-200">Username</span>
          <span class="relative">
            <User
              class="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-zinc-500"
              aria-hidden="true"
            />
            <input
              id="username"
              class="w-full border border-zinc-700 bg-zinc-900 px-11 py-3 text-base text-white outline-none transition focus:border-lime-300"
              class:border-red-400={errors.username}
              name="username"
              type="text"
              autocomplete="username"
              placeholder="alice123"
              bind:value={formData.username}
              aria-invalid={errors.username ? "true" : "false"}
              aria-describedby={errors.username ? "username-error" : undefined}
              oninput={() => clearFieldError("username")}
            />
          </span>
          {#if errors.username}
            <span
              id="username-error"
              class="text-sm font-semibold text-red-300"
            >
              {errors.username}
            </span>
          {/if}
        </label>

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
              autocomplete="new-password"
              placeholder="Minimum 8 characters"
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

        <label class="grid gap-2" for="confirm-password">
          <span class="text-sm font-bold text-zinc-200">Confirm password</span>
          <span class="relative">
            <Lock
              class="pointer-events-none absolute left-4 top-1/2 size-4 -translate-y-1/2 text-zinc-500"
              aria-hidden="true"
            />
            <input
              id="confirm-password"
              class="w-full border border-zinc-700 bg-zinc-900 px-11 py-3 text-base text-white outline-none transition focus:border-lime-300"
              class:border-red-400={errors.confirmPassword}
              name="confirmPassword"
              type="password"
              autocomplete="new-password"
              placeholder="Repeat password"
              bind:value={formData.confirmPassword}
              aria-invalid={errors.confirmPassword ? "true" : "false"}
              aria-describedby={errors.confirmPassword
                ? "confirm-password-error"
                : undefined}
              oninput={() => clearFieldError("confirmPassword")}
            />
          </span>
          {#if errors.confirmPassword}
            <span
              id="confirm-password-error"
              class="text-sm font-semibold text-red-300"
            >
              {errors.confirmPassword}
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
            Creating
          {:else}
            Create account
            <ArrowRight class="size-4" aria-hidden="true" />
          {/if}
        </button>
      </form>

      <p class="mt-7 text-center text-sm text-zinc-400">
        Already have an account?
        <a
          class="font-bold text-lime-300 underline decoration-zinc-700 underline-offset-4 hover:text-lime-200"
          href="/login"
        >
          Sign in
        </a>
      </p>
    </section>
  </section>
</main>
