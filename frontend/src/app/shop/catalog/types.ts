export type ViewMode   = "grid" | "list";
export type SortOption = "relevance" | "price-asc" | "price-desc" | "date-asc" | "date-desc" | "hours-asc" | "hours-desc";

export interface Filters {
  categories: string[];
  priceMin:   string;
  priceMax:   string;
  conditions: string[];
  yearMin:    number;
  yearMax:    number;
}

export interface Equipment {
  id:        number;
  title:     string;
  price:     number;
  year:      number;
  hours:     number;
  location:  string;
  condition: "New" | "Used";
  badge?:    "VERIFIED" | "NEW ARRIVAL";
  image:     string; // sky colour seed key
}
