// Translation dictionaries. Russian is the primary language and doubles as the
// key set: every other locale must provide the same keys, which TypeScript
// enforces via the Dictionary type below.
//
// Kazakh is the second locale. English is intentionally not offered — the
// marketplace is KZ-facing.

export const LOCALES = ["ru", "kk"] as const;
export type Locale = (typeof LOCALES)[number];

export const LOCALE_NAMES: Record<Locale, string> = {
  ru: "Рус",
  kk: "Қаз",
};

export const DEFAULT_LOCALE: Locale = "ru";

const ru = {
  // Navigation / shell
  "nav.catalog": "Каталог",
  "nav.sell": "Продать",
  "nav.deals": "Сделки",
  "nav.bookings": "Бронирования",
  "nav.favorites": "Избранное",
  "nav.notifications": "Уведомления",
  "nav.about": "О нас",
  "nav.login": "Войти",
  "nav.register": "Регистрация",
  "nav.signOut": "Выйти",
  "nav.account": "Профиль",
  "nav.admin": "Админ-панель",

  // Catalog — search
  "catalog.searchPlaceholder": "Поиск техники: экскаватор, кран, бетономешалка…",
  "catalog.search": "Найти",
  "catalog.clearSearch": "Очистить",
  "catalog.resultsFor": "Результаты по запросу",
  "catalog.found": "Найдено",
  "catalog.nothingFound": "Ничего не найдено",
  "catalog.nothingFoundHint": "Попробуйте изменить запрос или сбросить фильтры",
  "catalog.loading": "Загрузка…",
  "catalog.searchUnavailable": "Поиск временно недоступен",

  // Catalog — filters
  "filters.title": "Фильтры",
  "filters.reset": "Сбросить",
  "filters.category": "Категория",
  "filters.region": "Регион",
  "filters.priceRange": "Цена",
  "filters.min": "От",
  "filters.max": "До",
  "filters.condition": "Состояние",
  "filters.saleOrRent": "Продажа или аренда",
  "filters.apply": "Применить",
  "filters.clearAll": "Сбросить всё",

  // Values
  "condition.new": "Новое",
  "condition.used": "Б/у",
  "listingType.sale": "Продажа",
  "listingType.rental": "Аренда",

  // Sorting / view
  "sort.newest": "Сначала новые",
  "sort.priceAsc": "Сначала дешёвые",
  "sort.priceDesc": "Сначала дорогие",
  "sort.relevance": "По релевантности",
  "view.grid": "Плитка",
  "view.list": "Список",

  // Pagination
  "pagination.prev": "Назад",
  "pagination.next": "Вперёд",
  "pagination.showing": "Показано",
  "pagination.of": "из",
} as const;

/** Every locale must cover exactly the Russian key set. */
export type Dictionary = Record<keyof typeof ru, string>;
export type TranslationKey = keyof typeof ru;

const kk: Dictionary = {
  "nav.catalog": "Каталог",
  "nav.sell": "Сату",
  "nav.deals": "Мәмілелер",
  "nav.bookings": "Брондау",
  "nav.favorites": "Таңдаулылар",
  "nav.notifications": "Хабарламалар",
  "nav.about": "Біз туралы",
  "nav.login": "Кіру",
  "nav.register": "Тіркелу",
  "nav.signOut": "Шығу",
  "nav.account": "Профиль",
  "nav.admin": "Әкімші панелі",

  "catalog.searchPlaceholder": "Техника іздеу: экскаватор, кран, бетон араластырғыш…",
  "catalog.search": "Іздеу",
  "catalog.clearSearch": "Тазалау",
  "catalog.resultsFor": "Сұраныс бойынша нәтижелер",
  "catalog.found": "Табылды",
  "catalog.nothingFound": "Ештеңе табылмады",
  "catalog.nothingFoundHint": "Сұранысты өзгертіп немесе сүзгілерді тазалап көріңіз",
  "catalog.loading": "Жүктелуде…",
  "catalog.searchUnavailable": "Іздеу уақытша қолжетімсіз",

  "filters.title": "Сүзгілер",
  "filters.reset": "Тазалау",
  "filters.category": "Санат",
  "filters.region": "Аймақ",
  "filters.priceRange": "Бағасы",
  "filters.min": "Бастап",
  "filters.max": "Дейін",
  "filters.condition": "Күйі",
  "filters.saleOrRent": "Сату немесе жалға",
  "filters.apply": "Қолдану",
  "filters.clearAll": "Барлығын тазалау",

  "condition.new": "Жаңа",
  "condition.used": "Қолданылған",
  "listingType.sale": "Сату",
  "listingType.rental": "Жалға",

  "sort.newest": "Алдымен жаңалары",
  "sort.priceAsc": "Алдымен арзаны",
  "sort.priceDesc": "Алдымен қымбаты",
  "sort.relevance": "Өзектілігі бойынша",
  "view.grid": "Тор",
  "view.list": "Тізім",

  "pagination.prev": "Артқа",
  "pagination.next": "Алға",
  "pagination.showing": "Көрсетілді",
  "pagination.of": "ішінен",
};

export const dictionaries: Record<Locale, Dictionary> = { ru, kk };
