export type ViewMode   = "grid" | "list";
export type SortOption = "newest" | "price_asc" | "price_desc";

export interface Filters {
  categoryId:  string; // "" = all categories
  priceMin:    string;
  priceMax:    string;
  condition:   "" | "new" | "used";
  listingType: "" | "sale" | "rental";
}
