import { ApiError, type ApiErrorCode } from "$lib/api/auth";

import { apiUrl } from "./client";

export type ChatMessageType =
  | "text"
  | "image"
  | "system"
  | "file"
  | "unspecified";

export type ChatMessage = {
  id: string;
  room_id: string;
  sender_id: string;
  sender_username: string;
  sender_display_name: string;
  sender_avatar_url: string;
  content: string;
  type: ChatMessageType;
  created_at: string;
};

export type MessageHistory = {
  messages: ChatMessage[];
  has_more: boolean;
};

export type ListMessagesOptions = {
  limit?: number;
  before_id?: string;
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

type MessageDTO = {
  id?: unknown;
  room_id?: unknown;
  roomId?: unknown;
  sender_id?: unknown;
  senderId?: unknown;
  sender_username?: unknown;
  senderUsername?: unknown;
  sender_display_name?: unknown;
  senderDisplayName?: unknown;
  sender_avatar_url?: unknown;
  senderAvatarUrl?: unknown;
  content?: unknown;
  type?: unknown;
  created_at?: unknown;
  createdAt?: unknown;
};

type GetMessagesPayload = {
  messages?: MessageDTO[];
  has_more?: unknown;
  hasMore?: unknown;
};

export async function listMessages(
  roomID: string,
  token: string,
  options: ListMessagesOptions = {},
): Promise<MessageHistory> {
  const normalizedRoomID = roomID.trim();
  if (normalizedRoomID === "") {
    throw new ApiError({
      code: "bad_request",
      message: "Room ID is required.",
      status: 400,
    });
  }

  const query = new URLSearchParams({
    limit: String(normalizeLimit(options.limit)),
  });

  const beforeID = options.before_id?.trim();
  if (beforeID) {
    query.set("before_id", beforeID);
  }

  const encodedRoomID = encodeURIComponent(normalizedRoomID);
  const payload = await requestJSON<GetMessagesPayload>(
    `/api/v1/rooms/${encodedRoomID}/messages?${query.toString()}` as `/${string}`,
    {
      method: "GET",
      token,
    },
  );

  const messages = Array.isArray(payload.messages)
    ? payload.messages.map(messageFromDTO)
    : [];

  return {
    messages: sortMessagesChronologically(messages),
    has_more: booleanValue(payload.has_more ?? payload.hasMore),
  };
}

async function requestJSON<T>(
  path: `/${string}`,
  options: {
    method: "GET";
    token: string;
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
      },
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

function normalizeLimit(limit: number | undefined): number {
  if (limit === undefined || !Number.isFinite(limit) || limit <= 0) {
    return 50;
  }

  return Math.min(Math.trunc(limit), 100);
}

function messageFromDTO(dto: MessageDTO): ChatMessage {
  return {
    id: stringValue(dto.id),
    room_id: stringValue(dto.room_id ?? dto.roomId),
    sender_id: stringValue(dto.sender_id ?? dto.senderId),
    sender_username: stringValue(dto.sender_username ?? dto.senderUsername),
    sender_display_name: stringValue(
      dto.sender_display_name ?? dto.senderDisplayName,
    ),
    sender_avatar_url: stringValue(
      dto.sender_avatar_url ?? dto.senderAvatarUrl,
    ),
    content: stringValue(dto.content),
    type: messageTypeValue(dto.type),
    created_at: timestampValue(dto.created_at ?? dto.createdAt),
  };
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function booleanValue(value: unknown): boolean {
  return value === true;
}

function messageTypeValue(value: unknown): ChatMessageType {
  if (typeof value === "number") {
    switch (value) {
      case 1:
        return "text";
      case 2:
        return "image";
      case 3:
        return "system";
      case 4:
        return "file";
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
    .replace(/^message_type_/, "");

  switch (normalized) {
    case "text":
      return "text";
    case "image":
      return "image";
    case "system":
      return "system";
    case "file":
      return "file";
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

function sortMessagesChronologically(messages: ChatMessage[]): ChatMessage[] {
  return [...messages].sort((left, right) => {
    const leftTime = Date.parse(left.created_at);
    const rightTime = Date.parse(right.created_at);

    const safeLeftTime = Number.isNaN(leftTime) ? 0 : leftTime;
    const safeRightTime = Number.isNaN(rightTime) ? 0 : rightTime;

    if (safeLeftTime !== safeRightTime) {
      return safeLeftTime - safeRightTime;
    }

    return left.id.localeCompare(right.id);
  });
}
