"use client";

// src/app/account/company/page.tsx
// Shows the user's company + verification status, or the registration form if
// they don't have one yet. Selling is gated on having a company (see /shop/sell).

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { BadgeCheck, Clock, XCircle } from "lucide-react";
import { getMyCompany, type Company } from "@/lib/company";
import { getCurrentUser } from "@/lib/user";
import CompanyRegistrationForm from "@/components/CompanyRegistrationForm";
import { useI18n, type TranslationKey } from "@/lib/i18n";

// Holds keys rather than text: this map is built once at module load, so
// baking strings in here would freeze them to whatever locale was active first.
const STATUS_UI: Record<
  Company["status"],
  { icon: React.ReactNode; labelKey: TranslationKey; cls: string; noteKey: TranslationKey }
> = {
  pending: {
    icon: <Clock size={18} />,
    labelKey: "company.statusPending",
    cls: "bg-amber-50 border-amber-200 text-amber-800",
    noteKey: "company.notePending",
  },
  verified: {
    icon: <BadgeCheck size={18} />,
    labelKey: "company.statusVerified",
    cls: "bg-emerald-50 border-emerald-200 text-emerald-800",
    noteKey: "company.noteVerified",
  },
  rejected: {
    icon: <XCircle size={18} />,
    labelKey: "company.statusRejected",
    cls: "bg-rose-50 border-rose-200 text-rose-800",
    noteKey: "company.noteRejected",
  },
};

export default function CompanyPage() {
  const { t } = useI18n();
  const router = useRouter();
  const [company, setCompany] = useState<Company | null>(null);
  const [loading, setLoading] = useState(true);

  async function load() {
    const user = await getCurrentUser();
    if (!user) {
      router.push("/auth/login");
      return;
    }
    const c = await getMyCompany();
    setCompany(c);
    setLoading(false);
  }

  useEffect(() => { load(); /* eslint-disable-next-line */ }, []);

  if (loading) {
    return <div className="py-24 text-center text-gray-400">{t("common.loading")}</div>;
  }

  if (!company) {
    return <CompanyRegistrationForm onSuccess={load} />;
  }

  const ui = STATUS_UI[company.status];

  return (
    <div className="mx-auto max-w-lg px-4 py-10">
      <h1 className="mb-6 text-2xl font-extrabold text-gray-900">{t("company.title")}</h1>

      <div className={`mb-6 flex items-start gap-3 rounded-xl border px-4 py-3 ${ui.cls}`}>
        {ui.icon}
        <div>
          <p className="font-semibold">{t(ui.labelKey)}</p>
          <p className="mt-0.5 text-sm opacity-90">{t(ui.noteKey)}</p>
          {company.status === "rejected" && company.reviewer_note && (
            <p className="mt-2 text-sm font-medium">{t("company.reviewerNote")}: {company.reviewer_note}</p>
          )}
        </div>
      </div>

      <div className="space-y-3 rounded-xl border border-gray-200 p-5">
        <Row label={t("company.legalName")} value={company.name} />
        <Row label="BIN" value={company.bin} />
        <Row label={t("common.email")} value={company.email} />
        <Row label={t("common.phone")} value={company.phone} />
        <Row label={t("company.address")} value={company.address} />
        {company.website && <Row label={t("company.website")} value={company.website} />}
      </div>
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 border-b border-gray-100 pb-2 last:border-0">
      <span className="text-sm text-gray-500">{label}</span>
      <span className="text-right text-sm font-semibold text-gray-800">{value}</span>
    </div>
  );
}
