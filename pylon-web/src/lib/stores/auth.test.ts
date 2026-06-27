import { describe, expect, it, vi } from "vitest";

import type {
  AuthResult,
  AuthUser,
  LoginInput,
  RegisterInput,
  RefreshResult,
} from "$lib/api/auth";
import {
  AuthStore,
  authStorageKey,
  isTokenExpired,
  type AuthApi,
  type AuthStorage,
} from "./auth.svelte";

const user: AuthUser = {
  id: "user-1",
  username: "alice",
  email: "alice@example.com",
  display_name: "Alice",
  avatar_url: "",
  created_at: "2026-06-27T00:00:00Z",
};

describe("AuthStore", () => {
  it("persists auth state and restores it on init", async () => {
    const storage = new MemoryStorage();
    const result = authResult({
      token: tokenWithExp(secondsFromNow(3600)),
      refreshToken: tokenWithExp(secondsFromNow(86400)),
    });

    const store = new AuthStore({
      api: authApi({ login: vi.fn<loginFn>().mockResolvedValue(result) }),
      storage,
      autoInitialize: false,
    });

    await store.login({
      email: "alice@example.com",
      password: "password123",
    });

    const persisted = storage.getItem(authStorageKey);

    expect(persisted).not.toBeNull();
    expect(persisted).toContain(result.token);
    expect(store.isAuthenticated).toBe(true);

    const restored = new AuthStore({
      api: authApi(),
      storage,
      autoInitialize: false,
    });

    await restored.init();

    expect(restored.user).toEqual(user);
    expect(restored.token).toBe(result.token);
    expect(restored.refreshToken).toBe(result.refresh_token);
    expect(restored.isAuthenticated).toBe(true);
    expect(restored.isLoading).toBe(false);
  });

  it("returns false for isAuthenticated when the token is expired", async () => {
    const store = new AuthStore({
      api: authApi({
        login: vi.fn<loginFn>().mockResolvedValue(
          authResult({
            token: tokenWithExp(secondsFromNow(-60)),
            refreshToken: tokenWithExp(secondsFromNow(86400)),
            expiresAt: isoFromNow(-60),
          }),
        ),
      }),
      storage: new MemoryStorage(),
      autoInitialize: false,
    });

    await store.login({
      email: "alice@example.com",
      password: "password123",
    });

    expect(store.isAuthenticated).toBe(false);
    expect(isTokenExpired(store.token, store.expiresAt)).toBe(true);
  });

  it("refreshes an expired access token on init", async () => {
    const storage = new MemoryStorage();
    const expiredToken = tokenWithExp(secondsFromNow(-60));
    const newToken = tokenWithExp(secondsFromNow(3600));
    const newRefreshToken = tokenWithExp(secondsFromNow(86400));

    storage.setItem(
      authStorageKey,
      JSON.stringify({
        user,
        token: expiredToken,
        refreshToken: tokenWithExp(secondsFromNow(86400)),
        expiresAt: isoFromNow(-60),
        refreshExpiresAt: isoFromNow(86400),
      }),
    );

    const refreshToken = vi.fn<refreshTokenFn>().mockResolvedValue(
      refreshResult({
        token: newToken,
        refreshToken: newRefreshToken,
      }),
    );

    const store = new AuthStore({
      api: authApi({ refreshToken }),
      storage,
      autoInitialize: false,
    });

    await store.init();

    expect(refreshToken).toHaveBeenCalledTimes(1);
    expect(store.user).toEqual(user);
    expect(store.token).toBe(newToken);
    expect(store.refreshToken).toBe(newRefreshToken);
    expect(store.isAuthenticated).toBe(true);
    expect(storage.getItem(authStorageKey)).toContain(newToken);
  });

  it("logout clears auth state and persisted storage", async () => {
    const storage = new MemoryStorage();

    const store = new AuthStore({
      api: authApi({
        login: vi.fn<loginFn>().mockResolvedValue(
          authResult({
            token: tokenWithExp(secondsFromNow(3600)),
            refreshToken: tokenWithExp(secondsFromNow(86400)),
          }),
        ),
      }),
      storage,
      autoInitialize: false,
    });

    await store.login({
      email: "alice@example.com",
      password: "password123",
    });

    store.logout();

    expect(store.user).toBeNull();
    expect(store.token).toBeNull();
    expect(store.refreshToken).toBeNull();
    expect(store.expiresAt).toBeNull();
    expect(store.refreshExpiresAt).toBeNull();
    expect(store.isAuthenticated).toBe(false);
    expect(storage.getItem(authStorageKey)).toBeNull();
  });

  it("initializes without browser storage for SSR safety", async () => {
    const store = new AuthStore({
      api: authApi(),
      storage: null,
      autoInitialize: false,
    });

    await store.init();

    expect(store.user).toBeNull();
    expect(store.token).toBeNull();
    expect(store.isAuthenticated).toBe(false);
    expect(store.isLoading).toBe(false);
  });
});

type registerFn = (input: RegisterInput) => Promise<AuthResult>;
type loginFn = (input: LoginInput) => Promise<AuthResult>;
type refreshTokenFn = (refreshToken: string) => Promise<RefreshResult>;

class MemoryStorage implements AuthStorage {
  private readonly items = new Map<string, string>();

  getItem(key: string): string | null {
    return this.items.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.items.set(key, value);
  }

  removeItem(key: string): void {
    this.items.delete(key);
  }
}

function authApi(overrides: Partial<AuthApi> = {}): AuthApi {
  return {
    register: vi.fn<registerFn>().mockResolvedValue(
      authResult({
        token: tokenWithExp(secondsFromNow(3600)),
        refreshToken: tokenWithExp(secondsFromNow(86400)),
      }),
    ),
    login: vi.fn<loginFn>().mockResolvedValue(
      authResult({
        token: tokenWithExp(secondsFromNow(3600)),
        refreshToken: tokenWithExp(secondsFromNow(86400)),
      }),
    ),
    refreshToken: vi.fn<refreshTokenFn>().mockResolvedValue(
      refreshResult({
        token: tokenWithExp(secondsFromNow(3600)),
        refreshToken: tokenWithExp(secondsFromNow(86400)),
      }),
    ),
    ...overrides,
  };
}

function authResult(input: {
  token: string;
  refreshToken: string;
  expiresAt?: string;
  refreshExpiresAt?: string;
}): AuthResult {
  return {
    user,
    token: input.token,
    refresh_token: input.refreshToken,
    expires_at: input.expiresAt ?? isoFromNow(3600),
    refresh_expires_at: input.refreshExpiresAt ?? isoFromNow(86400),
  };
}

function refreshResult(input: {
  token: string;
  refreshToken: string;
  expiresAt?: string;
  refreshExpiresAt?: string;
}): RefreshResult {
  return {
    token: input.token,
    refresh_token: input.refreshToken,
    expires_at: input.expiresAt ?? isoFromNow(3600),
    refresh_expires_at: input.refreshExpiresAt ?? isoFromNow(86400),
  };
}

function tokenWithExp(exp: number): string {
  const payload = base64UrlEncode(JSON.stringify({ exp }));

  return `header.${payload}.signature`;
}

function base64UrlEncode(value: string): string {
  return btoa(value)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/g, "");
}

function secondsFromNow(offsetSeconds: number): number {
  return Math.floor(Date.now() / 1000) + offsetSeconds;
}

function isoFromNow(offsetSeconds: number): string {
  return new Date(Date.now() + offsetSeconds * 1000).toISOString();
}
