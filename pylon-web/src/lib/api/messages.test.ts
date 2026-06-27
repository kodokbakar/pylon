import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "$lib/api/auth";

import { listMessages } from "./messages";

const token = "access-token";

describe("messages API", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("loads room messages with bearer auth and sorts chronologically", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      jsonResponse({
        data: {
          messages: [
            {
              id: "msg-2",
              room_id: "room-1",
              sender_id: "user-2",
              sender_username: "bob",
              sender_display_name: "Bob",
              sender_avatar_url: "",
              content: "Second message",
              type: "MESSAGE_TYPE_TEXT",
              created_at: "2026-06-27T12:01:00Z",
            },
            {
              id: "msg-1",
              room_id: "room-1",
              sender_id: "user-1",
              sender_username: "alice",
              sender_display_name: "Alice",
              sender_avatar_url: "",
              content: "First message",
              type: 1,
              created_at: "2026-06-27T12:00:00Z",
            },
          ],
          has_more: true,
        },
      }),
    );

    vi.stubGlobal("fetch", fetchMock);

    const history = await listMessages("room-1", token, {
      limit: 25,
    });

    expect(history).toEqual({
      has_more: true,
      messages: [
        {
          id: "msg-1",
          room_id: "room-1",
          sender_id: "user-1",
          sender_username: "alice",
          sender_display_name: "Alice",
          sender_avatar_url: "",
          content: "First message",
          type: "text",
          created_at: "2026-06-27T12:00:00Z",
        },
        {
          id: "msg-2",
          room_id: "room-1",
          sender_id: "user-2",
          sender_username: "bob",
          sender_display_name: "Bob",
          sender_avatar_url: "",
          content: "Second message",
          type: "text",
          created_at: "2026-06-27T12:01:00Z",
        },
      ],
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/rooms/room-1/messages?limit=25",
      expect.objectContaining({
        method: "GET",
        headers: {
          Accept: "application/json",
          Authorization: "Bearer access-token",
        },
      }),
    );
  });

  it("passes before_id for paginated history", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      jsonResponse({
        data: {
          messages: [],
          has_more: false,
        },
      }),
    );

    vi.stubGlobal("fetch", fetchMock);

    await listMessages("room-1", token, {
      limit: 10,
      before_id: "msg-10",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/v1/rooms/room-1/messages?limit=10&before_id=msg-10",
      expect.objectContaining({
        method: "GET",
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
              code: "not_found",
              message: "room not found",
            },
          },
          404,
        ),
      ),
    );

    await expect(listMessages("room-1", token)).rejects.toMatchObject({
      name: "ApiError",
      code: "not_found",
      status: 404,
      message: "room not found",
    });
  });

  it("requires room id before making a request", async () => {
    vi.stubGlobal("fetch", vi.fn<typeof fetch>());

    await expect(listMessages(" ", token)).rejects.toBeInstanceOf(ApiError);
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
