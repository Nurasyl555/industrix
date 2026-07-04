// lib/notification.ts
// In-app notification feed (backend: notification module).

import { authGet, authPut } from "./api";

export interface Notification {
  id: string;
  user_id: string;
  type: string;
  message: string;
  link: string;
  read: boolean;
  created_at: string;
}

export async function listNotifications(): Promise<Notification[]> {
  return (await authGet<Notification[] | null>("/notifications")) ?? [];
}

export async function unreadCount(): Promise<number> {
  try {
    const res = await authGet<{ count: number }>("/notifications/unread-count");
    return res.count;
  } catch {
    return 0; // not signed in → nothing to show
  }
}

export function markRead(id: string) {
  return authPut<void>(`/notifications/${id}/read`);
}

export function markAllRead() {
  return authPut<void>("/notifications/read-all");
}
