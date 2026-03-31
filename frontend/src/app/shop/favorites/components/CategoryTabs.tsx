"use client";

interface CategoryTabsProps {
  tabs:     string[];
  active:   string;
  total:    number;
  onSelect: (tab: string) => void;
}

export function CategoryTabs({ tabs, active, total, onSelect }: CategoryTabsProps) {
  return (
    <div className="flex items-center gap-1 bg-gray-100 p-1 rounded-lg">
      {tabs.map((tab) => {
        const isActive = tab === active;
        const label    = tab === "All" ? `All Items(${total})` : tab;
        return (
          <button
            key={tab}
            onClick={() => onSelect(tab)}
            className={`
              px-4 py-1.5 rounded-md text-[13px] font-semibold transition-colors border-none cursor-pointer
              ${isActive
                ? "bg-blue-600 text-white shadow-sm"
                : "bg-transparent text-gray-600 hover:text-gray-900"}
            `}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}
