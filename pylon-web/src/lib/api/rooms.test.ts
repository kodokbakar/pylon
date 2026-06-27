import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "$lib/api/auth";

import { createRoom, listRooms } from "./rooms";

const token = "access-token";

describe("rooms API", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("lists rooms with bearer auth", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      jsonResponse({
        data: {
          rooms: [
            {
              id: "room-1",
              name: "General",
              type: "ROOM_TYPE_CHANNEL",
              created_by: "user-1",
              created_at: "2026-06-27T12:00:00Z",
            },
          ],
        },
      }),
    );

    vi.stubGlobal("fetch", fetchMock);

    const rooms = await listRooms(token);

    expect(rooms).toEqual([
      {
        id: "room-1",
        name: "General",
        type: "channel",
        created_by: "user-1",
        created_at: "2026-06-27T12:00:00Z",
      },
    ]);
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/rooms",
      expect.objectContaining({
        method: "GET",
        headers: {
          Accept: "application/json",
          Authorization: "Bearer access-token",
        },
      }),
    );
  });

  it("creates a group room", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      jsonResponse(
        {
          data: {
            room: {
              id: "room-2",
              name: "Backend Guild",
              type: 2,
              created_by: "user-1",
              created_at: "2026-06-27T12:05:00Z",
            },
          },
        },
        201,
      ),
    );

    vi.stubGlobal("fetch", fetchMock);

    const room = await createRoom(
      {
        name: " Backend Guild ",
        description: "Backend discussions",
      },
      token,
    );

    expect(room).toMatchObject({
      id: "room-2",
      name: "Backend Guild",
      type: "group",
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/rooms",
      expect.objectContaining({
        method: "POST",
        headers: {
          Accept: "application/json",
          Authorization: "Bearer access-token",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: "Backend Guild",
          type: "group",
          member_ids: [],
        }),
      }),
    );
  });

  it("wraps API failures as ApiError", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        jsonResponse(
          {
            success: false,
            error: {
              code: "service_unavailable",
              message: "room service unavailable",
            },
          },
          503,
        ),
      ),
    );

    await expect(listRooms(token)).rejects.toMatchObject({
      name: "ApiError",
      code: "service_unavailable",
      status: 503,
      message: "room service unavailable",
    });
  });

  it("requires token before making a request", async () => {
    vi.stubGlobal("fetch", vi.fn<typeof fetch>());

    await expect(listRooms("")).rejects.toBeInstanceOf(ApiError);
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
