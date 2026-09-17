import Link from "next/link";
import SiteLogo from "../components/SiteLogo";
import GymShowcase from "../components/GymShowcase";
import HomeAccountActions from "../components/HomeAccountActions";
import "../home.css";

const overviewItems = [
  {
    title: " مدیریت یکپارچه و امن ",
    description: "باشگاه‌، مربی، ورزشکار",
    tone: "green",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <rect x="3" y="3" width="7" height="7" rx="2" />
        <rect x="14" y="3" width="7" height="7" rx="2" />
        <rect x="3" y="14" width="7" height="7" rx="2" />
        <rect x="14" y="14" width="7" height="7" rx="2" />
      </svg>
    ),
  },
  {
    title: "تحت وب",
    description: "دسترسی از همه جا",
    tone: "blue",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="12" cy="12" r="9" />
        <path d="M3 12h18" />
        <path d="M12 3c2.4 2.5 3.7 5.5 3.7 9S14.4 18.5 12 21" />
        <path d="M12 3c-2.4 2.5-3.7 5.5-3.7 9S9.6 18.5 12 21" />
      </svg>
    ),
  },
  {
    title: "رشد مستمر",
    description: "تصمیم بهتر با گزارش‌های کاربردی",
    tone: "purple",
    icon: (
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="M4 19V5" />
        <path d="M4 19h16" />
        <path d="m7 15 4-4 3 2 5-6" />
        <path d="M15.5 7H19v3.5" />
      </svg>
    ),
  },
];

const enamadTrustSeal =
  "<a referrerpolicy='origin' target='_blank' href='https://trustseal.enamad.ir/?id=7602840&Code=5n5Id1FsuaMUVoh9vpxAgsDuZy8H2t3Z'><img referrerpolicy='origin' src='https://trustseal.enamad.ir/logo.aspx?id=7602840&Code=5n5Id1FsuaMUVoh9vpxAgsDuZy8H2t3Z' alt='' style='cursor:pointer' code='5n5Id1FsuaMUVoh9vpxAgsDuZy8H2t3Z'></a>";

export default async function Home({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  return (
    <main className="home">
      <header>
        <SiteLogo href={`/${locale}`} priority />
        <nav>
          <a href="#features">امکانات</a>
          <a href="#gyms">باشگاه‌ها</a>
        </nav>
        <HomeAccountActions locale={locale} />
      </header>
      <GymShowcase locale={locale} />
      <section className="home-hero">
        <div>
          <span>مدیریت حرفه‌ای باشگاه ورزشی</span>
          <h1>
            باشگاهت رو
            <br />
            <em>هوشمندتر</em> مدیریت کن.
          </h1>
          <p>ClubMax فضای یکپارچه مربی، ورزشکاره؛ از ثبت اعضا تا برنامه‌ها و پرداخت‌ها.</p>
          <div>
            <Link href={`/${locale}/signup`}>شروع ثبت‌نام ←</Link>
            <Link href={`/${locale}/login`}>حساب کاربری دارم</Link>
          </div>
        </div>
        <aside>
          <div className="hero-card">
            <span>در یک نگاه</span>
            <strong>کلاب ماکس</strong>
            {overviewItems.map((item) => (
              <div className="hero-card-item" key={item.title}>
                <span className={`hero-card-icon ${item.tone}`}>{item.icon}</span>
                <p>
                  {item.title}
                  <small>{item.description}</small>
                </p>
              </div>
            ))}
          </div>
        </aside>
      </section>
      <section id="features" className="home-section">
        <h2>ساخته‌شده برای دنیای واقعی باشگاه</h2>
        <div className="feature-grid">
          {[
            [
              "مربی",
              "ثبت باشگاه، مدیریت ورزشکاران، مدیریت شهریه ها، سیستم حضور و غیاب(به زودی) و...",
            ],
            [
              "ورزشکار",
              "عضویت در باشگاه‌ها و دسترسی آسان به تمامی باشگاه های ورزشی ثبت شده، مقایسه آسان،پرداخت شهریه بدون کارمزد و...",
            ],
          ].map((x) => (
            <article key={x[1]}>
              <i>{x[0]}</i>
              <h3>{x[1]}</h3>
              <p>{x[2]}</p>
            </article>
          ))}
        </div>
      </section>

      <footer>
        <SiteLogo href={`/${locale}`} priority />
        <p>سامانه مدیریت باشگاه‌های ورزشی</p>
        <nav className="home-footer-links" aria-label="اطلاعات و قوانین">
          <Link href={`/${locale}/about`}>درباره ما</Link>
          <Link href={`/${locale}/contact`}>تماس با ما</Link>
          <Link href={`/${locale}/terms`}>قوانین و مقررات</Link>
          <Link href={`/${locale}/privacy`}> حریم خصوصی</Link>
          <Link href={`/${locale}/refunds`}>لغو و استرداد وجه</Link>
          <Link href={`/${locale}/pricing`}>خدمات و قیمت‌ها</Link>
        </nav>
        <span dangerouslySetInnerHTML={{ __html: enamadTrustSeal }} />
      </footer>
    </main>
  );
}
