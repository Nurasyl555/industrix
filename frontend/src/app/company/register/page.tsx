// src/app/company/register/page.tsx
// Server component — reads access_token cookie and passes it to the form
// If not logged in → redirect to login, then back here after

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import CompanyRegistrationForm from "@/components/CompanyRegistrationForm";

export default async function CompanyRegisterPage() {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;

  if (!accessToken) {
    redirect("/auth/login?next=/company/register");
  }

  return (
    <CompanyRegistrationForm
      accessToken={accessToken}
      successRedirectPath="/"
    />
  );
}
