import type { Metadata } from "next";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "سیاست حریم خصوصی | ClubMax",
  description: "سیاست حریم خصوصی سامانه ClubMax",
};

export default async function PrivacyPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return <LegalPage locale={locale} eyebrow="امنیت و محرمانگی" title="سیاست حریم خصوصی" intro="این صفحه توضیح می‌دهد کلاب ماکس چه اطلاعاتی را برای ارائه خدمات پردازش می‌کند و چگونه از آن محافظت می‌شود.">
    <section className="terms-section"><h2>اطلاعاتی که پردازش می‌کنیم</h2><ol><li>اطلاعات حساب، مشخصات هویتی و اطلاعات تماس که کاربر هنگام ثبت‌نام یا تکمیل پروفایل وارد می‌کند.</li><li>اطلاعات باشگاه، عضویت، شهریه و تراکنش‌های لازم برای ارائه خدمت و پشتیبانی.</li><li>مدارک و اطلاعاتی که کاربر با اختیار خود برای استفاده از امکانات سامانه بارگذاری می‌کند.</li></ol></section>
    <section className="terms-section"><h2>هدف استفاده از اطلاعات</h2><ol><li>ایجاد و مدیریت حساب کاربری، ارائه خدمات باشگاه و ثبت پرداخت‌ها.</li><li>پاسخ‌گویی به تیکت‌ها، جلوگیری از تقلب و حفظ امنیت سامانه.</li><li>رعایت الزامات قانونی و رسیدگی به درخواست‌های مراجع صالح، در مواردی که قانون ایجاب کند.</li></ol></section>
    <section className="terms-section"><h2>دسترسی و حفاظت</h2><ol><li>اطلاعات فقط در حد نیاز برای ارائه خدمت در اختیار کاربر، باشگاه یا مربی دارای دسترسی مجاز قرار می‌گیرد.</li><li>کلاب ماکس اطلاعات کاربران را نمی‌فروشد یا اجاره نمی‌دهد.</li><li>کاربران مسئول حفظ رمز عبور خود هستند و باید هرگونه دسترسی مشکوک را از طریق پشتیبانی اعلام کنند.</li></ol></section>
  </LegalPage>;
}
