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
import { useI18n } from "@/lib/i18n";
import { getMyCompany } from "@/lib/company";
import { getCurrentUser } from "@/lib/user";
import { uploadImage } from "@/lib/media";
import { friendlyError } from "@/lib/api";
import Link from "next/link";
import Image from "next/image";

export default function SellPage() {
  const { t } = useI18n();
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

  const [imageUrl, setImageUrl] = useState("");
  const [uploading, setUploading] = useState(false);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  // Gate: selling requires a registered company (verification workflow).
  const [gate, setGate] = useState<"checking" | "no-company" | "ok">("checking");

  useEffect(() => {
    (async () => {
      const user = await getCurrentUser();
      if (!user) { router.push("/auth/login"); return; }
      const company = await getMyCompany();
      setGate(company ? "ok" : "no-company");
    })();
    listCategories().then(setCategories).catch(() => setCategories([]));
  }, [router]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!title.trim() || !categoryId || !price) {
      setError(t("sell.requiredFields"));
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
        image_url: imageUrl || undefined,
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

  if (gate === "checking") {
    return <div className="py-24 text-center text-gray-400">Loading…</div>;
  }

  if (gate === "no-company") {
    return (
      <div className="mx-auto max-w-lg px-4 py-16 text-center">
        <Card className="shadow-sm">
          <CardHeader>
            <CardTitle className="text-2xl font-bold">{t("sell.needCompanyTitle")}</CardTitle>
            <CardDescription>
              {t("sell.needCompanyText")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild className="w-full">
              <Link href="/account/company">{t("sell.registerCompany")}</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4 py-10">
      <Card className="w-full max-w-lg shadow-sm">
        <CardHeader>
          <CardTitle className="text-2xl font-bold">{t("sell.formTitle")}</CardTitle>
          <CardDescription>{t("sell.formSubtitle")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-4 py-3 text-sm text-destructive">{error}</div>
            )}

            <div className="space-y-1.5">
              <Label htmlFor="title">{t("sell.fieldTitle")}</Label>
              <Input id="title" value={title} onChange={(e) => setTitle(e.target.value)} placeholder={t("sell.titlePlaceholder")} />
            </div>

            {/* Photo */}
            <div className="space-y-1.5">
              <Label htmlFor="photo">{t("sell.photo")}</Label>
              <div className="flex items-center gap-4">
                <div className="relative h-24 w-32 shrink-0 overflow-hidden rounded-lg border border-dashed border-gray-300 bg-gray-50">
                  {imageUrl ? (
                    <Image src={imageUrl} alt="preview" fill className="object-cover" unoptimized />
                  ) : (
                    <div className="flex h-full items-center justify-center text-xs text-gray-400">No photo</div>
                  )}
                </div>
                <div className="space-y-1">
                  <input
                    id="photo"
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    disabled={uploading}
                    onChange={async (e) => {
                      const file = e.target.files?.[0];
                      if (!file) return;
                      setError("");
                      setUploading(true);
                      try {
                        setImageUrl(await uploadImage(file));
                      } catch (err) {
                        setError(friendlyError(err));
                      } finally {
                        setUploading(false);
                      }
                    }}
                    className="text-sm"
                  />
                  {uploading && <p className="text-xs text-gray-400">Uploading…</p>}
                </div>
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="category">{t("sell.category")}</Label>
              <select
                id="category"
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
              >
                <option value="">{t("sell.selectCategory")}</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">{t("sell.description")}</Label>
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
                <Label htmlFor="condition">{t("filters.condition")}</Label>
                <select
                  id="condition"
                  value={condition}
                  onChange={(e) => setCondition(e.target.value as "new" | "used")}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="used">{t("condition.used")}</option>
                  <option value="new">{t("condition.new")}</option>
                </select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="region">{t("sell.region")}</Label>
                <Input id="region" value={region} onChange={(e) => setRegion(e.target.value)} placeholder={t("sell.regionPlaceholder")} />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="listingType">{t("filters.saleOrRent")}</Label>
                <select
                  id="listingType"
                  value={listingType}
                  onChange={(e) => setListingType(e.target.value as ListingType)}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="sale">{t("listingType.sale")}</option>
                  <option value="rental">{t("listingType.rental")}</option>
                </select>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="price">{t("sell.price")}</Label>
                <Input id="price" type="number" min="0" value={price} onChange={(e) => setPrice(e.target.value)} />
              </div>
            </div>

            {listingType === "rental" && (
              <div className="space-y-1.5">
                <Label htmlFor="pricePeriod">{t("sell.per")}</Label>
                <select
                  id="pricePeriod"
                  value={pricePeriod}
                  onChange={(e) => setPricePeriod(e.target.value as PricePeriod)}
                  className="w-full h-9 rounded-md border border-input bg-transparent px-3 text-sm"
                >
                  <option value="day">{t("period.day")}</option>
                  <option value="week">{t("period.week")}</option>
                  <option value="month">{t("period.month")}</option>
                </select>
              </div>
            )}

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? t("sell.publishing") : t("sell.publish")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
