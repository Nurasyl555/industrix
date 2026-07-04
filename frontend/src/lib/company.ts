// lib/company.ts
// Calls to the Integrity Module (backend/modules/integrity/) — companies &
// verification. All routed through the authenticated proxy.

import { authGet, authPost } from "./api";

export type CompanyStatus = "pending" | "verified" | "rejected";

export interface Company {
  id: string;
  owner_id: string;
  name: string;
  bin: string;
  address: string;
  phone: string;
  email: string;
  website: string;
  status: CompanyStatus;
  verified: boolean;
  reviewer_note: string;
  created_at: string;
}

export interface CreateCompanyInput {
  name: string;
  bin: string;
  address: string;
  phone: string;
  email: string;
  website: string;
}

/** The current user's company, or null if they haven't registered one (404). */
export async function getMyCompany(): Promise<Company | null> {
  try {
    return await authGet<Company>("/my-company");
  } catch {
    return null;
  }
}

export function createCompany(input: CreateCompanyInput) {
  return authPost<Company>("/companies", input);
}
