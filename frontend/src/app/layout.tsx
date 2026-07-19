import type { Metadata } from "next";
import "./globals.css";
import { Geist } from "next/font/google";
import { cn } from "@/lib/utils";
import { Navbar } from "@/components/layout/Navbar"
import {Footer} from "@/components/layout/Footer"
import { I18nProvider } from "@/lib/i18n"

const geist = Geist({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: "Industrix",
  description: "ESG Industrix platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // lang starts as the default locale; I18nProvider updates it client-side
  // once a stored preference is applied.
  return (
    <html lang="ru" className={cn("font-sans", geist.variable)}>
      <body className="antialiased">
        <I18nProvider>
          <Navbar />
          <main>{children}</main>
          <Footer/>
        </I18nProvider>
      </body>
    </html>
  );
}