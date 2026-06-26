import { createConnectTransport } from "@connectrpc/connect-web";
import { PUBLIC_PYLON_API_URL } from "$env/static/public";

const rawApiBaseUrl = (PUBLIC_PYLON_API_URL ?? "").trim();

export const apiBaseUrl = rawApiBaseUrl.replace(/\/+$/, "");

export function apiUrl(path: `/${string}`): string {
  return `${apiBaseUrl}${path}`;
}

export const transport = createConnectTransport({
  baseUrl: apiUrl("/api"),
});
