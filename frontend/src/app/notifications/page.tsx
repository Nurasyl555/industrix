"use client";

// src/app/notifications/page.tsx
// The current user's notification feed. Clicking an item marks it read and
// follows its link; "Mark all read" clears the bell badge.

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Bell, CheckCheck } from "lucide-react";
import {
  listNotifications,
  markRead,
  markAllRead,
  type Notification,
} from "@/lib/notification";
import { friendlyError } from "@/lib/api";
import { Button } from "@/components/ui/button";

function timeAgo(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

export default function NotificationsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setItems(await listNotifications());
  }

  useEffect(() => {
    load()
      .catch((err) => {
        if (friendlyError(err) === "Please sign in to continue.") {
          router.push("/auth/login");
          return;
        }
        setError(friendlyError(err));
      })
      .finally(() => setLoading(false));
  }, [router]);

  async function handleClick(n: Notification) {
    if (!n.read) {
      await markRead(n.id).catch(() => {});
      setItems((prev) => prev.map((x) => (x.id === n.id ? { ...x, read: true } : x)));
    }
    if (n.link) router.push(n.link);
  }

  async function handleMarkAll() {
    await markAllRead().catch(() => {});
    setItems((prev) => prev.map((x) => ({ ...x, read: true })));
  }

  const hasUnread = items.some((n) => !n.read);

  return (
    <div className="min-h-screen bg-white">
      <div className="mx-auto max-w-2xl px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-2xl font-extrabold text-gray-900">Notifications</h1>
          {hasUnread && (
            <Button variant="outline" size="sm" onClick={handleMarkAll}>
              <CheckCheck size={15} /> Mark all read
            </Button>
          )}
        </div>

        {error && <div className="mb-4 rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>}

        {loading ? (
          <div className="py-24 text-center text-[14px] text-gray-400">Loading…</div>
        ) : items.length === 0 ? (
          <div className="py-24 text-center text-[14px] text-gray-400">
            <Bell size={28} className="mx-auto mb-2 text-gray-300" />
            No notifications yet.
          </div>
        ) : (
          <div className="flex flex-col gap-2">
            {items.map((n) => (
              <button
                key={n.id}
                onClick={() => handleClick(n)}
                className={`flex items-start gap-3 rounded-xl border p-4 text-left transition-colors ${
                  n.read ? "border-gray-200 bg-white" : "border-blue-200 bg-blue-50/40"
                } hover:border-blue-300`}
              >
                {!n.read && <span className="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-blue-500" />}
                <div className={`min-w-0 flex-1 ${n.read ? "pl-5" : ""}`}>
                  <p className="text-sm text-gray-800">{n.message}</p>
                  <p className="mt-0.5 text-xs text-gray-400">{timeAgo(n.created_at)}</p>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
