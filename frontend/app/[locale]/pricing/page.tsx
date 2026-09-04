import type { Metadata } from "next";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "خدمات و قیمت‌ها | ClubMax",
  description: "توضیح خدمات و نحوه تعیین شهریه در ClubMax",
};

export default async function PricingPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return <LegalPage locale={locale} eyebrow="خدمات و مبالغ پرداختی" title="خدمات، شهریه و قیمت‌ها" intro="شفافیت مبلغ و دلیل پرداخت، پیش از ثبت هر تراکنش.">
    <section className="terms-section"><h2>خدمات کلاب ماکس</h2><ol><li>کلاب ماکس سامانه مدیریت باشگاه، مربی و ورزشکار است و امکاناتی مانند مدیریت عضویت، اطلاعات باشگاه، ورزشکاران، شهریه و پشتیبانی را فراهم می‌کند.</li><li>استفاده از سامانه در حال حاضر شامل فروش پلن یا اشتراک مستقل از سوی ClubMax نیست.</li></ol></section>
    <section className="terms-section"><h2>مبلغی که کاربر پرداخت می‌کند</h2><ol><li>پرداخت‌های انجام‌شده از طریق سامانه، شهریه باشگاه، کلاس یا عضویت ورزشی است.</li><li>مبلغ شهریه، نحوه محاسبه آن و مشخصات خدمت توسط مربی یا باشگاه تعریف می‌شود.</li><li>کاربر پیش از تأیید پرداخت، مبلغ و اطلاعات مرتبط با باشگاه یا خدمت را مشاهده می‌کند و باید آن‌ها را بررسی کند.</li></ol></section>
    <section className="terms-section"><h2>تغییر مبلغ یا شرایط</h2><ol><li>هرگونه تغییر در شهریه یا شرایط کلاس از سوی مربی یا باشگاه مربوط اعلام و برای پرداخت‌های بعدی اعمال می‌شود.</li><li>برای اطلاع از سیاست لغو و رسیدگی به خطای پرداخت، صفحه <a href={`/${locale}/refunds`}>لغو و استرداد وجه</a> را مطالعه کنید.</li></ol></section>
  </LegalPage>;
}
