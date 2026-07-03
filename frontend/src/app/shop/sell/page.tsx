"use client";

// src/app/shop/sell/page.tsx
// Seller flow: create equipment (Catalog module) then a listing for it
// (Listing module) in one form. Redirects to the new listing's detail page.

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { listCategories, createEquipment, type Category } from "@/lib/catalog";
import { createListing, publishListing, type ListingType, type PricePeriod } from "@/lib/listing";
import { friendlyError } from "@/lib/api";

export default function SellPage() {
  const router = useRouter();

  const [categories, setCategories] = useState<Category[]>([]);
  const [categoryId, setCategoryId] = useState("");
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [condition, setCondition] = useState<"new" | "used">("used");
  const [region, setRegion] = useState("");

  const [listingType, setListingType] = useState<ListingType>("sale");
  const [price, setPrice] = useState("");
  const [pricePeriod, setPricePeriod] = useState<PricePeriod>("day");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    listCategories().then(setCategories).catch(() => setCategories([]));
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!title.trim() || !categoryId || !price) {
      setError("Please fill in title, category, and price.");
      return;
    }

    setLoading(true);
    try {
      const equipment = await createEquipment({
        category_id: categoryId,
        title,
        description,
        condition,
        region,
      });

      const listing = await createListing({
        equipment_id: equipment.id,
        listing_type: listingType,
        price: Number(price),
        price_period: listingType === "rental" ? pricePeriod : undefined,
      });
      // New listings start as "draft" — publish immediately since there's
      // no moderation queue in this MVP yet.
      await publishListing(listing.id);

      router.push(`/shop/details?id=${listing.id}`);
    } catch (err) {
      if (friendlyError(err) === "Please sign in to continue.") {
        router.push("/auth/login");
        return;
      }
      setError(friendlyError(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4 py-10">
      <Card className="w-full max-w-lg shadow-sm">
        <CardHeader>
          <CardTitle className="text-2xl font-bold">List your equipment</CardTitle>
          <CardDescription>Note: this listing goes live immediately — moderation isn&apos;t built yet.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="title">Title</Label>
              <Input id="title" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="2018 Caterpillar 320 Excavator" />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="category">Category</Label>
              <select
                id="category"
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              >
                <option value="">Select a category…</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">Description</Label>
              <textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm"
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="condition">Condition</Label>
                <select
                  id="condition"
                  value={condition}
                  onChange={(e) => setCondition(e.target.value as "new" | "used")}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="used">Used</option>
                  <option value="new">New</option>
                </select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="region">Region</Label>
                <Input id="region" value={region} onChange={(e) => setRegion(e.target.value)} placeholder="Almaty" />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="listingType">Sale or Rent</Label>
                <select
                  id="listingType"
                  value={listingType}
                  onChange={(e) => setListingType(e.target.value as ListingType)}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="sale">For Sale</option>
                  <option value="rental">For Rent</option>
                </select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="price">Price ($)</Label>
                <Input id="price" type="number" min="0" value={price} onChange={(e) => setPrice(e.target.value)} />
              </div>
            </div>

            {listingType === "rental" && (
              <div className="space-y-1.5">
                <Label htmlFor="pricePeriod">Per</Label>
                <select
                  id="pricePeriod"
                  value={pricePeriod}
                  onChange={(e) => setPricePeriod(e.target.value as PricePeriod)}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="day">Day</option>
                  <option value="week">Week</option>
                  <option value="month">Month</option>
                </select>
              </div>
            )}

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Publishing…" : "Publish Listing"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
