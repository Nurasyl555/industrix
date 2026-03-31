// lib/auth.ts
// All calls to the Identity Module (modules/identity/)
// Base URL from env — set NEXT_PUBLIC_API_URL in .env.local

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

// ── Types ────────────────────────────────────────────────────────────────────

export interface TokenPair {
  accessToken: string;
  refreshToken: string;
}

export interface ApiError {
  code: "NOT_FOUND" | "UNAUTHORIZED" | "VALIDATION" | "CONFLICT" | "INTERNAL";
  message: string;
  details?: Record<string, unknown>;
}

// ── Helpers ───────────────────────────────────────────────────────────────────

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const err: ApiError = await res.json().catch(() => ({
      code: "INTERNAL",
      message: `Request failed (${res.status})`,
    }));
    throw err;
  }

  // 201 Created with empty body
  if (res.status === 201) return undefined as T;

  return res.json();
}

// ── Identity Module endpoints ─────────────────────────────────────────────────

/** Step 1 of registration — create account, triggers OTP via email/phone */
export function registerWithEmail(email: string, password: string) {
  return post<void>("/auth/email/register", { email, password });
}

/** Login with email + password → returns JWT pair */
export function loginWithEmail(email: string, password: string) {
  return post<TokenPair>("/auth/email/login", { email, password });
}

/** Request OTP for phone login/registration */
export function requestPhoneOTP(phone: string) {
  return post<void>("/auth/phone/login", { phone });
}

/** Verify 6-digit OTP → returns JWT pair */
export function verifyPhoneOTP(phone: string, code: string) {
  return post<TokenPair>("/auth/phone/verify", { phone, code });
}

/** Refresh access token */
export function refreshToken(refresh_token: string) {
  return post<TokenPair>("/auth/refresh", { refresh_token });
}

// ── Error helpers ─────────────────────────────────────────────────────────────

export function isApiError(e: unknown): e is ApiError {
  return typeof e === "object" && e !== null && "code" in e && "message" in e;
}

export function friendlyError(e: unknown): string {
  if (isApiError(e)) {
    switch (e.code) {
      case "UNAUTHORIZED":
        return "Incorrect email or password.";
      case "CONFLICT":
        return "An account with this email already exists.";
      case "VALIDATION":
        return e.message ?? "Please check your input.";
      case "NOT_FOUND":
        return "Account not found.";
      default:
        return e.message ?? "Something went wrong. Please try again.";
    }
  }
  return "Something went wrong. Please try again.";
}
