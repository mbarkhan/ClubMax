import type { ReactNode } from "react";
import Link from "next/link";
import SiteLogo from "./SiteLogo";
import "../terms.css";

const legalLinks = [
  ["about", "درباره ما"],
  ["contact", "تماس با ما"],
  ["terms", "قوانین و مقررات"],
  ["privacy", "سیاست حریم خصوصی"],
  ["refunds", "لغو و استرداد وجه"],
  ["pricing", "خدمات و قیمت‌ها"],
] as const;

export default function LegalPage({
  locale,
  title,
  eyebrow,
  intro,
  children,
}: {
  locale: string;
  title: string;
  eyebrow: string;
  intro: string;
  children: ReactNode;
}) {
  return (
    <main className="terms-page">
      <header className="terms-header">
        <SiteLogo href={`/${locale}`} priority />
        <nav aria-label="مسیرهای اصلی">
          <Link href={`/${locale}`}>صفحه اصلی</Link>
          <Link className="terms-login" href={`/${locale}/login`}>ورود</Link>
          <Link className="terms-join" href={`/${locale}/signup`}>ثبت‌نام رایگان</Link>
        </nav>
      </header>
      <section className="terms-hero">
        <div>
          <span>{eyebrow}</span>
          <h1>{title}</h1>
          <p>{intro}</p>
          <small>آخرین به‌روزرسانی: ۱۳ شهریور ۱۴۰۵</small>
        </div>
      </section>
      <div className="terms-layout">
        <aside>
          <strong>اطلاعات و قوانین</strong>
          <nav>
            {legalLinks.map(([path, label]) => <Link key={path} href={`/${locale}/${path}`}>{label}</Link>)}
          </nav>
        </aside>
        <article className="terms-content">{children}</article>
      </div>
      <footer className="terms-footer">
        <SiteLogo href={`/${locale}`} />
        <p>سامانه مدیریت باشگاه، مربی و ورزشکار</p>
        <Link href={`/${locale}`}>بازگشت به صفحه اصلی</Link>
      </footer>
    </main>
  );
}
