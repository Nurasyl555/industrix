// src/app/page.tsx
// Demo page — quick access to all auth flows for testing

import Link from "next/link";
import { cookies } from "next/headers";
import LogoutButton from "@/components/LogoutButton";

export default async function HomePage() {
  const cookieStore = await cookies();
  const isLoggedIn  = !!cookieStore.get("access_token")?.value;

  return (
    <div className="min-h-screen bg-muted/40 flex items-center justify-center px-4">
      <div className="w-full max-w-md space-y-6">

        {/* Status badge */}
        <div className="text-center space-y-1">
          <h1 className="text-2xl font-bold">IndustriX — Dev Sandbox</h1>
          <p className="text-sm text-muted-foreground">Auth flow testing</p>
          <span className={`inline-block mt-2 text-xs font-medium px-3 py-1 rounded-full ${
            isLoggedIn
              ? "bg-green-100 text-green-800"
              : "bg-muted text-muted-foreground"
          }`}>
            {isLoggedIn ? "● Logged in (access_token cookie present)" : "○ Not logged in"}
          </span>
        </div>

        {/* Auth pages */}
        <div className="rounded-xl border bg-background shadow-sm overflow-hidden">
          <div className="px-4 py-3 border-b bg-muted/30">
            <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Identity Module flows
            </p>
          </div>
          <div className="divide-y">
            <DemoLink
              href="/auth/register"
              label="Register"
              description="Role select → details → password → OTP"
              method="POST /auth/email/register"
            />
            <DemoLink
              href="/auth/login"
              label="Login"
              description="Email + password → JWT cookie"
              method="POST /auth/email/login"
            />
            <DemoLink
              href="/auth/verify?phone=%2B77001234567&email=test@example.com"
              label="Verify OTP"
              description="6-digit box input + countdown timer"
              method="POST /auth/phone/verify"
            />
            <DemoLink
              href="/auth/forgot-password"
              label="Forgot password"
              description="Send code → verify OTP"
              method="POST /auth/phone/login"
            />
          </div>
        </div>

        {/* Integrity Module */}
        <div className="rounded-xl border bg-background shadow-sm overflow-hidden">
          <div className="px-4 py-3 border-b bg-muted/30">
            <p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
              Integrity Module flows
            </p>
          </div>
          <div className="divide-y">
            <DemoLink
              href="/company/register"
              label="Register company"
              description="BIN mask (12-digit) → company profile"
              method="POST /api/v1/companies · requires login"
            />
          </div>
        </div>

        {/* Logout button — only when logged in */}
        {isLoggedIn && (
          <form action="/api/auth/session" method="DELETE">
            <LogoutButton />
          </form>
        )}

        <p className="text-center text-xs text-muted-foreground">
          This page is for development only — remove before production
        </p>
      </div>
    </div>
  );
}

// ── Sub-components ────────────────────────────────────────────────────────────

function DemoLink({
  href,
  label,
  description,
  method,
}: {
  href: string;
  label: string;
  description: string;
  method: string;
}) {
  return (
    <Link
      href={href}
      className="flex items-center justify-between px-4 py-3.5 hover:bg-muted/40 transition-colors group"
    >
      <div className="space-y-0.5">
        <p className="text-sm font-medium group-hover:text-primary transition-colors">
          {label}
        </p>
        <p className="text-xs text-muted-foreground">{description}</p>
        <p className="text-xs font-mono text-muted-foreground/60">{method}</p>
      </div>
      <span className="text-muted-foreground group-hover:text-primary transition-colors text-lg">
        →
      </span>
    </Link>
  );
}
