"use client";

// src/app/auth/forgot-password/page.tsx
// Step 1 (Image 5): Enter email or phone → Send code
// Step 2 (Image 6): 6-digit OTP boxes + countdown timer + Verify
// Identity Module: POST /auth/phone/login (send OTP) → POST /auth/phone/verify

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { requestPhoneOTP, verifyPhoneOTP, friendlyError } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { cn } from "@/lib/utils";

const OTP_LENGTH   = 6;
const TIMER_START  = 60; // seconds — matches Redis OTP TTL

export default function ForgotPasswordPage() {
  const router = useRouter();

  const [step, setStep]         = useState<1 | 2>(1);
  const [contact, setContact]   = useState("");  // email or phone
  const [contactErr, setContactErr] = useState("");
  const [apiError, setApiError] = useState("");
  const [loading, setLoading]   = useState(false);

  // ── OTP state ────────────────────────────────────────────────────────────────
  const [digits, setDigits]     = useState<string[]>(Array(OTP_LENGTH).fill(""));
  const inputRefs               = useRef<(HTMLInputElement | null)[]>([]);
  const [secondsLeft, setSecondsLeft] = useState(TIMER_START);
  const [canResend, setCanResend]     = useState(false);
  const timerRef                = useRef<ReturnType<typeof setInterval> | null>(null);

  // ── Timer ─────────────────────────────────────────────────────────────────────
  const startTimer = useCallback(() => {
    setSecondsLeft(TIMER_START);
    setCanResend(false);
    if (timerRef.current) clearInterval(timerRef.current);
    timerRef.current = setInterval(() => {
      setSecondsLeft((s) => {
        if (s <= 1) { clearInterval(timerRef.current!); setCanResend(true); return 0; }
        return s - 1;
      });
    }, 1000);
  }, []);

  useEffect(() => () => { if (timerRef.current) clearInterval(timerRef.current); }, []);

  const formatTime = (s: number) =>
    `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;

  // ── Step 1: Send code ─────────────────────────────────────────────────────────
  async function handleSendCode(e: React.FormEvent) {
    e.preventDefault();
    setContactErr(""); setApiError("");
    if (!contact.trim()) { setContactErr("Please enter your email or phone number"); return; }
    setLoading(true);
    try {
      // Identity module expects phone for OTP; if email is entered backend resolves it
      await requestPhoneOTP(contact.trim());
      setStep(2);
      startTimer();
      // Focus first OTP box after render
      setTimeout(() => inputRefs.current[0]?.focus(), 100);
    } catch (err) {
      setApiError(friendlyError(err));
    } finally {
      setLoading(false);
    }
  }

  // ── OTP input handlers ────────────────────────────────────────────────────────
  function handleDigitChange(index: number, value: string) {
    // Handle paste of full code
    if (value.length > 1) {
      const cleaned = value.replace(/\D/g, "").slice(0, OTP_LENGTH);
      const next = [...digits];
      for (let i = 0; i < OTP_LENGTH; i++) next[i] = cleaned[i] ?? "";
      setDigits(next);
      setApiError("");
      inputRefs.current[Math.min(cleaned.length, OTP_LENGTH - 1)]?.focus();
      return;
    }
    const digit = value.replace(/\D/g, "").slice(-1);
    const next = [...digits];
    next[index] = digit;
    setDigits(next);
    setApiError("");
    if (digit && index < OTP_LENGTH - 1) inputRefs.current[index + 1]?.focus();
  }

  function handleKeyDown(index: number, e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Backspace") {
      if (digits[index]) {
        const next = [...digits]; next[index] = ""; setDigits(next);
      } else if (index > 0) {
        inputRefs.current[index - 1]?.focus();
      }
    }
    if (e.key === "ArrowLeft" && index > 0) inputRefs.current[index - 1]?.focus();
    if (e.key === "ArrowRight" && index < OTP_LENGTH - 1) inputRefs.current[index + 1]?.focus();
  }

  const otpCode   = digits.join("");
  const isComplete = otpCode.length === OTP_LENGTH;

  // ── Step 2: Verify ────────────────────────────────────────────────────────────
  async function handleVerify() {
    if (!isComplete) return;
    setApiError("");
    setLoading(true);
    try {
      const tokens = await verifyPhoneOTP(contact.trim(), otpCode);
      await fetch("/api/auth/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tokens),
      });
      router.push("/");
      router.refresh();
    } catch (err) {
      const e = err as { code?: string };
      if (e?.code === "UNAUTHORIZED") {
        setApiError("Incorrect code. Please check and try again.");
      } else if (e?.code === "NOT_FOUND") {
        setApiError("This code has expired. Request a new one.");
        setDigits(Array(OTP_LENGTH).fill(""));
        inputRefs.current[0]?.focus();
      } else {
        setApiError(friendlyError(err));
      }
    } finally {
      setLoading(false);
    }
  }

  // Auto-submit when all 6 filled
  useEffect(() => {
    if (isComplete && !loading) handleVerify();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isComplete, otpCode]);

  async function handleResend() {
    if (!canResend) return;
    setDigits(Array(OTP_LENGTH).fill(""));
    setApiError("");
    try {
      await requestPhoneOTP(contact.trim());
      startTimer();
      setTimeout(() => inputRefs.current[0]?.focus(), 100);
    } catch (err) {
      setApiError(friendlyError(err));
    }
  }

  // ── Render ────────────────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-lg shadow-sm">

        {/* ── Step 1 ── */}
        {step === 1 && (
          <>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-2xl font-bold">Forgot Password?</CardTitle>
            </CardHeader>
            <CardContent>
              {apiError && (
                <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}
              <form onSubmit={handleSendCode} noValidate className="space-y-4">
                <div>
                  <Input
                    placeholder="Enter your email or phone number"
                    value={contact}
                    onChange={(e) => { setContact(e.target.value); setContactErr(""); }}
                    className={contactErr ? "border-destructive" : ""}
                  />
                  {contactErr && <p className="text-xs text-destructive mt-1">{contactErr}</p>}
                </div>
                <div className="pt-6">
                  <Button type="submit" className="w-full" disabled={loading}>
                    {loading ? "Sending…" : "Send code"}
                  </Button>
                </div>
              </form>
            </CardContent>
          </>
        )}

        {/* ── Step 2: OTP ── */}
        {step === 2 && (
          <>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-2xl font-bold">Forgot Password?</CardTitle>
              <CardDescription className="text-sm">
                Enter the 6-digit code we sent to your email or phone number.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              {apiError && (
                <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}

              {/* 6 individual digit boxes — matching Image 6 exactly */}
              <div className="flex justify-center gap-2">
                {digits.map((digit, i) => (
                  <input
                    key={i}
                    ref={(el) => { inputRefs.current[i] = el; }}
                    type="text"
                    inputMode="numeric"
                    maxLength={6}
                    value={digit || (i === digits.findIndex(d => !d) ? "" : digit)}
                    onChange={(e) => handleDigitChange(i, e.target.value)}
                    onKeyDown={(e) => handleKeyDown(i, e)}
                    onFocus={(e) => e.target.select()}
                    disabled={loading}
                    aria-label={`Digit ${i + 1}`}
                    className={cn(
                      "w-12 h-12 text-center text-lg font-semibold rounded-lg border bg-background",
                      "focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent",
                      "transition-colors caret-transparent",
                      digit ? "border-primary" : "border-border",
                      apiError ? "border-destructive" : "",
                      loading ? "opacity-60 cursor-not-allowed" : ""
                    )}
                  />
                ))}
              </div>

              {/* Timer / resend */}
              <p className="text-center text-sm text-muted-foreground">
                {canResend ? (
                  <button
                    type="button"
                    onClick={handleResend}
                    className="text-primary hover:underline font-medium"
                  >
                    Resend code
                  </button>
                ) : (
                  <>You can resend the code in{" "}
                    <span className="font-mono font-semibold text-foreground tabular-nums">
                      {formatTime(secondsLeft)}
                    </span>
                  </>
                )}
              </p>

              <Button
                className="w-full"
                onClick={handleVerify}
                disabled={!isComplete || loading}
              >
                {loading ? "Verifying…" : "Verify"}
              </Button>
            </CardContent>
          </>
        )}
      </Card>
    </div>
  );
}
