"use client";

// src/app/auth/verify/page.tsx
// ⚠️ DEMO MODE: accepts "123456" without hitting the backend
// Remove DEMO_CODE and the demo branch in handleVerify() before production
import { Suspense } from "react";
import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { verifyPhoneOTP, requestPhoneOTP, friendlyError } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { cn } from "@/lib/utils";

const OTP_LENGTH  = 6;
const TIMER_START = 60;
const DEMO_CODE   = "123456"; // ⚠️ remove before production

function VerifyContent() {
  const router  = useRouter();
  const params  = useSearchParams();
  const phone   = params.get("phone") ?? "";
  const email   = params.get("email") ?? "";

  const [digits, setDigits]           = useState<string[]>(Array(OTP_LENGTH).fill(""));
  const inputRefs                     = useRef<(HTMLInputElement | null)[]>([]);
  const [apiError, setApiError]       = useState("");
  const [loading, setLoading]         = useState(false);
  const [success, setSuccess]         = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(TIMER_START);
  const [canResend, setCanResend]     = useState(false);
  const timerRef                      = useRef<ReturnType<typeof setInterval> | null>(null);

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

  useEffect(() => {
    startTimer();
    setTimeout(() => inputRefs.current[0]?.focus(), 100);
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [startTimer]);

  const formatTime = (s: number) =>
    `${String(Math.floor(s / 60)).padStart(2, "0")}:${String(s % 60).padStart(2, "0")}`;

  function handleDigitChange(index: number, value: string) {
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
      if (digits[index]) { const next = [...digits]; next[index] = ""; setDigits(next); }
      else if (index > 0) inputRefs.current[index - 1]?.focus();
    }
    if (e.key === "ArrowLeft" && index > 0) inputRefs.current[index - 1]?.focus();
    if (e.key === "ArrowRight" && index < OTP_LENGTH - 1) inputRefs.current[index + 1]?.focus();
  }

  const otpCode    = digits.join("");
  const isComplete = otpCode.length === OTP_LENGTH;

  async function handleVerify() {
    if (!isComplete || loading || success) return;
    setApiError("");
    setLoading(true);

    // ⚠️ DEMO bypass — remove this block when backend is ready
    if (otpCode === DEMO_CODE) {
      await new Promise((r) => setTimeout(r, 600));
      setSuccess(true);
      setLoading(false);
      setTimeout(() => { router.push("/"); router.refresh(); }, 1000);
      return;
    }

    // Real API
    try {
      const tokens = await verifyPhoneOTP(phone, otpCode);
      await fetch("/api/auth/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tokens),
      });
      setSuccess(true);
      setTimeout(() => { router.push("/"); router.refresh(); }, 1000);
    } catch (err) {
      const e = err as { code?: string };
      if (e?.code === "UNAUTHORIZED") {
        setApiError("Incorrect code. Please check and try again.");
      } else if (e?.code === "NOT_FOUND") {
        setApiError("This code has expired. Please request a new one.");
        setDigits(Array(OTP_LENGTH).fill(""));
        inputRefs.current[0]?.focus();
      } else {
        setApiError(friendlyError(err));
      }
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (isComplete && !loading && !success) handleVerify();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isComplete, otpCode]);

  async function handleResend() {
    if (!canResend) return;
    setDigits(Array(OTP_LENGTH).fill(""));
    setApiError("");
    try {
      await requestPhoneOTP(phone);
      startTimer();
      setTimeout(() => inputRefs.current[0]?.focus(), 100);
    } catch (err) {
      setApiError(friendlyError(err));
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-lg shadow-sm">
        <CardHeader className="text-center pb-4">
          <CardTitle className="text-2xl font-bold">Verify your account</CardTitle>
          <CardDescription>
            Enter the 6-digit code we sent to{" "}
            <span className="font-medium text-foreground">{phone || email}</span>
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">

          {/* ⚠️ Demo hint — remove before production */}
          <div className="rounded-md border border-dashed border-amber-300 bg-amber-50 px-4 py-2 text-center text-xs text-amber-700">
            Demo mode — use code <span className="font-mono font-bold">{DEMO_CODE}</span>
          </div>

          {success && (
            <div className="rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
              Verified! Redirecting…
            </div>
          )}

          {apiError && (
            <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {apiError}
            </div>
          )}

          <div className="flex justify-center gap-2">
            {digits.map((digit, i) => (
              <input
                key={i}
                ref={(el) => { inputRefs.current[i] = el; }}
                type="text"
                inputMode="numeric"
                maxLength={6}
                value={digit}
                onChange={(e) => handleDigitChange(i, e.target.value)}
                onKeyDown={(e) => handleKeyDown(i, e)}
                onFocus={(e) => e.target.select()}
                disabled={loading || success}
                aria-label={`Digit ${i + 1}`}
                className={cn(
                  "w-12 h-12 text-center text-lg font-semibold rounded-lg border bg-background",
                  "focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent",
                  "transition-colors caret-transparent",
                  digit ? "border-primary" : "border-border",
                  apiError ? "border-destructive" : "",
                  (loading || success) ? "opacity-60 cursor-not-allowed" : ""
                )}
              />
            ))}
          </div>

          <p className="text-center text-sm text-muted-foreground">
            {canResend ? (
              <button type="button" onClick={handleResend} className="text-primary hover:underline font-medium">
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

          <Button className="w-full" onClick={handleVerify} disabled={!isComplete || loading || success}>
            {loading ? "Verifying…" : "Verify"}
          </Button>

        </CardContent>
      </Card>
    </div>
  );
}

export default function VerifyPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center">Loading...</div>}>
      <VerifyContent />
    </Suspense>
  );
}