import {
  Heart, Share2, ArrowLeftRight,
  Mail, ChevronDown, ChevronLeft, ChevronRight,
  ShieldCheck, Truck,
} from "lucide-react";

export default function IndustrixProductPage() {
  const gallery = [
  "/pics/sample.jpg",
  "/pics/sample-1.jpeg",
  "/pics/sample-2.jpeg",
  "/pics/sample-3.jpeg",
  "/pics/sample-4.jpg",
  ];

  const specs = [
    { label: "Year",                 value: "2021" },
    { label: "Hours Used",           value: "2,450 hrs" },
    { label: "Engine Model",         value: "Cat C4.4 ACERT" },
    { label: "Net Power - ISO 9249", value: "107 kW / 143 hp" },
    { label: "Operating Weight",     value: "21,900 kg" },
    { label: "Bucket Capacity",      value: "1.2 m³" },
    { label: "Max Digging Depth",    value: "6.72 m" },
    { label: "Condition",            value: "EXCELLENT", badge: true },
  ];

  const similarEquipment = [
    { name: "Komatsu PC210LC-11", year: "2019", price: "$89,000",  location: "Taraz, KZ",    image: "/pics/sample.jpg" },
    { name: "Deere 210G LC",      year: "2020", price: "$112,500", location: "Astana, KZ",   image: "/pics/sample.jpg" },
    { name: "Volvo EC220E",       year: "2018", price: "$76,000",  location: "Karaganda, KZ",image: "/pics/sample.jpg" },
  ];

  const calendarDays: [number, string][] = [
    [26, "neutral"], [27, "neutral"], [28, "neutral"], [29, "neutral"], [30, "neutral"],
    [1, "free"],  [2, "free"],
    [3, "booked"], [4, "booked"], [5, "booked"],
    [6, "selected"],
    [7, "free"],  [8, "free"], [9, "free"],
  ];

  return (
    <main className="min-h-screen bg-slate-50 text-slate-900">
      <div className="mx-auto max-w-7xl px-4 pb-16 pt-6 sm:px-6 lg:px-8">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px] xl:grid-cols-[minmax(0,1fr)_380px]">

          {/* ── LEFT COLUMN ── */}
          <section className="space-y-6">

            {/* Gallery */}
            <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="relative aspect-video w-full bg-linear-to-br from-sky-300 via-sky-500 to-sky-700">
                <img
                  src={gallery[0]}
                  alt="Caterpillar 320 GC Hydraulic Excavator"
                  className="h-full w-full object-cover"
                />
                <div className="absolute bottom-4 left-4 flex flex-wrap gap-2">
                  <span className="rounded-md bg-black/70 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
                    Video Available
                  </span>
                  <span className="rounded-md bg-sky-600 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-white">
                    Verified Listing
                  </span>
                </div>
              </div>
            </div>

            {/* Thumbnails */}
            <div className="flex flex-wrap gap-3">
              {gallery.slice(1).map((image, index) => (
                <button
                  key={index}
                  className="flex h-20 w-20 items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm transition hover:border-sky-500"
                >
                  <img src={image} alt={`Thumbnail ${index + 1}`} className="h-full w-full object-cover" />
                </button>
              ))}
              <button className="flex h-20 w-20 items-center justify-center rounded-xl border border-slate-200 bg-slate-100 text-sm font-semibold text-slate-500 shadow-sm hover:border-sky-500 transition">
                +12
              </button>
            </div>

            {/* Title + actions */}
            <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <div>
                <p className="mb-2 text-sm font-semibold text-sky-700">Caterpillar Inc.</p>
                <h1 className="text-3xl font-bold tracking-tight text-slate-900 sm:text-4xl">
                  Caterpillar 320 GC Hydraulic Excavator
                </h1>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {[
                  { icon: <Heart size={16} />,         label: "Save" },
                  { icon: <Share2 size={16} />,        label: "Share" },
                  { icon: <ArrowLeftRight size={16} />,label: "Compare" },
                ].map(({ icon, label }) => (
                  <button
                    key={label}
                    aria-label={label}
                    className="flex h-9 w-9 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 shadow-sm transition hover:border-slate-300 hover:text-slate-800"
                  >
                    {icon}
                  </button>
                ))}
              </div>
            </div>

            {/* Specs */}
            <section className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
              <div className="border-b border-slate-200 px-6 py-5">
                <h2 className="text-xl font-semibold text-slate-900">Technical Specifications</h2>
              </div>
              <div className="grid grid-cols-1 gap-x-10 gap-y-6 px-6 py-6 sm:grid-cols-2">
                {specs.map((item) => (
                  <div key={item.label} className="flex items-center justify-between gap-4 border-b border-slate-100 pb-4 last:border-b-0">
                    <span className="text-sm text-slate-500">{item.label}</span>
                    {item.badge ? (
                      <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold tracking-wide text-emerald-700">
                        {item.value}
                      </span>
                    ) : (
                      <span className="text-right text-sm font-semibold text-slate-800">{item.value}</span>
                    )}
                  </div>
                ))}
              </div>
              <div className="px-6 pb-6 text-center">
                <button className="inline-flex items-center gap-1 text-sm font-medium text-slate-500 transition hover:text-sky-700">
                  Show All Specifications <ChevronDown size={14} />
                </button>
              </div>
            </section>

            {/* Description */}
            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-slate-900">Description</h2>
              <p className="max-w-4xl text-base leading-8 text-slate-600">
                This Caterpillar 320 GC hydraulic excavator is in excellent condition and has been regularly
                serviced at authorized dealerships. It features a high-efficiency hydraulic system that reduces
                fuel consumption by up to 20%. Includes standard bucket and quick-coupler. Recent maintenance
                report available upon request. Perfectly suited for light to medium-duty applications.
              </p>
            </section>

            {/* Similar Equipment */}
            <section className="space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-2xl font-semibold text-slate-900">Similar Equipment</h2>
                <div className="flex items-center gap-2">
                  {[{ icon: <ChevronLeft size={16} />, label: "Previous" }, { icon: <ChevronRight size={16} />, label: "Next" }].map(({ icon, label }) => (
                    <button
                      key={label}
                      aria-label={label}
                      className="flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:border-slate-300 hover:text-slate-800"
                    >
                      {icon}
                    </button>
                  ))}
                </div>
              </div>
              <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {similarEquipment.map((item) => (
                  <article
                    key={item.name}
                    className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm transition hover:-translate-y-0.5 hover:shadow-md"
                  >
                    <div className="aspect-16/10 w-full overflow-hidden bg-slate-200">
                      <img src={item.image} alt={item.name} className="h-full w-full object-cover" />
                    </div>
                    <div className="space-y-2 p-4">
                      <h3 className="text-base font-semibold text-slate-900">{item.name}</h3>
                      <p className="text-sm text-slate-500">{item.year}</p>
                      <div className="flex items-center justify-between gap-4">
                        <p className="text-2xl font-bold text-sky-600">{item.price}</p>
                        <p className="text-sm text-slate-400">{item.location}</p>
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          </section>

          {/* ── RIGHT COLUMN ── */}
          <aside className="space-y-4">
            <section className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">

              {/* Price + CTAs */}
              <div className="space-y-4">
                <h2 className="text-4xl font-bold tracking-tight text-slate-900">$124,500</h2>
                <div className="space-y-3">
                  <button className="flex h-12 w-full items-center justify-center gap-2 rounded-xl bg-sky-600 text-sm font-semibold text-white transition hover:bg-sky-700">
                    <Mail size={16} /> Contact Seller
                  </button>
                  <button className="flex h-12 w-full items-center justify-center rounded-xl border border-sky-600 bg-white text-sm font-semibold text-sky-700 transition hover:bg-sky-50">
                    Make Offer
                  </button>
                </div>
                <div className="flex items-center gap-3 py-1">
                  <div className="h-px flex-1 bg-slate-200" />
                  <span className="text-xs font-medium uppercase tracking-[0.22em] text-slate-400">Or Rent For</span>
                  <div className="h-px flex-1 bg-slate-200" />
                </div>
                <button className="flex h-12 w-full items-center justify-center rounded-xl bg-slate-950 text-sm font-semibold text-white transition hover:bg-slate-800">
                  Rent $1,200 / day
                </button>
              </div>

              {/* Calendar */}
              <div className="mt-8 space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-slate-900">Rental Availability</h3>
                  <div className="flex items-center gap-3 text-xs font-medium">
                    <span className="flex items-center gap-1 text-emerald-600">
                      <span className="h-2 w-2 rounded-full bg-emerald-500" /> FREE
                    </span>
                    <span className="flex items-center gap-1 text-rose-500">
                      <span className="h-2 w-2 rounded-full bg-rose-400" /> BOOKED
                    </span>
                  </div>
                </div>
                <div className="rounded-2xl bg-slate-50 p-4">
                  <div className="mb-3 grid grid-cols-7 text-center text-xs font-medium uppercase text-slate-400">
                    {["M","T","W","T","F","S","S"].map((d, i) => (
                      <span key={i}>{d}</span>
                    ))}
                  </div>
                  <div className="grid grid-cols-7 gap-2">
                    {calendarDays.map(([day, status], i) => (
                      <div
                        key={i}
                        className={[
                          "flex aspect-square items-center justify-center rounded-lg text-sm font-medium",
                          status === "neutral"  && "bg-slate-100 text-slate-400",
                          status === "free"     && "bg-emerald-50 text-emerald-700",
                          status === "booked"   && "bg-rose-50 text-rose-600",
                          status === "selected" && "border-2 border-sky-500 bg-white text-sky-700",
                        ].filter(Boolean).join(" ")}
                      >
                        {day}
                      </div>
                    ))}
                  </div>
                  <p className="mt-3 text-center text-xs text-slate-400">Minimum rental period: 3 days</p>
                </div>
              </div>

              {/* Seller */}
              <div className="mt-8 space-y-4">
                <h3 className="text-sm font-semibold text-slate-900">Seller Information</h3>
                <div className="flex gap-3 rounded-2xl border border-slate-200 p-4">
                  <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-slate-100 text-slate-400">
                    <Truck size={20} />
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-1">
                      <p className="truncate font-semibold text-slate-900">Titan Heavy Equip</p>
                      <span className="text-sky-600 text-sm">✓</span>
                    </div>
                    <p className="text-sm text-amber-500">
                      ★★★★★ <span className="text-slate-500">4.8 (124)</span>
                    </p>
                  </div>
                </div>
                <dl className="space-y-3 text-sm">
                  {[
                    { label: "Response Rate", value: "98% (under 2h)", green: true },
                    { label: "Member Since",  value: "April 2019" },
                    { label: "Location",      value: "Almaty, KZ" },
                  ].map(({ label, value, green }) => (
                    <div key={label} className="flex items-center justify-between gap-4">
                      <dt className="text-slate-500">{label}</dt>
                      <dd className={`font-semibold ${green ? "text-emerald-600" : "text-slate-800"}`}>{value}</dd>
                    </div>
                  ))}
                </dl>
                <button className="flex h-11 w-full items-center justify-center rounded-xl border border-slate-200 bg-white text-sm font-semibold text-slate-700 transition hover:bg-slate-50">
                  View Seller Profile
                </button>
              </div>
            </section>

            {/* Trade safe */}
            <section className="rounded-2xl border border-amber-200 bg-amber-50 p-4 shadow-sm">
              <div className="flex items-start gap-3">
                <ShieldCheck size={18} className="mt-0.5 shrink-0 text-amber-500" />
                <div>
                  <h3 className="text-sm font-semibold text-amber-900">Trade Safely</h3>
                  <p className="mt-1 text-sm leading-6 text-amber-800/80">
                    Always inspect machinery in person or via a third-party inspector before final payment.
                    Use our secure escrow service for transactions.
                  </p>
                </div>
              </div>
            </section>
          </aside>

        </div>
      </div>
    </main>
  );
}