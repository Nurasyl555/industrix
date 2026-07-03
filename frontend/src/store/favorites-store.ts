import { create } from "zustand";
import { persist } from "zustand/middleware";
import { type ListingView } from "@/lib/listing";

interface FavoritesStore {
  items:  ListingView[];
  add:    (item: ListingView) => void;
  remove: (id: string) => void;
  toggle: (item: ListingView) => void;
  isFav:  (id: string) => boolean;
}

export const useFavorites = create<FavoritesStore>()(
  persist(
    (set, get) => ({
      items: [],

      add: (item) =>
        set((s) => ({ items: [...s.items.filter((i) => i.id !== item.id), item] })),

      remove: (id) =>
        set((s) => ({ items: s.items.filter((i) => i.id !== id) })),

      toggle: (item) =>
        get().isFav(item.id) ? get().remove(item.id) : get().add(item),

      isFav: (id) => get().items.some((i) => i.id === id),
    }),
    { name: "industrix-favorites" }
  )
);
