import { ApiError, type ApiErrorCode } from "$lib/api/auth";

import { apiUrl } from "./client";

export type RoomType = "direct" | "group" | "channel" | "unspecified";

export type Room = {
  id: string;
  name: string;
  type: RoomType;
  created_by: string;
  created_at: string;
  description?: string;
};

export type CreateRoomInput = {
  name: string;
  description?: string;
  type?: Extract<RoomType, "group" | "channel" | "direct">;
  member_ids?: string[];
};

type ApiSuccess<T> = {
  success?: boolean;
  data: T;
};

type ApiFailure = {
  success?: false;
  error?: {
    code?: string;
    message?: string;
    details?: unknown;
  };
  code?: string;
  message?: string;
};

type RoomDTO = {
  id?: unknown;
  name?: unknown;
  type?: unknown;
  created_by?: unknown;
  createdBy?: unknown;
  created_at?: unknown;
  createdAt?: unknown;
};

type ListRoomsPayload = {
  rooms?: RoomDTO[];
};

type CreateRoomPayload = {
  room?: RoomDTO;
};

export async function listRooms(token: string): Promise<Room[]> {
  const payload = await requestJSON<ListRoomsPayload>("/api/v1/rooms", {
    method: "GET",
    token,
  });

  return Array.isArray(payload.rooms) ? payload.rooms.map(roomFromDTO) : [];
}

export async function createRoom(
  input: CreateRoomInput,
  token: string,
): Promise<Room> {
  const payload = await requestJSON<CreateRoomPayload>("/api/v1/rooms", {
    method: "POST",
    token,
    body: {
      name: input.name.trim(),
      type: input.type ?? "group",
      member_ids: input.member_ids ?? [],
    },
  });

  if (!payload.room) {
    throw new ApiError({
      code: "invalid_response",
      message: "Pylon API returned an invalid room response.",
      status: 0,
      details: payload,
    });
  }

  return roomFromDTO(payload.room);
}

async function requestJSON<T>(
  path: `/${string}`,
  options: {
    method: "GET" | "POST";
    token: string;
    body?: unknown;
  },
): Promise<T> {
  const token = options.token.trim();
  if (token === "") {
    throw new ApiError({
      code: "unauthorized",
      message: "Authentication token is required.",
      status: 401,
    });
  }

  let response: Response;

  try {
    response = await fetch(apiUrl(path), {
      method: options.method,
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${token}`,
        ...(options.body === undefined
          ? {}
          : { "Content-Type": "application/json" }),
      },
      body:
        options.body === undefined ? undefined : JSON.stringify(options.body),
    });
  } catch (error) {
    throw new ApiError({
      code: "network_error",
      message: "Unable to reach Pylon API.",
      status: 0,
      details: error,
    });
  }

  const payload = await readJSON<ApiSuccess<T> | ApiFailure>(response);

  if (!response.ok) {
    throw apiFailureToError(payload, response.status);
  }

  if (!isSuccessPayload<T>(payload)) {
    throw new ApiError({
      code: "invalid_response",
      message: "Pylon API returned an invalid response.",
      status: response.status,
      details: payload,
    });
  }

  return payload.data;
}

async function readJSON<T>(response: Response): Promise<T> {
  try {
    return (await response.json()) as T;
  } catch (error) {
    throw new ApiError({
      code: "invalid_response",
      message: "Pylon API returned invalid JSON.",
      status: response.status,
      details: error,
    });
  }
}

function isSuccessPayload<T>(
  payload: ApiSuccess<T> | ApiFailure,
): payload is ApiSuccess<T> {
  return typeof payload === "object" && payload !== null && "data" in payload;
}

function apiFailureToError(
  payload: ApiSuccess<unknown> | ApiFailure,
  status: number,
): ApiError {
  const failure = payload as ApiFailure;
  const code = failure.error?.code ?? failure.code ?? statusToCode(status);
  const message =
    failure.error?.message ?? failure.message ?? "Pylon API request failed.";

  return new ApiError({
    code,
    message,
    status,
    details: payload,
  });
}

function statusToCode(status: number): ApiErrorCode {
  switch (status) {
    case 400:
      return "bad_request";
    case 401:
      return "unauthorized";
    case 403:
      return "forbidden";
    case 404:
      return "not_found";
    case 409:
      return "already_exists";
    case 503:
      return "service_unavailable";
    default:
      return "internal_error";
  }
}

function roomFromDTO(dto: RoomDTO): Room {
  return {
    id: stringValue(dto.id),
    name: stringValue(dto.name),
    type: roomTypeValue(dto.type),
    created_by: stringValue(dto.created_by ?? dto.createdBy),
    created_at: timestampValue(dto.created_at ?? dto.createdAt),
  };
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function roomTypeValue(value: unknown): RoomType {
  if (typeof value === "number") {
    switch (value) {
      case 1:
        return "direct";
      case 2:
        return "group";
      case 3:
        return "channel";
      default:
        return "unspecified";
    }
  }

  if (typeof value !== "string") {
    return "unspecified";
  }

  const normalized = value
    .trim()
    .toLowerCase()
    .replace(/^room_type_/, "");

  switch (normalized) {
    case "direct":
      return "direct";
    case "group":
      return "group";
    case "channel":
      return "channel";
    default:
      return "unspecified";
  }
}

function timestampValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }

  if (typeof value !== "object" || value === null) {
    return "";
  }

  const timestamp = value as {
    seconds?: number | string | bigint;
    nanos?: number;
  };

  if (timestamp.seconds === undefined) {
    return "";
  }

  const seconds =
    typeof timestamp.seconds === "bigint"
      ? Number(timestamp.seconds)
      : Number(timestamp.seconds);

  if (!Number.isFinite(seconds)) {
    return "";
  }

  const milliseconds =
    seconds * 1000 + Math.floor((timestamp.nanos ?? 0) / 1_000_000);

  return new Date(milliseconds).toISOString();
}
