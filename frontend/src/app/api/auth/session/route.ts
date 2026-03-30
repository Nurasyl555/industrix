// src/app/api/auth/session/route.ts
import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

const COOKIE_OPTS = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax" as const,
  path: "/",
  maxAge: 60 * 60 * 24 * 7, // 7 days
};

// POST — set tokens after login/verify
export async function POST(req: NextRequest) {
  const { accessToken, refreshToken } = await req.json();
  const cookieStore = await cookies();
  cookieStore.set("access_token", accessToken, COOKIE_OPTS);
  cookieStore.set("refresh_token", refreshToken, {
    ...COOKIE_OPTS,
    maxAge: 60 * 60 * 24 * 30,
  });
  return NextResponse.json({ ok: true });
}

// DELETE — logout
export async function DELETE() {
  const cookieStore = await cookies();
  cookieStore.delete("access_token");
  cookieStore.delete("refresh_token");
  return NextResponse.json({ ok: true });
}
