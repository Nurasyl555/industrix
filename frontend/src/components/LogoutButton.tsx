"use client";

// src/components/LogoutButton.tsx

export default function LogoutButton() {
  async function handleLogout() {
    await fetch("/api/auth/session", { method: "DELETE" });
    window.location.reload();
  }

  return (
    <button
      type="button"
      onClick={handleLogout}
      className="w-full rounded-lg border border-destructive/40 py-2.5 text-sm font-medium text-destructive hover:bg-destructive/5 transition-colors"
    >
      Clear session (logout)
    </button>
  );
}
