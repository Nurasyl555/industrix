// lib/catalog.ts
// Calls to the Catalog Module (backend/modules/catalog/)

import { publicGet, authPost, authPut, authDelete } from "./api";

export interface Category {
  id: string;
  name: string;
  slug: string;
  parent_id?: string;
}

export interface Equipment {
  id: string;
  owner_id: string;
  category_id: string;
  title: string;
  description: string;
  condition: "new" | "used";
  region: string;
  created_at: string;
  updated_at: string;
}

export interface CreateEquipmentInput {
  category_id: string;
  title: string;
  description: string;
  condition: "new" | "used";
  region: string;
}

export function listCategories() {
  return publicGet<Category[]>("/catalog/categories");
}

export function getEquipment(id: string) {
  return publicGet<Equipment>(`/catalog/equipment/${id}`);
}

export function createEquipment(input: CreateEquipmentInput) {
  return authPost<Equipment>("/catalog/equipment", input);
}

export function updateEquipment(id: string, input: Partial<CreateEquipmentInput>) {
  return authPut<Equipment>(`/catalog/equipment/${id}`, input);
}

export function deleteEquipment(id: string) {
  return authDelete<void>(`/catalog/equipment/${id}`);
}
