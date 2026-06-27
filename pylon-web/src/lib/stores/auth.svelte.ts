import {
  login as apiLogin,
  refreshToken as apiRefreshToken,
  register as apiRegister,
  type AuthResult,
  type AuthUser,
  type LoginInput,
  type RefreshResult,
  type RegisterInput,
} from "$lib/api/auth";

export const authStorageKey = "pylon-auth";

const tokenRefreshSkewMs = 30_000;

export type AuthStorage = Pick<Storage, "getItem" | "setItem" | "removeItem">;

export type AuthApi = {
  register(input: RegisterInput): Promise<AuthResult>;
  login(input: LoginInput): Promise<AuthResult>;
  refreshToken(refreshToken: string): Promise<RefreshResult>;
};

export type PersistedAuth = {
  user: AuthUser | null;
  token: string | null;
  refreshToken: string | null;
  expiresAt: string | null;
  refreshExpiresAt: string | null;
};

type AuthStoreOptions = {
  api?: AuthApi;
  storage?: AuthStorage | null;
  autoInitialize?: boolean;
};

const defaultAuthApi: AuthApi = {
  register: apiRegister,
  login: apiLogin,
  refreshToken: apiRefreshToken,
};

export class AuthStore {
  user = $state<AuthUser | null>(null);
  token = $state<string | null>(null);
  refreshToken = $state<string | null>(null);
  expiresAt = $state<string | null>(null);
  refreshExpiresAt = $state<string | null>(null);
  isLoading = $state(true);
  error = $state<string | null>(null);

  isAuthenticated = $derived(
    this.token !== null && !isTokenExpired(this.token, this.expiresAt),
  );

  private readonly api: AuthApi;
  private readonly storage: AuthStorage | null;
  private initialized = false;

  constructor(options: AuthStoreOptions = {}) {
    this.api = options.api ?? defaultAuthApi;
    this.storage =
      options.storage === undefined ? resolveBrowserStorage() : options.storage;

    if (options.autoInitialize === false) {
      this.isLoading = false;
      return;
    }

    void this.init();
  }

  async init(): Promise<void> {
    if (this.initialized) {
      return;
    }

    this.initialized = true;
    this.isLoading = true;
    this.error = null;

    try {
      this.loadFromStorage();

      if (
        this.token !== null &&
        this.refreshToken !== null &&
        isTokenExpired(this.token, this.expiresAt)
      ) {
        await this.refreshAccessToken();
      }
    } finally {
      this.isLoading = false;
    }
  }

  async register(input: RegisterInput): Promise<AuthResult> {
    this.isLoading = true;
    this.error = null;

    try {
      const result = await this.api.register(input);
      this.applyAuthResult(result);
      return result;
    } catch (error) {
      this.error = getErrorMessage(error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  async login(input: LoginInput): Promise<AuthResult> {
    this.isLoading = true;
    this.error = null;

    try {
      const result = await this.api.login(input);
      this.applyAuthResult(result);
      return result;
    } catch (error) {
      this.error = getErrorMessage(error);
      throw error;
    } finally {
      this.isLoading = false;
    }
  }

  async refreshAccessToken(): Promise<RefreshResult | null> {
    if (this.refreshToken === null) {
      this.logout();
      return null;
    }

    if (isTokenExpired(this.refreshToken, this.refreshExpiresAt)) {
      this.logout();
      return null;
    }

    try {
      const result = await this.api.refreshToken(this.refreshToken);
      this.applyRefreshResult(result);
      return result;
    } catch (error) {
      this.logout();
      this.error = getErrorMessage(error);
      throw error;
    }
  }

  logout(): void {
    this.user = null;
    this.token = null;
    this.refreshToken = null;
    this.expiresAt = null;
    this.refreshExpiresAt = null;
    this.error = null;
    this.isLoading = false;

    this.removeFromStorage();
  }

  private applyAuthResult(result: AuthResult): void {
    this.user = result.user;
    this.token = result.token;
    this.refreshToken = result.refresh_token;
    this.expiresAt = result.expires_at;
    this.refreshExpiresAt = result.refresh_expires_at;
    this.error = null;

    this.saveToStorage();
  }

  private applyRefreshResult(result: RefreshResult): void {
    this.token = result.token;
    this.refreshToken = result.refresh_token;
    this.expiresAt = result.expires_at;
    this.refreshExpiresAt = result.refresh_expires_at;
    this.error = null;

    this.saveToStorage();
  }

  private loadFromStorage(): void {
    if (this.storage === null) {
      return;
    }

    const rawValue = this.storage.getItem(authStorageKey);
    if (rawValue === null) {
      return;
    }

    const persisted = parsePersistedAuth(rawValue);
    if (persisted === null) {
      this.removeFromStorage();
      return;
    }

    this.user = persisted.user;
    this.token = persisted.token;
    this.refreshToken = persisted.refreshToken;
    this.expiresAt = persisted.expiresAt;
    this.refreshExpiresAt = persisted.refreshExpiresAt;
  }

  private saveToStorage(): void {
    if (this.storage === null) {
      return;
    }

    if (
      this.user === null &&
      this.token === null &&
      this.refreshToken === null
    ) {
      this.removeFromStorage();
      return;
    }

    const payload: PersistedAuth = {
      user: this.user,
      token: this.token,
      refreshToken: this.refreshToken,
      expiresAt: this.expiresAt,
      refreshExpiresAt: this.refreshExpiresAt,
    };

    try {
      this.storage.setItem(authStorageKey, JSON.stringify(payload));
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }

  private removeFromStorage(): void {
    if (this.storage === null) {
      return;
    }

    try {
      this.storage.removeItem(authStorageKey);
    } catch (error) {
      this.error = getErrorMessage(error);
    }
  }
}

export const auth = new AuthStore();

export function isTokenExpired(
  token: string | null,
  fallbackExpiresAt: string | null = null,
): boolean {
  const expiresAtMs = getTokenExpiresAtMs(token, fallbackExpiresAt);

  if (expiresAtMs === null) {
    return true;
  }

  return expiresAtMs <= Date.now();
}

export function shouldRefreshToken(
  token: string | null,
  fallbackExpiresAt: string | null = null,
): boolean {
  const expiresAtMs = getTokenExpiresAtMs(token, fallbackExpiresAt);

  if (expiresAtMs === null) {
    return true;
  }

  return expiresAtMs <= Date.now() + tokenRefreshSkewMs;
}

function getTokenExpiresAtMs(
  token: string | null,
  fallbackExpiresAt: string | null,
): number | null {
  const jwtExpiresAtMs = decodeJwtExpiresAtMs(token);
  if (jwtExpiresAtMs !== null) {
    return jwtExpiresAtMs;
  }

  if (fallbackExpiresAt === null || fallbackExpiresAt.trim() === "") {
    return null;
  }

  const parsedExpiresAt = Date.parse(fallbackExpiresAt);
  if (Number.isNaN(parsedExpiresAt)) {
    return null;
  }

  return parsedExpiresAt;
}

function decodeJwtExpiresAtMs(token: string | null): number | null {
  if (token === null) {
    return null;
  }

  const [, payload] = token.split(".");
  if (!payload) {
    return null;
  }

  try {
    const decodedPayload = JSON.parse(decodeBase64Url(payload)) as {
      exp?: unknown;
    };

    if (typeof decodedPayload.exp !== "number") {
      return null;
    }

    return decodedPayload.exp * 1000;
  } catch {
    return null;
  }
}

function decodeBase64Url(value: string): string {
  if (typeof atob !== "function") {
    throw new Error("base64 decoder is not available");
  }

  const normalized = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = normalized.padEnd(
    normalized.length + ((4 - (normalized.length % 4)) % 4),
    "=",
  );

  return atob(padded);
}

function parsePersistedAuth(rawValue: string): PersistedAuth | null {
  try {
    const parsed = JSON.parse(rawValue) as Partial<PersistedAuth>;

    return {
      user: isAuthUser(parsed.user) ? parsed.user : null,
      token: stringOrNull(parsed.token),
      refreshToken: stringOrNull(parsed.refreshToken),
      expiresAt: stringOrNull(parsed.expiresAt),
      refreshExpiresAt: stringOrNull(parsed.refreshExpiresAt),
    };
  } catch {
    return null;
  }
}

function isAuthUser(value: unknown): value is AuthUser {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const user = value as Partial<AuthUser>;

  return (
    typeof user.id === "string" &&
    typeof user.username === "string" &&
    typeof user.email === "string"
  );
}

function stringOrNull(value: unknown): string | null {
  if (typeof value !== "string") {
    return null;
  }

  const trimmed = value.trim();

  return trimmed === "" ? null : trimmed;
}

function resolveBrowserStorage(): AuthStorage | null {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage;
}

function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return "Unexpected auth error.";
}
