// lib/review.ts
// Calls to the Marketplace Module (backend/modules/marketplace/) — reviews &
// reputation. An "entity" here is the party being reviewed — for the MVP that
// is the seller (a user id).

import { authGet, authPost } from "./api";

export interface Review {
  id: string;
  author_id: string;
  target_entity_id: string;
  rating: number;
  comment: string;
  created_at: string;
}

export interface Reputation {
  entity_id: string;
  average_rating: number;
  review_count: number;
  tier: "gold" | "silver" | "bronze" | "none";
}

interface ReviewPage {
  items: Review[] | null;
  total: number;
  page: number;
  limit: number;
}

/** List reviews for an entity (seller). Reviews are readable by any signed-in user. */
export async function listReviews(entityId: string): Promise<Review[]> {
  const res = await authGet<ReviewPage>(`/reviews/${entityId}`);
  return res.items ?? [];
}

/** Reputation for an entity, or null if it has no reviews yet (backend 404s). */
export async function getReputation(entityId: string): Promise<Reputation | null> {
  try {
    return await authGet<Reputation>(`/reviews/${entityId}/reputation`);
  } catch {
    return null;
  }
}

export function postReview(targetEntityId: string, rating: number, comment: string) {
  return authPost<Review>("/reviews", {
    target_entity_id: targetEntityId,
    rating,
    comment,
  });
}
