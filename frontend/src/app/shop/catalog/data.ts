import { type Equipment, type SortOption, type Filters } from "@/types";

export const ALL_EQUIPMENT: Equipment[] = [
  { id: 1,  title: "2018 Caterpillar 320 GC Hydraulic...", price: 148000, year: 2018, hours: 3420, location: "Houston, TX",  condition: "Used", badge: "VERIFIED",    image: "0" },
  { id: 2,  title: "2015 Komatsu PC210LC-10...",           price: 92500,  year: 2015, hours: 6850, location: "Chicago, IL",  condition: "Used", badge: "NEW ARRIVAL", image: "1" },
  { id: 3,  title: "2020 John Deere 350G LC Excavator",   price: 215000, year: 2020, hours: 2100, location: "Phoenix, AZ",  condition: "Used",                        image: "2" },
  { id: 4,  title: "2012 Volvo EC220DL Crawler Excavator",price: 64000,  year: 2012, hours: 9200, location: "Atlanta, GA",  condition: "Used",                        image: "3" },
  { id: 5,  title: "2019 Doosan DX225LC-5...",             price: 126500, year: 2019, hours: 4150, location: "Denver, CO",   condition: "Used",                        image: "4" },
  { id: 6,  title: "2021 Kubota KX040-4 Mini Excavator",  price: 58900,  year: 2021, hours: 850,  location: "Seattle, WA",  condition: "Used",                        image: "5" },
  { id: 7,  title: "2017 Hitachi ZX350LC-6 Excavator",    price: 178000, year: 2017, hours: 4800, location: "Dallas, TX",   condition: "Used", badge: "VERIFIED",    image: "0" },
  { id: 8,  title: "2022 Caterpillar 308 Mini Excavator", price: 89000,  year: 2022, hours: 420,  location: "Miami, FL",    condition: "New",  badge: "NEW ARRIVAL", image: "1" },
  { id: 9,  title: "2016 Liebherr R 920 Excavator",       price: 112000, year: 2016, hours: 7200, location: "Portland, OR", condition: "Used",                        image: "2" },
  { id: 10, title: "2023 Bobcat E60 Compact Excavator",   price: 67500,  year: 2023, hours: 180,  location: "Austin, TX",   condition: "New",  badge: "NEW ARRIVAL", image: "3" },
  { id: 11, title: "2014 Hyundai R220LC-9A Excavator",    price: 74000,  year: 2014, hours: 8900, location: "Cleveland, OH", condition: "Used",                       image: "4" },
  { id: 12, title: "2020 Takeuchi TB2150 Excavator",      price: 155000, year: 2020, hours: 1600, location: "Nashville, TN", condition: "Used", badge: "VERIFIED",   image: "5" },
];

export function applyFiltersAndSort(
  items: Equipment[],
  filters: Filters,
  sort: SortOption
): Equipment[] {
  let result = [...items];

  if (filters.categories.length > 0) {
    // In a real app, each item would have a category field.
    // Here all mock items are excavators — filter is a no-op for demo.
  }

  if (filters.conditions.length > 0) {
    result = result.filter((i) => filters.conditions.includes(i.condition));
  }

  if (filters.priceMin !== "") {
    result = result.filter((i) => i.price >= Number(filters.priceMin));
  }
  if (filters.priceMax !== "") {
    result = result.filter((i) => i.price <= Number(filters.priceMax));
  }

  result = result.filter(
    (i) => i.year >= filters.yearMin && i.year <= filters.yearMax
  );

  switch (sort) {
    case "price-asc":  result.sort((a, b) => a.price - b.price);  break;
    case "price-desc": result.sort((a, b) => b.price - a.price);  break;
    case "date-asc":   result.sort((a, b) => a.year  - b.year);   break;
    case "date-desc":  result.sort((a, b) => b.year  - a.year);   break;
    case "hours-asc":  result.sort((a, b) => a.hours - b.hours);  break;
    case "hours-desc": result.sort((a, b) => b.hours - a.hours);  break;
    default: break; // relevance — original order
  }

  return result;
}

export function paginate<T>(items: T[], page: number, perPage: number): T[] {
  const start = (page - 1) * perPage;
  return items.slice(start, start + perPage);
}
