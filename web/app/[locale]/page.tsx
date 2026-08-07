import Link from "next/link";

const copy = {
  fa: {eyebrow: "مدیریت هوشمند ورزشگاه", title: "همه‌چیز برای اداره باشگاه، یک‌جا.", body: "GyMax ارتباط مربی، ورزشکار و باشگاه را ساده می‌کند.", action: "ورود به سامانه", language: "English"},
  en: {eyebrow: "Smart gym management", title: "Everything your gym needs, in one place.", body: "GyMax simplifies the connection between coaches, athletes and gyms.", action: "Sign in", language: "فارسی"},
};

export default async function Home({params}: {params: Promise<{locale: "fa" | "en"}>}) {
  const {locale} = await params;
  const t = copy[locale] ?? copy.fa;
  return <main><nav><strong>GyMax</strong><Link href={locale === "fa" ? "/en" : "/fa"}>{t.language}</Link></nav><section><p className="eyebrow">{t.eyebrow}</p><h1>{t.title}</h1><p className="lead">{t.body}</p><button type="button">{t.action}</button></section></main>;
}
