// lib/user.ts
// Current-user / session helpers used by the Navbar and account pages.

import { authGet } from "./api";

export interface CurrentUser {
  id: string;
  email: string | null;
  phone: string | null;
  first_name: string;
  last_name: string;
  avatar_url: string;
  company_id: string;
}

/** Returns the signed-in user, or null if not authenticated. */
export async function getCurrentUser(): Promise<CurrentUser | null> {
  try {
    return await authGet<CurrentUser>("/users/me");
  } catch {
    return null;
  }
}

/** Clears the session cookie (server-side). */
export async function logout(): Promise<void> {
  await fetch("/api/auth/session", { method: "DELETE" });
}
