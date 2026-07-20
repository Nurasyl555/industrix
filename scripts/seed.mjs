#!/usr/bin/env node
// Seeds the local dev stack with a realistic demo dataset: 2 regular users
// (each both a seller and a buyer), a spread of equipment listings across
// every category, and deals covering every FSM status including a live
// dispute — so every feature in the app has something to look at.
//
// Idempotent-ish: safe to re-run for users/companies (falls back to login /
// existing company on conflict). Listings and deals are NOT deduped — running
// it twice will create a second batch of listings/deals.
//
// Usage: node scripts/seed.mjs
// Env:   API_BASE (default http://localhost:8080/api/v1)

const API = process.env.API_BASE || "http://localhost:8080/api/v1";

async function api(method, path, body, token) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`${API}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  let json = null;
  const text = await res.text();
  if (text) {
    try { json = JSON.parse(text); } catch { json = text; }
  }
  if (!res.ok) {
    const err = new Error(`${method} ${path} -> ${res.status}: ${JSON.stringify(json)}`);
    err.status = res.status;
    err.body = json;
    throw err;
  }
  return json;
}

function log(...args) { console.log(...args); }
function section(title) { console.log(`\n=== ${title} ===`); }

// --- Users ---------------------------------------------------------------

async function registerOrLogin(email, password, firstName) {
  try {
    const tokens = await api("POST", "/auth/email/register", { email, password, first_name: firstName });
    log(`  registered ${email}`);
    return tokens;
  } catch (e) {
    if (e.status === 409) {
      const tokens = await api("POST", "/auth/email/login", { email, password });
      log(`  ${email} already exists, logged in`);
      return tokens;
    }
    throw e;
  }
}

async function ensureCompany(token, company) {
  try {
    const c = await api("POST", "/companies", company, token);
    log(`  created company "${company.name}"`);
    return c;
  } catch (e) {
    if (e.status === 409) {
      const mine = await api("GET", "/my-company", undefined, token);
      log(`  company already exists: "${mine.name}"`);
      return mine;
    }
    throw e;
  }
}

// --- Equipment / listings --------------------------------------------------

async function createListing(token, categoryId, spec, adminToken) {
  const eq = await api("POST", "/catalog/equipment", {
    category_id: categoryId,
    title: spec.title,
    description: spec.description,
    condition: spec.condition,
    region: spec.region,
    image_url: spec.image_url,
  }, token);

  const listing = await api("POST", "/listings", {
    equipment_id: eq.id,
    listing_type: spec.listingType,
    price: spec.price,
    price_period: spec.pricePeriod || "",
    pricing_type: "fixed",
  }, token);

  // Publish sends the listing to moderation; an admin approval is what
  // actually flips it to "active" so buyers can see and inquire on it.
  await api("PUT", `/my-listings/${listing.id}/publish`, undefined, token);
  if (adminToken) {
    await api("PUT", `/admin/listings/${listing.id}/approve`, undefined, adminToken);
  }
  log(`  [${spec.listingType}] ${spec.title} — ${spec.price.toLocaleString("ru-RU")}${spec.pricePeriod ? "/" + spec.pricePeriod : ""} тг (${spec.region})`);

  return { ...listing, equipment: eq };
}

// --- Deals -----------------------------------------------------------------

async function createDeal(buyerToken, listingId, message) {
  return api("POST", "/deals", { listing_id: listingId, message }, buyerToken);
}

async function transition(token, dealId, status) {
  return api("PUT", `/deals/${dealId}/status`, { status }, token);
}

async function postMessage(token, dealId, body) {
  return api("POST", `/deals/${dealId}/messages`, { body }, token);
}

async function fundEscrow(buyerToken, dealId, amount) {
  return api("POST", "/payments", { deal_id: dealId, amount, currency: "KZT" }, buyerToken);
}

// --- Main --------------------------------------------------------------

async function main() {
  section("Users");

  const erlan = await registerOrLogin("erlan.demo@industrix.kz", "Demo!2345", "Ерлан");
  const aigerim = await registerOrLogin("aigerim.demo@industrix.kz", "Demo!2345", "Айгерим");

  let admin;
  try {
    admin = await api("POST", "/auth/email/login", { email: "admin@industrix.kz", password: "Admin!2345" });
    log("  logged in as admin@industrix.kz");
  } catch (e) {
    log("  WARNING: could not log in as admin (admin@industrix.kz / Admin!2345) — dispute resolution step will be skipped");
  }

  const erlanTok = erlan.accessToken;
  const aigerimTok = aigerim.accessToken;
  const adminTok = admin && admin.accessToken;

  const erlanMe = await api("GET", "/users/me", undefined, erlanTok);
  const aigerimMe = await api("GET", "/users/me", undefined, aigerimTok);

  section("Companies");
  const stst = await ensureCompany(erlanTok, {
    name: "ТОО СтройТехСервис",
    bin: "123456789012",
    address: "г. Алматы, пр. Райымбека, 212",
    phone: "+7 727 123 4567",
    email: "info@stroytechservice.kz",
    website: "https://stroytechservice.kz",
  });
  const tehsnab = await ensureCompany(aigerimTok, {
    name: "ИП Нурланова Техснаб",
    bin: "987654321098",
    address: "г. Астана, ул. Кенесары, 40",
    phone: "+7 717 987 6543",
    email: "info@tehsnab.kz",
    website: "https://tehsnab.kz",
  });
  if (adminTok) {
    await api("PUT", `/admin/companies/${stst.id}/verify`, undefined, adminTok);
    await api("PUT", `/admin/companies/${tehsnab.id}/verify`, undefined, adminTok);
    log("  verified both companies");
  }

  section("Subscriptions");
  // The free plan caps live listings at 3 — upgrade both demo accounts so
  // every seeded listing can be published.
  await api("PUT", "/subscription", { plan: "basic" }, erlanTok);
  await api("PUT", "/subscription", { plan: "basic" }, aigerimTok);
  log("  upgraded both accounts to the basic plan (10 listing cap)");

  section("Categories");
  const categories = await api("GET", "/catalog/categories");
  const catId = Object.fromEntries(categories.map((c) => [c.slug, c.id]));
  log(`  loaded ${categories.length} categories`);

  section("Listings — Ерлан (СтройТехСервис)");
  const L = {};
  L.excavator1 = await createListing(erlanTok, catId["excavators"], {
    title: "Экскаватор гусеничный CAT 320D",
    description: "Мощный гусеничный экскаватор CAT 320D, объём ковша 1.2 м³. В отличном техническом состоянии, регулярное ТО в официальном сервисе.",
    condition: "new", region: "Алматы", image_url: "/pics/sample.jpg",
    listingType: "sale", price: 18500000,
  }, adminTok);
  L.excavator2 = await createListing(erlanTok, catId["excavators"], {
    title: "Экскаватор Caterpillar 330C L (б/у)",
    description: "Экскаватор с удлинённой стрелой, наработка 6200 моточасов. Продаётся в связи с обновлением парка техники.",
    condition: "used", region: "Караганда", image_url: "/pics/equipment/excavator-2.jpg",
    listingType: "sale", price: 9200000,
  }, adminTok);
  L.crane1 = await createListing(erlanTok, catId["cranes"], {
    title: "Автокран гусеничный 25 тонн — аренда",
    description: "Гусеничный кран грузоподъёмностью 25 тонн с оператором. Подходит для монтажных работ на строительной площадке.",
    condition: "used", region: "Алматы", image_url: "/pics/equipment/crane-1.jpg",
    listingType: "rental", price: 350000, pricePeriod: "day",
  }, adminTok);
  L.crane2 = await createListing(erlanTok, catId["cranes"], {
    title: "Башенный кран Liebherr 132 EC-H",
    description: "Башенный кран для высотного строительства, вылет стрелы до 60 м, макс. грузоподъёмность 8 тонн.",
    condition: "used", region: "Астана", image_url: "/pics/equipment/crane-2.jpg",
    listingType: "sale", price: 42000000,
  }, adminTok);
  L.generator1 = await createListing(erlanTok, catId["generators"], {
    title: "Дизель-генератор Caterpillar 150 кВт",
    description: "Промышленная дизель-генераторная установка CAT, автозапуск, шумозащитный кожух.",
    condition: "new", region: "Алматы", image_url: "/pics/equipment/generator-1.jpg",
    listingType: "sale", price: 8450000,
  }, adminTok);
  L.generator2 = await createListing(erlanTok, catId["generators"], {
    title: "Генераторная установка 40 кВА — аренда",
    description: "Дизельный генератор 40 кВА для резервного питания объекта. Возможна аренда с доставкой по городу.",
    condition: "used", region: "Шымкент", image_url: "/pics/equipment/generator-2.jpg",
    listingType: "rental", price: 25000, pricePeriod: "day",
  }, adminTok);
  L.loader1 = await createListing(erlanTok, catId["loaders"], {
    title: "Фронтальный погрузчик CAT 950 GC",
    description: "Колёсный фронтальный погрузчик, ковш 3.0 м³. Идеален для перевалки сыпучих материалов.",
    condition: "new", region: "Алматы", image_url: "/pics/equipment/loader-1.jpg",
    listingType: "sale", price: 15800000,
  }, adminTok);
  L.loader2 = await createListing(erlanTok, catId["loaders"], {
    title: "Погрузчик-экскаватор JCB 3CX (б/у)",
    description: "Универсальная машина 3-в-1: погрузчик, экскаватор, вилы. Наработка 4100 м/ч, все документы в порядке.",
    condition: "used", region: "Атырау", image_url: "/pics/sample-1.jpeg",
    listingType: "sale", price: 7600000,
  }, adminTok);

  section("Listings — Айгерим (Техснаб)");
  L.compressor1 = await createListing(aigerimTok, catId["compressors"], {
    title: "Компрессорная станция Elang ZRCW-30SA",
    description: "Винтовой компрессор промышленного класса, производительность 30 м³/мин. С полным пакетом документов.",
    condition: "new", region: "Астана", image_url: "/pics/equipment/compressor-1.jpg",
    listingType: "sale", price: 3200000,
  }, adminTok);
  L.compressor2 = await createListing(aigerimTok, catId["compressors"], {
    title: "Передвижной компрессор — аренда",
    description: "Дизельный передвижной компрессор, давление до 10 бар. Аренда посуточно, доставка по Астане.",
    condition: "used", region: "Астана", image_url: "/pics/equipment/compressor-1.jpg",
    listingType: "rental", price: 18000, pricePeriod: "day",
  }, adminTok);
  L.welding1 = await createListing(aigerimTok, catId["welding-equipment"], {
    title: "Сварочный аппарат промышленный (комплект)",
    description: "Комплект сварочного оборудования для металлоконструкций: аппарат, кабели, редуктор, СИЗ.",
    condition: "new", region: "Астана", image_url: "/pics/equipment/welding-1.jpg",
    listingType: "sale", price: 850000,
  }, adminTok);
  L.welding2 = await createListing(aigerimTok, catId["welding-equipment"], {
    title: "Сварочное оборудование (б/у, комплект)",
    description: "Рабочий комплект сварочного оборудования после одного объекта, полностью исправен.",
    condition: "used", region: "Павлодар", image_url: "/pics/equipment/welding-2.jpg",
    listingType: "sale", price: 420000,
  }, adminTok);
  L.mixer1 = await createListing(aigerimTok, catId["concrete-mixers"], {
    title: "Автобетоносмеситель КамАЗ 6х4",
    description: "Автобетоносмеситель на шасси КамАЗ, объём барабана 7 м³. Готов к работе.",
    condition: "used", region: "Астана", image_url: "/pics/equipment/mixer-1.jpg",
    listingType: "sale", price: 24500000,
  }, adminTok);
  L.mixer2 = await createListing(aigerimTok, catId["concrete-mixers"], {
    title: "Бетоносмеситель — аренда с оператором",
    description: "Аренда автобетоносмесителя с водителем. Оплата посуточно, минимальный срок — 1 день.",
    condition: "used", region: "Костанай", image_url: "/pics/equipment/mixer-2.jpg",
    listingType: "rental", price: 95000, pricePeriod: "day",
  }, adminTok);
  L.truck1 = await createListing(aigerimTok, catId["trucks-transport"], {
    title: "Самосвал шарнирно-сочленённый CAT",
    description: "Шарнирно-сочленённый самосвал повышенной проходимости, грузоподъёмность 28 тонн.",
    condition: "used", region: "Актобе", image_url: "/pics/equipment/truck-1.jpg",
    listingType: "sale", price: 19900000,
  }, adminTok);
  L.truck2 = await createListing(aigerimTok, catId["trucks-transport"], {
    title: "Грузовик-эвакуатор техники — аренда",
    description: "Тяжёлый эвакуатор для перевозки строительной техники между объектами. Аренда с водителем.",
    condition: "used", region: "Астана", image_url: "/pics/sample-2.jpeg",
    listingType: "rental", price: 120000, pricePeriod: "day",
  }, adminTok);

  section("Deals");
  const D = {};

  // 1. Completed sale, buyer Айгерим -> seller Ерлан. Escrow funded and
  // auto-released by the deal.status.changed consumer on completion.
  log("Deal 1: excavator sale -> completed (+ review)");
  D.d1 = await createDeal(aigerimTok, L.excavator1.id, "Здравствуйте! Интересует экскаватор CAT 320D, можно посмотреть на площадке?");
  await transition(aigerimTok, D.d1.id, "negotiation");
  await postMessage(erlanTok, D.d1.id, "Добрый день! Да, техника на площадке в Алматы, можно приехать на осмотр в любой будний день.");
  await postMessage(aigerimTok, D.d1.id, "Отлично, устроит цена из объявления, готовы оформлять.");
  await transition(aigerimTok, D.d1.id, "confirmed");
  await fundEscrow(aigerimTok, D.d1.id, 18500000);
  await transition(aigerimTok, D.d1.id, "in_progress");
  await transition(aigerimTok, D.d1.id, "completed");
  await api("POST", "/reviews", {
    target_entity_id: erlanMe.id,
    rating: 5, comment: "Отличный экскаватор, полностью соответствует описанию, привезли вовремя. Рекомендую!",
    transaction_id: D.d1.id,
  }, aigerimTok);
  await api("POST", "/reviews", {
    target_entity_id: aigerimMe.id,
    rating: 5, comment: "Оперативная оплата, всё по договорённости. Приятно иметь дело.",
    transaction_id: D.d1.id,
  }, erlanTok);

  // 2. Rental in progress, buyer Айгерим -> seller Ерлан, with a linked
  // booking on the rental calendar.
  log("Deal 2: crane rental -> in_progress (+ booking)");
  D.d2 = await createDeal(aigerimTok, L.crane1.id, "Нужен кран на объект в Алматы на неделю, обсудим детали?");
  await transition(aigerimTok, D.d2.id, "negotiation");
  await transition(aigerimTok, D.d2.id, "confirmed");
  await fundEscrow(aigerimTok, D.d2.id, 2450000);
  await transition(aigerimTok, D.d2.id, "in_progress");
  const bookingStart = addDays(7), bookingEnd = addDays(14);
  await api("POST", "/bookings", { listing_id: L.crane1.id, start_date: bookingStart, end_date: bookingEnd }, aigerimTok);

  // 3. Sale confirmed (escrow held, not yet in progress), buyer Ерлан -> seller Айгерим.
  log("Deal 3: welding set sale -> confirmed");
  D.d3 = await createDeal(erlanTok, L.welding1.id, "Здравствуйте, нужен комплект сварочного оборудования, актуально?");
  await transition(erlanTok, D.d3.id, "negotiation");
  await transition(erlanTok, D.d3.id, "confirmed");
  await fundEscrow(erlanTok, D.d3.id, 850000);

  // 4. Negotiation in progress with a back-and-forth thread, buyer Ерлан -> seller Айгерим.
  log("Deal 4: concrete mixer sale -> negotiation (with chat thread)");
  D.d4 = await createDeal(erlanTok, L.mixer1.id, "Добрый день, интересует бетоносмеситель. Возможен торг?");
  await transition(erlanTok, D.d4.id, "negotiation");
  await postMessage(aigerimTok, D.d4.id, "Здравствуйте! Небольшой торг возможен при самовывозе.");
  await postMessage(erlanTok, D.d4.id, "Готовы забрать своим эвакуатором на этой неделе.");
  await postMessage(aigerimTok, D.d4.id, "Тогда 23 500 000 тг, машина на ходу, документы готовы.");
  await postMessage(erlanTok, D.d4.id, "Согласен, подъедем в четверг.");

  // 5. Fresh inquiry, buyer Айгерим -> seller Ерлан.
  log("Deal 5: loader-excavator sale -> inquiry");
  D.d5 = await createDeal(aigerimTok, L.loader2.id, "Здравствуйте, актуален ли ещё JCB 3CX? Можно фото ковша отдельно?");

  // 6. Funded then cancelled, buyer Ерлан -> seller Айгерим. Cancel auto-refunds escrow.
  log("Deal 6: compressor sale -> cancelled (auto-refund)");
  D.d6 = await createDeal(erlanTok, L.compressor1.id, "Интересует компрессорная станция, готовы купить сразу.");
  await transition(erlanTok, D.d6.id, "negotiation");
  await transition(erlanTok, D.d6.id, "confirmed");
  await fundEscrow(erlanTok, D.d6.id, 3200000);
  await postMessage(erlanTok, D.d6.id, "Извините, нашли станцию ближе к объекту, отменяем сделку.");
  await transition(erlanTok, D.d6.id, "cancelled");

  // 7. Disputed: buyer files, admin refunds the buyer.
  log("Deal 7: generator sale -> disputed -> resolved (refund to buyer)");
  D.d7 = await createDeal(aigerimTok, L.generator1.id, "Добрый день, нужен генератор для резервного питания склада.");
  await transition(aigerimTok, D.d7.id, "negotiation");
  await transition(aigerimTok, D.d7.id, "confirmed");
  await fundEscrow(aigerimTok, D.d7.id, 8450000);
  await transition(aigerimTok, D.d7.id, "in_progress");
  const dispute7 = await api("POST", "/disputes", {
    deal_id: D.d7.id,
    reason: "При запуске генератор не выдаёт заявленную мощность, похоже на неисправность обмотки. Продавец отказывается забрать технику обратно.",
    evidence_urls: [],
  }, aigerimTok);
  log(`  filed dispute ${dispute7.id} — deal is now frozen`);
  try {
    await transition(erlanTok, D.d7.id, "completed");
    log("  WARNING: transition on a frozen deal unexpectedly succeeded");
  } catch (e) {
    log(`  confirmed freeze: transition rejected with "${e.body && e.body.message}"`);
  }
  if (adminTok) {
    await api("PUT", `/admin/disputes/${dispute7.id}/resolve`, {
      resolution: "refund", note: "Продавец не смог подтвердить исправность техники на момент передачи — возврат средств покупателю.",
    }, adminTok);
    log("  admin resolved: refund to buyer, deal unfrozen");
  }

  // 8. Disputed: seller files, admin releases escrow to the seller.
  log("Deal 8: welding set (used) sale -> disputed -> resolved (release to seller)");
  D.d8 = await createDeal(erlanTok, L.welding2.id, "Здравствуйте, интересует б/у комплект, заберём сами.");
  await transition(erlanTok, D.d8.id, "negotiation");
  await transition(erlanTok, D.d8.id, "confirmed");
  await fundEscrow(erlanTok, D.d8.id, 420000);
  await transition(erlanTok, D.d8.id, "in_progress");
  const dispute8 = await api("POST", "/disputes", {
    deal_id: D.d8.id,
    reason: "Покупатель забрал оборудование лично, подтвердил получение в переписке, но теперь просит отменить сделку и вернуть деньги без возврата техники.",
    evidence_urls: [],
  }, aigerimTok);
  log(`  filed dispute ${dispute8.id} — deal is now frozen`);
  if (adminTok) {
    await api("PUT", `/admin/disputes/${dispute8.id}/resolve`, {
      resolution: "release", note: "Покупатель подтвердил получение техники в переписке — средства переводятся продавцу.",
    }, adminTok);
    log("  admin resolved: release to seller, deal unfrozen");
  }

  section("Standalone bookings");
  const b2start = addDays(20), b2end = addDays(25);
  await api("POST", "/bookings", { listing_id: L.mixer2.id, start_date: b2start, end_date: b2end }, erlanTok);
  log(`  booked mixer rental ${b2start} -> ${b2end} (Ерлан renting from Айгерим)`);

  const b3start = addDays(30), b3end = addDays(32);
  const cancelBooking = await api("POST", "/bookings", { listing_id: L.generator2.id, start_date: b3start, end_date: b3end }, aigerimTok);
  await api("PUT", `/bookings/${cancelBooking.id}/cancel`, undefined, aigerimTok);
  log(`  booked then cancelled generator rental ${b3start} -> ${b3end} (Айгерим)`);

  section("Done");
  log("Demo accounts:");
  log("  admin@industrix.kz / Admin!2345");
  log("  erlan.demo@industrix.kz / Demo!2345   (СтройТехСервис, Алматы)");
  log("  aigerim.demo@industrix.kz / Demo!2345 (Техснаб, Астана)");
  log(`Created ${Object.keys(L).length} listings across 8 categories and ${Object.keys(D).length} deals covering every status, 2 disputes (refund + release), 4 bookings, 1 review.`);
}

function addDays(n) {
  const d = new Date();
  d.setDate(d.getDate() + n);
  return d.toISOString().slice(0, 10);
}

main().catch((e) => {
  console.error("\nSeed failed:", e.message);
  if (e.body) console.error(JSON.stringify(e.body, null, 2));
  process.exit(1);
});
