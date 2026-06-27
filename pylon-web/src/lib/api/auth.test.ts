import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError, login, register } from "./auth";

const authPayload = {
  user: {
    id: "user-1",
    username: "alice",
    email: "alice@example.com",
    display_name: "Alice",
    avatar_url: "",
    created_at: "2026-06-27T00:00:00Z",
  },
  token: "access-token",
  refresh_token: "refresh-token",
  expires_at: "2026-06-27T00:15:00Z",
  refresh_expires_at: "2026-07-04T00:00:00Z",
};

describe("auth API", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("register posts JSON to the REST register endpoint", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse({ data: authPayload }, 201));
    vi.stubGlobal("fetch", fetchMock);

    const result = await register({
      username: "alice",
      email: "alice@example.com",
      password: "password123",
    });

    expect(result).toEqual(authPayload);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/register",
      expect.objectContaining({
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({
          username: "alice",
          email: "alice@example.com",
          password: "password123",
        }),
      }),
    );
  });

  it("login posts JSON to the REST login endpoint", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse({ data: authPayload }));
    vi.stubGlobal("fetch", fetchMock);

    const result = await login({
      email: "alice@example.com",
      password: "password123",
    });

    expect(result.token).toBe("access-token");
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/auth/login",
      expect.objectContaining({
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({
          email: "alice@example.com",
          password: "password123",
        }),
      }),
    );
  });

  it("wraps REST error responses as ApiError", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse(
        {
          success: false,
          error: {
            code: "unauthorized",
            message: "invalid credentials",
          },
        },
        401,
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    await expect(
      login({
        email: "alice@example.com",
        password: "wrong-password",
      }),
    ).rejects.toMatchObject({
      name: "ApiError",
      code: "unauthorized",
      status: 401,
      message: "invalid credentials",
    });
  });

  it("wraps network failures as ApiError", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("connection refused")),
    );

    await expect(
      login({
        email: "alice@example.com",
        password: "password123",
      }),
    ).rejects.toBeInstanceOf(ApiError);
  });
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
