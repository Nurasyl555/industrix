"use client";

// src/app/shop/deals/[id]/page.tsx
// Full-screen realtime conversation for a single deal. History loads over
// REST; new messages stream in over a WebSocket (see lib/deal.dealSocketURL).

import { useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Send } from "lucide-react";
import {
  getDeal,
  listDealMessages,
  postDealMessage,
  dealSocketURL,
  closeDeal,
  type Deal,
  type DealMessage,
} from "@/lib/deal";
import { getListing, type ListingView } from "@/lib/listing";
import { friendlyError } from "@/lib/api";
import { Button } from "@/components/ui/button";

function mySide(deal: Deal): string {
  return deal.role === "buyer" ? deal.buyer_id : deal.seller_id;
}

export default function DealChatPage() {
  const params = useParams();
  const router = useRouter();
  const dealId = String(params.id);

  const [deal, setDeal] = useState<Deal | null>(null);
  const [listing, setListing] = useState<ListingView | null>(null);
  const [messages, setMessages] = useState<DealMessage[]>([]);
  const [body, setBody] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [connected, setConnected] = useState(false);

  const wsRef = useRef<WebSocket | null>(null);
  const seen = useRef<Set<string>>(new Set());
  const bottomRef = useRef<HTMLDivElement | null>(null);

  function appendUnique(msg: DealMessage) {
    if (seen.current.has(msg.id)) return;
    seen.current.add(msg.id);
    setMessages((prev) => [...prev, msg]);
  }

  // Load deal + history, then open the socket.
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const d = await getDeal(dealId);
        if (cancelled) return;
        setDeal(d);
        getListing(d.listing_id).then((l) => !cancelled && setListing(l)).catch(() => {});

        const history = await listDealMessages(dealId);
        if (cancelled) return;
        history.forEach((m) => {
          seen.current.add(m.id);
        });
        setMessages(history);
      } catch (err) {
        if (cancelled) return;
        if (friendlyError(err) === "Please sign in to continue.") {
          router.push("/auth/login");
          return;
        }
        setError(friendlyError(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => { cancelled = true; };
  }, [dealId, router]);

  // Open WebSocket after the deal is confirmed (so we know we're a participant).
  useEffect(() => {
    if (!deal) return;
    const ws = new WebSocket(dealSocketURL(dealId));
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onmessage = (e) => {
      try {
        appendUnique(JSON.parse(e.data) as DealMessage);
      } catch { /* ignore malformed frame */ }
    };

    return () => { ws.close(); };
  }, [deal, dealId]);

  // Auto-scroll to newest.
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSend() {
    const text = body.trim();
    if (!text || !deal || deal.status === "closed") return;
    setBody("");

    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      // Server echoes it back over the socket → appendUnique renders it once.
      ws.send(JSON.stringify({ body: text }));
    } else {
      // Socket not up — fall back to REST and render the returned message.
      try {
        appendUnique(await postDealMessage(dealId, text));
      } catch (err) {
        setError(friendlyError(err));
        setBody(text);
      }
    }
  }

  async function handleClose() {
    if (!deal) return;
    try {
      await closeDeal(dealId);
      setDeal({ ...deal, status: "closed" });
    } catch (err) {
      setError(friendlyError(err));
    }
  }

  const me = deal ? mySide(deal) : "";

  return (
    <div className="mx-auto flex h-[calc(100vh-4rem)] max-w-3xl flex-col px-4 py-4">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-200 pb-3">
        <div className="flex items-center gap-3 min-w-0">
          <Link href="/shop/deals" className="text-gray-500 hover:text-gray-900">
            <ArrowLeft size={20} />
          </Link>
          <div className="min-w-0">
            <p className="truncate text-[15px] font-bold text-gray-900">
              {listing ? listing.title : "Conversation"}
            </p>
            <p className="text-xs text-gray-500">
              {deal && (deal.role === "buyer" ? "You inquired · " : "Inquiry received · ")}
              <span className={connected ? "text-emerald-600" : "text-gray-400"}>
                {connected ? "● live" : "○ offline"}
              </span>
              {deal?.status === "closed" && " · closed"}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {listing && (
            <Button asChild variant="outline" size="sm">
              <Link href={`/shop/details?id=${deal?.listing_id}`}>Listing</Link>
            </Button>
          )}
          {deal?.status === "inquiry" && (
            <Button variant="outline" size="sm" onClick={handleClose}>Close</Button>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto py-4">
        {error && <div className="mb-3 rounded-md bg-destructive/10 px-4 py-2 text-sm text-destructive">{error}</div>}
        {loading ? (
          <div className="py-24 text-center text-sm text-gray-400">Loading…</div>
        ) : messages.length === 0 ? (
          <div className="py-24 text-center text-sm text-gray-400">No messages yet. Say hello 👋</div>
        ) : (
          <div className="flex flex-col gap-2">
            {messages.map((m) => {
              const isMine = m.sender_id === me;
              return (
                <div key={m.id} className={`flex ${isMine ? "justify-end" : "justify-start"}`}>
                  <div
                    className={`max-w-[75%] rounded-2xl px-3.5 py-2 text-sm ${
                      isMine ? "bg-blue-600 text-white" : "bg-gray-100 text-gray-800"
                    }`}
                  >
                    {m.body}
                  </div>
                </div>
              );
            })}
            <div ref={bottomRef} />
          </div>
        )}
      </div>

      {/* Composer */}
      {deal?.status === "closed" ? (
        <div className="border-t border-gray-200 py-4 text-center text-sm text-gray-400">
          This deal is closed — replies are disabled.
        </div>
      ) : (
        <div className="flex items-center gap-2 border-t border-gray-200 pt-3">
          <input
            value={body}
            onChange={(e) => setBody(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") handleSend(); }}
            placeholder="Type a message…"
            className="flex-1 rounded-full border border-gray-200 px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <Button onClick={handleSend} disabled={!body.trim()} className="rounded-full h-10 w-10 p-0">
            <Send size={16} />
          </Button>
        </div>
      )}
    </div>
  );
}
