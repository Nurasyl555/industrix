import { create } from "zustand";
import { persist } from "zustand/middleware";
import { type Equipment } from "@/app/shop/catalog/types";

interface FavoritesStore {
  items:     Equipment[];
  add:       (item: Equipment) => void;
  remove:    (id: number)      => void;
  toggle:    (item: Equipment) => void;
  isFav:     (id: number)      => boolean;
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
