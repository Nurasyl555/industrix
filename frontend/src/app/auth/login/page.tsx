"use client";

// src/app/auth/login/page.tsx
// Figma DS-011 — matches Image 4 exactly
// POST /auth/email/login → TokenPair → httpOnly cookie → redirect /

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { loginWithEmail, friendlyError } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function LoginPage() {
  const router = useRouter();

  const [email, setEmail]       = useState("");
  const [password, setPassword] = useState("");
  const [emailErr, setEmailErr]       = useState("");
  const [passwordErr, setPasswordErr] = useState("");
  const [apiError, setApiError] = useState("");
  const [loading, setLoading]   = useState(false);

  function validate() {
    let ok = true;
    setEmailErr(""); setPasswordErr("");
    if (!email.trim()) {
      setEmailErr("Email is required"); ok = false;
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setEmailErr("Enter a valid email address"); ok = false;
    }
    if (!password) {
      setPasswordErr("Password is required"); ok = false;
    } else if (password.length < 8) {
      setPasswordErr("At least 8 characters"); ok = false;
    }
    return ok;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setApiError("");
    if (!validate()) return;
    setLoading(true);
    try {
      const tokens = await loginWithEmail(email, password);
      await fetch("/api/auth/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tokens),
      });
      router.push("/");
      router.refresh();
    } catch (err) {
      setApiError(friendlyError(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-md shadow-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">Log In</CardTitle>
        </CardHeader>

        <CardContent>
          {apiError && (
            <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {apiError}
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate className="space-y-4">
            <div className="space-y-1.5">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="Email..."
                value={email}
                onChange={(e) => { setEmail(e.target.value); setEmailErr(""); }}
                className={emailErr ? "border-destructive" : ""}
              />
              {emailErr && <p className="text-xs text-destructive">{emailErr}</p>}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                placeholder="Password"
                value={password}
                onChange={(e) => { setPassword(e.target.value); setPasswordErr(""); }}
                className={passwordErr ? "border-destructive" : ""}
              />
              {passwordErr && <p className="text-xs text-destructive">{passwordErr}</p>}
            </div>

            <div className="text-right">
              <Link
                href="/auth/forgot-password"
                className="text-sm text-primary hover:underline"
              >
                Forgot password?
              </Link>
            </div>

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Signing in…" : "Continue"}
            </Button>
          </form>
        </CardContent>

        <CardFooter className="flex-col gap-1 text-center text-sm text-muted-foreground">
          <span>Already have an account?</span>
          <Link href="/auth/register" className="font-semibold text-primary hover:underline">
            Sign In
          </Link>
        </CardFooter>
      </Card>
    </div>
  );
}
