"use client";

// src/components/CompanyRegistrationForm.tsx
// Integrity Module: POST /api/v1/companies (BearerAuth)
// BIN: 12-digit KZ format, displayed as XXXX XXXX XXXX, sent as raw 12-digit string

import { useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";

// ── Types ─────────────────────────────────────────────────────────────────────

interface CreateCompanyRequest {
  name: string;
  bin: string;
  email: string;
  phone: string;
  address: string;
  website: string;
}

interface FieldErrors {
  name?: string;
  bin?: string;
  email?: string;
  phone?: string;
  address?: string;
}

// ── BIN helpers ───────────────────────────────────────────────────────────────

const sanitizeBIN = (raw: string) => raw.replace(/\D/g, "").slice(0, 12);
const formatBIN   = (digits: string) => digits.replace(/(\d{4})(?=\d)/g, "$1 ");
const isBINValid  = (digits: string) => digits.length === 12;

// ── API ───────────────────────────────────────────────────────────────────────

async function createCompany(data: CreateCompanyRequest, accessToken: string) {
  const res = await fetch("/api/v1/companies", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(data),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err?.message ?? `Request failed (${res.status})`);
  }

  return res.json();
}

// ── Component ─────────────────────────────────────────────────────────────────

interface Props {
  accessToken: string;
  successRedirectPath?: string;
}

export default function CompanyRegistrationForm({
  accessToken,
  successRedirectPath = "/",
}: Props) {
  const router = useRouter();

  const [binDigits, setBinDigits] = useState("");
  const [fields, setFields] = useState({
    name: "",
    email: "",
    phone: "",
    address: "",
    website: "",
  });

  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [apiError, setApiError]       = useState("");
  const [loading, setLoading]         = useState(false);
  const [success, setSuccess]         = useState(false);

  // ── Handlers ─────────────────────────────────────────────────────────────────

  const handleBINChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setBinDigits(sanitizeBIN(e.target.value));
    setFieldErrors((p) => ({ ...p, bin: undefined }));
  }, []);

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFields((p) => ({ ...p, [name]: value }));
    setFieldErrors((p) => ({ ...p, [name]: undefined }));
  }, []);

  // ── Validation ────────────────────────────────────────────────────────────────

  function validate(): boolean {
    const errors: FieldErrors = {};
    if (!fields.name.trim())  errors.name = "Company name is required";
    if (!isBINValid(binDigits)) errors.bin = "BIN must be exactly 12 digits";
    if (!fields.email.trim()) {
      errors.email = "Email is required";
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
      errors.email = "Enter a valid email";
    }
    if (!fields.phone.trim())   errors.phone = "Phone number is required";
    if (!fields.address.trim()) errors.address = "Address is required";
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  }

  // ── Submit ────────────────────────────────────────────────────────────────────

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setApiError("");
    if (!validate()) return;
    setLoading(true);
    try {
      await createCompany(
        {
          name:    fields.name.trim(),
          bin:     binDigits,           // raw 12 digits — no spaces
          email:   fields.email.trim(),
          phone:   fields.phone.trim(),
          address: fields.address.trim(),
          website: fields.website.trim(),
        },
        accessToken
      );
      setSuccess(true);
      setTimeout(() => router.push(successRedirectPath), 1500);
    } catch (err) {
      setApiError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  // ── Render ────────────────────────────────────────────────────────────────────

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4 py-10">
      <Card className="w-full max-w-lg shadow-sm">
        <CardHeader className="text-center pb-2">
          <CardTitle className="text-2xl font-bold">Register your company</CardTitle>
          <CardDescription>
            Fill in your company details. Submission creates a pending review in the Integrity Module.
          </CardDescription>
        </CardHeader>

        <CardContent>
          {/* Success */}
          {success && (
            <div className="mb-4 rounded-md bg-green-50 border border-green-200 px-4 py-3 text-sm text-green-800">
              Company registered! Redirecting…
            </div>
          )}

          {/* API error */}
          {apiError && (
            <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">
              {apiError}
            </div>
          )}

          <form onSubmit={handleSubmit} noValidate className="space-y-4">

            {/* Company name */}
            <div className="space-y-1.5">
              <Label htmlFor="name">Legal company name <span className="text-destructive">*</span></Label>
              <Input
                id="name"
                name="name"
                placeholder="Industrix Technologies LLP"
                value={fields.name}
                onChange={handleChange}
                className={fieldErrors.name ? "border-destructive" : ""}
              />
              {fieldErrors.name && <p className="text-xs text-destructive">{fieldErrors.name}</p>}
            </div>

            {/* BIN */}
            <div className="space-y-1.5">
              <Label htmlFor="bin">
                BIN — Business Identification Number <span className="text-destructive">*</span>
              </Label>
              <div className="relative">
                <Input
                  id="bin"
                  type="text"
                  inputMode="numeric"
                  placeholder="XXXX XXXX XXXX"
                  value={formatBIN(binDigits)}
                  onChange={handleBINChange}
                  maxLength={14}
                  className={
                    "font-mono tracking-widest pr-16 " +
                    (fieldErrors.bin ? "border-destructive" : "")
                  }
                />
                {/* Live counter */}
                <span className={
                  "absolute right-3 top-1/2 -translate-y-1/2 text-xs tabular-nums font-medium " +
                  (isBINValid(binDigits) ? "text-green-600" : "text-muted-foreground")
                }>
                  {binDigits.length}/12
                </span>
              </div>
              {fieldErrors.bin && <p className="text-xs text-destructive">{fieldErrors.bin}</p>}
            </div>

            {/* Email + Phone */}
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="email">Email <span className="text-destructive">*</span></Label>
                <Input
                  id="email"
                  name="email"
                  type="email"
                  placeholder="contact@company.kz"
                  value={fields.email}
                  onChange={handleChange}
                  className={fieldErrors.email ? "border-destructive" : ""}
                />
                {fieldErrors.email && <p className="text-xs text-destructive">{fieldErrors.email}</p>}
              </div>

              <div className="space-y-1.5">
                <Label htmlFor="phone">Phone <span className="text-destructive">*</span></Label>
                <Input
                  id="phone"
                  name="phone"
                  type="tel"
                  placeholder="+7 700 000 0000"
                  value={fields.phone}
                  onChange={handleChange}
                  className={fieldErrors.phone ? "border-destructive" : ""}
                />
                {fieldErrors.phone && <p className="text-xs text-destructive">{fieldErrors.phone}</p>}
              </div>
            </div>

            {/* Address */}
            <div className="space-y-1.5">
              <Label htmlFor="address">Legal address <span className="text-destructive">*</span></Label>
              <Input
                id="address"
                name="address"
                placeholder="Almaty, Bostandyk district…"
                value={fields.address}
                onChange={handleChange}
                className={fieldErrors.address ? "border-destructive" : ""}
              />
              {fieldErrors.address && <p className="text-xs text-destructive">{fieldErrors.address}</p>}
            </div>

            {/* Website (optional) */}
            <div className="space-y-1.5">
              <Label htmlFor="website">Website <span className="text-muted-foreground text-xs">(optional)</span></Label>
              <Input
                id="website"
                name="website"
                type="url"
                placeholder="https://company.kz"
                value={fields.website}
                onChange={handleChange}
              />
            </div>

            <div className="pt-2">
              <Button type="submit" className="w-full" disabled={loading || success}>
                {loading ? "Registering…" : "Register company"}
              </Button>
            </div>

          </form>
        </CardContent>
      </Card>
    </div>
  );
}
