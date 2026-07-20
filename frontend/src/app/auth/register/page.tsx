"use client";

// src/app/auth/register/page.tsx
// Step 1: Role select (Image 1)
// Step 2: Full name / email / phone (Image 2)
// Step 3: Create password (Image 3)
// Identity Module: POST /auth/email/register → redirect to /auth/verify

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { registerWithEmail, friendlyError } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardFooter, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { useI18n, type TranslationKey } from "@/lib/i18n";

// The role's stored value is separate from its label: the label is translated,
// so using it as the value would make the selection change meaning with the
// interface language.
type Role = "buyer" | "seller" | "provider";
type Step = 1 | 2 | 3;

const ROLES: { value: Role; labelKey: TranslationKey }[] = [
  { value: "buyer", labelKey: "role.buyer" },
  { value: "seller", labelKey: "role.seller" },
  { value: "provider", labelKey: "role.provider" },
];

export default function RegisterPage() {
  const router = useRouter();
  const { t } = useI18n();

  const [step, setStep] = useState<Step>(1);

  // Step 1
  const [role, setRole] = useState<Role | null>(null);
  const [roleErr, setRoleErr] = useState("");

  // Step 2
  const [fullName, setFullName] = useState("");
  const [email, setEmail]       = useState("");
  const [phone, setPhone]       = useState("");
  const [nameErr, setNameErr]   = useState("");
  const [emailErr, setEmailErr] = useState("");
  const [phoneErr, setPhoneErr] = useState("");

  // Step 3
  const [password, setPassword]           = useState("");
  const [confirmPassword, setConfirm]     = useState("");
  const [passwordErr, setPasswordErr]     = useState("");
  const [confirmErr, setConfirmErr]       = useState("");

  const [apiError, setApiError] = useState("");
  const [loading, setLoading]   = useState(false);

  // ── Validation ──────────────────────────────────────────────────────────────

  function validateStep1() {
    if (!role) { setRoleErr(t("valid.roleRequired")); return false; }
    setRoleErr(""); return true;
  }

  function validateStep2() {
    let ok = true;
    setNameErr(""); setEmailErr(""); setPhoneErr("");
    if (!fullName.trim()) { setNameErr(t("valid.nameRequired")); ok = false; }
    if (!email.trim()) { setEmailErr(t("valid.emailRequired")); ok = false; }
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { setEmailErr(t("valid.emailInvalid")); ok = false; }
    if (!phone.trim()) { setPhoneErr(t("valid.phoneRequired")); ok = false; }
    else if (!/^\+?[\d\s\-()]{7,15}$/.test(phone)) { setPhoneErr(t("valid.phoneInvalid")); ok = false; }
    return ok;
  }

  function validateStep3() {
    let ok = true;
    setPasswordErr(""); setConfirmErr("");
    if (!password) { setPasswordErr(t("valid.passwordRequired")); ok = false; }
    else if (password.length < 8 || !/[a-zA-Z]/.test(password) || !/\d/.test(password)) {
      setPasswordErr(t("valid.passwordWeak")); ok = false;
    }
    if (!confirmPassword) { setConfirmErr(t("valid.passwordRepeat")); ok = false; }
    else if (password !== confirmPassword) { setConfirmErr(t("valid.passwordMismatch")); ok = false; }
    return ok;
  }

  function handleNext() {
    if (step === 1 && validateStep1()) setStep(2);
    else if (step === 2 && validateStep2()) setStep(3);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setApiError("");
    if (!validateStep3()) return;
    setLoading(true);
    try {
      const tokens = await registerWithEmail(email, password, fullName);
      await fetch("/api/auth/session", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(tokens),
      });
      router.push("/");
      router.refresh();
    } catch (err) {
      setApiError(friendlyError(err));
      const e = err as { code?: string };
      if (e?.code === "CONFLICT") setStep(2);
    } finally {
      setLoading(false);
    }
  }

  // ── Shared footer ────────────────────────────────────────────────────────────

  const Footer = () => (
    <CardFooter className="flex-col gap-0.5 text-center text-sm text-muted-foreground pt-0">
      <span>{t("auth.register.haveAccount")}</span>
      <Link href="/auth/login" className="font-bold text-primary hover:underline">
        {t("auth.register.toLogin")}
      </Link>
    </CardFooter>
  );

  // ── Render ───────────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-lg shadow-sm">

        {/* ── Step 1: Role ── */}
        {step === 1 && (
          <>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-2xl font-bold">{t("auth.register.title")}</CardTitle>
              <CardDescription>{t("auth.register.chooseRole")}</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              {apiError && (
                <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}
              {ROLES.map((r) => (
                <button
                  key={r.value}
                  type="button"
                  onClick={() => { setRole(r.value); setRoleErr(""); }}
                  className={cn(
                    "w-full text-left px-4 py-3 rounded-lg border text-sm transition-colors",
                    role === r.value
                      ? "border-primary bg-background font-medium"
                      : "border-border bg-background hover:border-muted-foreground/40"
                  )}
                >
                  {t(r.labelKey)}
                </button>
              ))}
              {roleErr && <p className="text-xs text-destructive">{roleErr}</p>}
              <div className="pt-2">
                <Button className="w-full" onClick={handleNext}>{t("common.continue")}</Button>
              </div>
            </CardContent>
            <Footer />
          </>
        )}

        {/* ── Step 2: Details ── */}
        {step === 2 && (
          <>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-2xl font-bold">{t("auth.register.title")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {apiError && (
                <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}
              <div>
                <Input
                  placeholder={t("common.fullName")}
                  value={fullName}
                  onChange={(e) => { setFullName(e.target.value); setNameErr(""); }}
                  className={nameErr ? "border-destructive" : ""}
                />
                {nameErr && <p className="text-xs text-destructive mt-1">{nameErr}</p>}
              </div>
              <div>
                <Input
                  type="email"
                  placeholder={t("common.email")}
                  value={email}
                  onChange={(e) => { setEmail(e.target.value); setEmailErr(""); }}
                  className={emailErr ? "border-destructive" : ""}
                />
                {emailErr && <p className="text-xs text-destructive mt-1">{emailErr}</p>}
              </div>
              <div>
                <Input
                  type="tel"
                  placeholder={t("common.phone")}
                  value={phone}
                  onChange={(e) => { setPhone(e.target.value); setPhoneErr(""); }}
                  className={phoneErr ? "border-destructive" : ""}
                />
                {phoneErr && <p className="text-xs text-destructive mt-1">{phoneErr}</p>}
              </div>
              <div className="pt-2">
                <Button className="w-full" onClick={handleNext}>{t("common.continue")}</Button>
              </div>
            </CardContent>
            <Footer />
          </>
        )}

        {/* ── Step 3: Password ── */}
        {step === 3 && (
          <>
            <CardHeader className="text-center pb-4">
              <CardTitle className="text-2xl font-bold">{t("auth.register.passwordTitle")}</CardTitle>
            </CardHeader>
            <CardContent>
              {apiError && (
                <div className="mb-3 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{apiError}</div>
              )}
              <form onSubmit={handleSubmit} noValidate className="space-y-3">
                <div>
                  <Input
                    type="password"
                    placeholder={t("auth.register.createPassword")}
                    value={password}
                    onChange={(e) => { setPassword(e.target.value); setPasswordErr(""); }}
                    className={passwordErr ? "border-destructive" : ""}
                  />
                  {passwordErr && <p className="text-xs text-destructive mt-1">{passwordErr}</p>}
                </div>
                <div>
                  <Input
                    type="password"
                    placeholder={t("auth.register.repeatPassword")}
                    value={confirmPassword}
                    onChange={(e) => { setConfirm(e.target.value); setConfirmErr(""); }}
                    className={confirmErr ? "border-destructive" : ""}
                  />
                  {confirmErr && <p className="text-xs text-destructive mt-1">{confirmErr}</p>}
                </div>
                <p className="text-sm font-medium text-foreground">
                  Use at least 8 characters, including a letter and a number
                </p>
                <div className="pt-2">
                  <Button type="submit" className="w-full" disabled={loading}>
                    {loading ? t("auth.register.creating") : t("common.continue")}
                  </Button>
                </div>
              </form>
            </CardContent>
            <Footer />
          </>
        )}
      </Card>
    </div>
  );
}
