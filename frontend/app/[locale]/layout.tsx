import type {Metadata} from "next";
import {notFound} from "next/navigation";
import "@fontsource/vazirmatn/400.css";
import "@fontsource/vazirmatn/500.css";
import "@fontsource/vazirmatn/600.css";
import "@fontsource/vazirmatn/700.css";
import "../globals.css";
import "../member.css";

const locales = ["fa", "en"] as const;

export const metadata: Metadata = {
  title: "GyMax",
  description: "سامانه مدیریت باشگاه ورزشی",
};

export function generateStaticParams() { return locales.map((locale) => ({locale})); }

export default async function LocaleLayout({children, params}: Readonly<{children: React.ReactNode; params: Promise<{locale: string}>}>) {
  const {locale} = await params;
  if (!locales.includes(locale as (typeof locales)[number])) notFound();
  const rtl = locale === "fa";
  return <html lang={locale} dir={rtl ? "rtl" : "ltr"}><body>{children}</body></html>;
}
