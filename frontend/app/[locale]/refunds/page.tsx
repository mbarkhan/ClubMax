import type { Metadata } from "next";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "لغو و استرداد وجه | ClubMax",
  description: "سیاست لغو، انصراف و رسیدگی به خطای پرداخت ClubMax",
};

export default async function RefundsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return <LegalPage locale={locale} eyebrow="شفافیت پرداخت" title="لغو، انصراف و استرداد وجه" intro="این سیاست فقط مربوط به پرداخت شهریه‌ای است که از طریق سامانه برای باشگاه یا کلاس انجام می‌شود.">
    <section className="terms-section"><h2>ماهیت پرداخت</h2><ol><li>ClubMax در حال حاضر اشتراک یا پلن فروشی ندارد.</li><li>مبلغ پرداختی، شهریه باشگاه، کلاس یا عضویت است که توسط مربی یا باشگاه تعریف شده و مبلغ آن پیش از پرداخت به ورزشکار نمایش داده می‌شود.</li><li>پس از پرداخت موفق، شهریه به حساب مربی یا باشگاه مربوط تخصیص می‌یابد.</li></ol></section>
    <section className="terms-section"><h2>لغو کلاس یا عضویت</h2><ol><li>لغو کلاس، عضویت یا انصراف ورزشکار پس از ثبت پرداخت موفق، شامل استرداد شهریه نیست؛ زیرا مبلغ بابت شهریه دریافت و برای مربی یا باشگاه تخصیص داده شده است.</li><li>این سیاست مانع اجرای هرگونه الزام قانونی یا دستور مرجع صالح نخواهد بود.</li></ol></section>
    <section className="terms-section"><h2>خطای پرداخت</h2><ol><li>اگر وجه از حساب ورزشکار کسر شد اما پرداخت در سامانه موفق ثبت نشد، کاربر باید کد پیگیری تراکنش را از طریق تیکت پشتیبانی ارسال کند.</li><li>درخواست حداکثر ظرف ۴۸ ساعت بررسی و نتیجه به کاربر اعلام می‌شود.</li><li>در صورت احراز خطای پرداخت، اصلاح ثبت تراکنش یا پیگیری بازگشت وجه مطابق نتیجه بررسی و ضوابط درگاه پرداخت انجام می‌شود.</li></ol></section>
  </LegalPage>;
}
