import { ConnectError, createClient } from "@connectrpc/connect";

import { apiUrl, transport } from "./client";
import {
  GatewayService,
  type RefreshTokenResponse,
} from "../../gen/pylon/gateway/v1/gateway_service_pb";

export type ApiErrorCode =
  | "bad_request"
  | "unauthorized"
  | "forbidden"
  | "not_found"
  | "already_exists"
  | "service_unavailable"
  | "network_error"
  | "invalid_response"
  | "internal_error"
  | "unknown_error";

export class ApiError extends Error {
  readonly code: ApiErrorCode | string;
  readonly status: number;
  readonly details?: unknown;

  constructor(input: {
    code: ApiErrorCode | string;
    message: string;
    status: number;
    details?: unknown;
  }) {
    super(input.message);
    this.name = "ApiError";
    this.code = input.code;
    this.status = input.status;
    this.details = input.details;
  }
}

export type AuthUser = {
  id: string;
  username: string;
  email: string;
  display_name?: string;
  avatar_url?: string;
  created_at?: string;
};

export type AuthResult = {
  user: AuthUser;
  token: string;
  refresh_token: string;
  expires_at: string;
  refresh_expires_at: string;
};

export type RefreshResult = {
  token: string;
  refresh_token: string;
  expires_at: string;
  refresh_expires_at: string;
};

export type RegisterInput = {
  username: string;
  email: string;
  password: string;
};

export type LoginInput = {
  email: string;
  password: string;
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

const gatewayClient = createClient(GatewayService, transport);

export async function register(input: RegisterInput): Promise<AuthResult> {
  return postJSON<AuthResult>("/api/v1/auth/register", input);
}

export async function login(input: LoginInput): Promise<AuthResult> {
  return postJSON<AuthResult>("/api/v1/auth/login", input);
}

export async function refreshToken(
  refreshToken: string,
): Promise<RefreshResult> {
  try {
    const response = await gatewayClient.refreshToken({
      refreshToken,
    });

    return refreshTokenResponseToResult(response);
  } catch (error) {
    throw toApiError(error);
  }
}

async function postJSON<T>(path: `/${string}`, body: unknown): Promise<T> {
  let response: Response;

  try {
    response = await fetch(apiUrl(path), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify(body),
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

function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error;
  }

  if (error instanceof ConnectError) {
    return new ApiError({
      code: error.code.toString(),
      message: error.rawMessage || error.message,
      status: 0,
      details: error,
    });
  }

  if (error instanceof Error) {
    return new ApiError({
      code: "unknown_error",
      message: error.message,
      status: 0,
      details: error,
    });
  }

  return new ApiError({
    code: "unknown_error",
    message: "Unknown API error.",
    status: 0,
    details: error,
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

type RefreshTokenResponseWithRefreshExpiry = RefreshTokenResponse & {
  refreshExpiresAt?: { seconds: bigint; nanos: number };
};

function refreshTokenResponseToResult(
  response: RefreshTokenResponse,
): RefreshResult {
  const responseWithRefreshExpiry =
    response as RefreshTokenResponseWithRefreshExpiry;

  return {
    token: response.accessToken,
    refresh_token: response.refreshToken,
    expires_at: timestampToISO(response.expiresAt),
    refresh_expires_at: timestampToISO(
      responseWithRefreshExpiry.refreshExpiresAt,
    ),
  };
}

function timestampToISO(
  value: { seconds: bigint; nanos: number } | undefined,
): string {
  if (!value) {
    return "";
  }

  const milliseconds =
    Number(value.seconds) * 1000 + Math.floor(value.nanos / 1_000_000);

  return new Date(milliseconds).toISOString();
}
