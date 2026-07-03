// lib/api.ts
// Shared fetch helpers for the Catalog/Listing/Deal modules.
//
// Public endpoints (browsing) hit the Go backend directly.
// Authenticated endpoints go through /api/proxy/*, which attaches the
// httpOnly access_token cookie server-side — see app/api/proxy/[...path].

const API = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "");

export interface ApiError {
  code: "NOT_FOUND" | "UNAUTHORIZED" | "VALIDATION" | "CONFLICT" | "INTERNAL";
  message: string;
  details?: Record<string, unknown>;
}

async function parseError(res: Response): Promise<never> {
  const err: ApiError = await res.json().catch(() => ({
    code: "INTERNAL",
    message: `Request failed (${res.status})`,
  }));
  throw err;
}

async function handle<T>(res: Response): Promise<T> {
  if (!res.ok) return parseError(res);
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
}

// ── Public (no auth) — calls the backend directly ──────────────────────────

export function publicGet<T>(path: string): Promise<T> {
  return fetch(`${API}${path}`).then(handle<T>);
}

// ── Authenticated — routed through the same-origin proxy ───────────────────

export function authGet<T>(path: string): Promise<T> {
  return fetch(`/api/proxy${path}`).then(handle<T>);
}

export function authPost<T>(path: string, body: unknown): Promise<T> {
  return fetch(`/api/proxy${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  }).then(handle<T>);
}

export function authPut<T>(path: string, body?: unknown): Promise<T> {
  return fetch(`/api/proxy${path}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: body ? JSON.stringify(body) : undefined,
  }).then(handle<T>);
}

export function authDelete<T>(path: string): Promise<T> {
  return fetch(`/api/proxy${path}`, { method: "DELETE" }).then(handle<T>);
}

// ── Error helpers ─────────────────────────────────────────────────────────

export function isApiError(e: unknown): e is ApiError {
  return typeof e === "object" && e !== null && "code" in e && "message" in e;
}

export function friendlyError(e: unknown): string {
  if (isApiError(e)) {
    switch (e.code) {
      case "UNAUTHORIZED":
        return "Please sign in to continue.";
      case "CONFLICT":
        return "This already exists.";
      case "VALIDATION":
        return e.message ?? "Please check your input.";
      case "NOT_FOUND":
        return "Not found.";
      default:
        return e.message ?? "Something went wrong. Please try again.";
    }
  }
  return "Something went wrong. Please try again.";
}
