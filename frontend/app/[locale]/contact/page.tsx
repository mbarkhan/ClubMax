import type { Metadata } from "next";
import LegalPage from "../../components/LegalPage";

export const metadata: Metadata = {
  title: "تماس با ما | ClubMax",
  description: "اطلاعات تماس و پشتیبانی ClubMax",
};

export default async function ContactPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return <LegalPage locale={locale} eyebrow="پاسخ‌گویی و پشتیبانی" title="تماس با کلاب ماکس" intro="برای پرسش‌ها، پیگیری خطای پرداخت و مسائل مربوط به حساب کاربری در کنار شما هستیم.">
    <section className="terms-section"><h2>راه‌های تماس</h2><ol><li>شماره تماس پاسخ‌گو: <a dir="ltr" href="tel:09357230033">09357230033</a></li><li>ایمیل: <a dir="ltr" href="mailto:clubmax@gmail.com">clubmax@gmail.com</a></li><li>ساعت پاسخ‌گویی تلفنی: شنبه تا چهارشنبه، ساعت ۹ تا ۱۶.</li><li>کاربران عضو می‌توانند درخواست‌های خود را از طریق سامانه تیکت پشتیبانی نیز ثبت کنند.</li></ol></section>
    <section className="terms-section"><h2>نشانی پستی</h2><ol><li>آذربایجان غربی، مهاباد، کوی زیبا، گلستان ۱، پلاک ۱۶</li><li>کدپستی: <span dir="ltr">5914815144</span></li></ol></section>
    <section className="terms-section"><h2>پیگیری پرداخت</h2><ol><li>اگر مبلغی از حساب کسر شد اما پرداخت در سامانه ثبت نشد، کد پیگیری تراکنش را از طریق تیکت پشتیبانی ارسال کنید.</li><li>درخواست حداکثر ظرف ۴۸ ساعت بررسی و نتیجه از طریق سامانه اعلام می‌شود.</li></ol></section>
  </LegalPage>;
}
