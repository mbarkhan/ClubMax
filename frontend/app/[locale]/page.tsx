import Link from "next/link";
import SiteLogo from "../components/SiteLogo";
import "../home.css";
export default async function Home({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return (
    <main className="home">
      <header>
        <SiteLogo href={`/${locale}`} priority />
        <nav>
          <a href="#features">امکانات</a>
          <a href="#gyms">باشگاه‌ها</a>
          <a href="#news">اخبار</a>
        </nav>
        <div>
          <Link className="login-link" href={`/${locale}/login`}>
            ورود
          </Link>
          <Link className="join-link" href={`/${locale}/signup`}>
            ثبت‌نام رایگان
          </Link>
        </div>
      </header>
      <section className="home-hero">
        <div>
          <span>مدیریت حرفه‌ای باشگاه ورزشی</span>
          <h1>
            باشگاهت را
            <br />
            <em>هوشمندتر</em> مدیریت کن.
          </h1>
          <p>ClubMax فضای یکپارچه مربی، ورزشکاره؛ از ثبت اعضا تا برنامه‌ها و پرداخت‌ها.</p>
          <div>
            <Link href={`/${locale}/signup`}>شروع ثبت‌نام ←</Link>
            <Link href={`/${locale}/login`}>حساب کاربری دارم</Link>
          </div>
          <small>● نسخه توسعه در حال تکمیل</small>
        </div>
        <aside>
          <div className="hero-card">
            <span>نمای کلی امروز</span>
            <strong>همه چیز یک‌جا</strong>
            {[
              ["۳۶", "باشگاه فعال", "مدیریت متمرکز"],
              ["✓", "اطلاعات امن", "دسترسی نقش‌محور"],
              ["↑", "رشد مستمر", "گزارش‌های کاربردی"],
            ].map((x) => (
              <div key={x[1]}>
                <i>{x[0]}</i>
                <p>
                  {x[1]}
                  <small>{x[2]}</small>
                </p>
              </div>
            ))}
          </div>
        </aside>
      </section>
      <section id="features" className="home-section">
        <span>برای هر نقش، یک تجربه اختصاصی</span>
        <h2>ساخته‌شده برای دنیای واقعی باشگاه</h2>
        <div className="feature-grid">
          {[
            ["م", "برای مربیان", "ثبت باشگاه، مدیریت ورزشکاران و پیگیری امور روزمره."],
            ["و", "برای ورزشکاران", "عضویت در باشگاه‌ها و دسترسی آسان به اطلاعات شخصی."],
            ["ا", "امن و یکپارچه", "اطلاعات متمرکز، مجوزهای دقیق و تجربه‌ای مناسب موبایل."],
          ].map((x) => (
            <article key={x[1]}>
              <i>{x[0]}</i>
              <h3>{x[1]}</h3>
              <p>{x[2]}</p>
            </article>
          ))}
        </div>
      </section>
      <section id="gyms" className="placeholder-section">
        <div>
          <span>باشگاه‌های منتخب</span>
          <h2>باشگاه مناسب خودت را پیدا کن</h2>
          <p>فهرست و جست‌وجوی باشگاه‌ها در مرحله بعد اینجا قرار می‌گیرد.</p>
        </div>
        <b>به‌زودی</b>
      </section>
      <section id="news" className="home-section news">
        <span>اخبار و تازه‌ها</span>
        <h2>از ClubMax چه خبر؟</h2>
        <div className="news-grid">
          {["راهنمای شروع برای مربیان", "امنیت اطلاعات ورزشکاران", "نسخه بعدی ClubMax"].map(
            (x, i) => (
              <article key={x}>
                <small>خبر {i + 1}</small>
                <h3>{x}</h3>
                <p>این بخش در ادامه پروژه با محتوای واقعی تکمیل خواهد شد.</p>
              </article>
            )
          )}
        </div>
      </section>
      <footer>
        <SiteLogo href={`/${locale}`} priority />
        <p>سامانه مدیریت باشگاه، مربی و ورزشکار</p>
      </footer>
    </main>
  );
}
