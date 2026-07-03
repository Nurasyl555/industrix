// src/app/api/proxy/[...path]/route.ts
//
// Same-origin proxy for authenticated backend calls. The access_token lives
// in an httpOnly cookie (set by /api/auth/session) precisely so client JS
// can't read it — that's the point, it blocks XSS token theft. This route
// runs on the server, reads the cookie, and forwards it as a Bearer token
// to the Go backend. Client code should call /api/proxy/<path> instead of
// the backend directly for anything that requires auth.
import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

const API = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1").replace(/\/$/, "");

async function forward(req: NextRequest, method: string, path: string[]) {
  const cookieStore = await cookies();
  const token = cookieStore.get("access_token")?.value;

  if (!token) {
    return NextResponse.json({ code: "UNAUTHORIZED", message: "Not signed in" }, { status: 401 });
  }

  const url = `${API}/${path.join("/")}${req.nextUrl.search}`;
  const hasBody = method !== "GET" && method !== "DELETE";

  const res = await fetch(url, {
    method,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: hasBody ? await req.text() : undefined,
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": res.headers.get("Content-Type") ?? "application/json" },
  });
}

type Ctx = { params: Promise<{ path: string[] }> };

export async function GET(req: NextRequest, ctx: Ctx) {
  return forward(req, "GET", (await ctx.params).path);
}
export async function POST(req: NextRequest, ctx: Ctx) {
  return forward(req, "POST", (await ctx.params).path);
}
export async function PUT(req: NextRequest, ctx: Ctx) {
  return forward(req, "PUT", (await ctx.params).path);
}
export async function DELETE(req: NextRequest, ctx: Ctx) {
  return forward(req, "DELETE", (await ctx.params).path);
}
