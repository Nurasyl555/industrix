import Link from "next/link";
import { ChevronRight } from "lucide-react";

export function Breadcrumb() {
  return (
    <nav className="flex items-center gap-1.5 text-[13px]">
      <Link href="/" className="text-gray-500 hover:text-gray-900 no-underline transition-colors">
        Home
      </Link>
      <ChevronRight size={13} className="text-gray-400" />
      <span className="font-semibold text-gray-900">Favorites</span>
    </nav>
  );
}
