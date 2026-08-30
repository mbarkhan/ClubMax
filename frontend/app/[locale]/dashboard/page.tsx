"use client";
import Link from "next/link";
import GymCover from "../../components/GymCover";
import DashboardSidebar from "../../components/DashboardSidebar";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import "../../admin.css";
const API = process.env.NEXT_PUBLIC_API_URL!;
type User = {
  id: string;
  role: "admin" | "coach" | "athlete";
  first_name: string;
  last_name: string;
  phone?: string;
  national_id?: string;
  active: boolean;
  face_image_id?: string;
};
type Gym = {
  id: string;
  name: string;
  province: string;
  city: string;
  sport: string;
  sport_style: string;
  active_weekdays: number[];
  start_time: string;
  end_time: string;
  tuition_amount: number;
  joined: boolean;
  unpaid_months: number;
  outstanding_amount: number;
  monthly_payment_due: boolean;
};
type Overview = {
  users: number;
  coaches: number;
  athletes: number;
  gyms: number;
  active_athletes: number;
  payment_count: number;
  gross_payments: number;
  platform_revenue: number;
  withdrawal_fees: number;
  pending_withdrawals: number;
};
const days = ["شنبه", "یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه"];
const names = { admin: "مدیر سایت", coach: "مربی", athlete: "ورزشکار" };
export default function Dashboard({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    router = useRouter();
  const [user, setUser] = useState<User | null>(null),
    [overview, setOverview] = useState<Overview | null>(null);
  useEffect(() => {
    const token = localStorage.getItem("clubmax_access_token");
    if (!token) {
      router.replace(`/${locale}/login`);
      return;
    }
    fetch(`${API}/v1/users/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async (r) => {
        if (!r.ok) throw 0;
        const u = await r.json();
        setUser(u);
        localStorage.setItem("clubmax_user", JSON.stringify(u));
        if (u.role === "admin")
          return fetch(`${API}/v1/admin/overview`, {
            headers: { Authorization: `Bearer ${token}` },
          })
            .then((x) => x.json())
            .then(setOverview);
      })
      .catch(() => {
        localStorage.clear();
        router.replace(`/${locale}/login`);
      });
  }, [locale, router]);
  if (!user)
    return (
      <div className="screen-loader">
        <i>G</i>
        <span>در حال آماده‌سازی...</span>
      </div>
    );
  return (
    <div className={`shell ${user.role === "athlete" ? "athlete-shell" : ""}`}>
      <DashboardSidebar user={user} locale={locale} active="dashboard" />
      <main className="dash">
        <header>
          <div>
            <span>داشبورد</span>
            <h1>سلام، {user.first_name} 👋</h1>
          </div>
        </header>

        {user.role === "admin" ? (
          <Admin overview={overview} locale={locale} />
        ) : (
          <Member user={user} locale={locale} />
        )}
      </main>
    </div>
  );
}
function Admin({ overview, locale }: { overview: Overview | null; locale: string }) {
  return (
    <>
      <div className="admin-grid">
        <Metric title="کل کاربران" value={overview?.users || 0} />
        <Metric title="باشگاه‌ها" value={overview?.gyms || 0} />
        <Metric title="ورزشکاران فعال" value={overview?.active_athletes || 0} />
        <Metric title="برداشت‌های باز" value={overview?.pending_withdrawals || 0} />
      </div>
      <div className="admin-modules">
        <AdminModule
          icon="ک"
          title="کاربران و دسترسی‌ها"
          text={`${(overview?.coaches || 0).toLocaleString("fa-IR")} مربی و ${(overview?.athletes || 0).toLocaleString("fa-IR")} ورزشکار؛ مشاهده، ارتقا به مدیر و حذف حساب.`}
          href={`/${locale}/admin/users`}
        />
        <AdminModule
          icon="ب"
          title="باشگاه‌ها"
          text="نمای کامل باشگاه، مربی، اعضای فعال، مبلغ شهریه و گردش پرداخت‌ها."
          href={`/${locale}/admin/gyms`}
        />
        <AdminModule
          icon="پ"
          title="پرداخت شهریه"
          text={`${(overview?.payment_count || 0).toLocaleString("fa-IR")} پرداخت ثبت‌شده با امکان پیگیری پرداخت‌کننده، باشگاه و کد رهگیری.`}
          href={`/${locale}/admin/payments`}
        />
        <AdminModule
          icon="م"
          title="امور مالی و تسویه"
          text="تنظیم کارمزد و حدود برداشت، بررسی درخواست مربیان و ثبت کد پیگیری بانکی."
          href={`/${locale}/admin/finance`}
        />
      </div>
    </>
  );
}
function AdminModule({
  icon,
  title,
  text,
  href,
}: {
  icon: string;
  title: string;
  text: string;
  href: string;
}) {
  return (
    <article className="admin-module">
      <div className="module-icon">{icon}</div>
      <h2>{title}</h2>
      <p>{text}</p>
      <Link href={href}>ورود به بخش ←</Link>
    </article>
  );
}
function Member({ user, locale }: { user: User; locale: string }) {
  if (user.role === "athlete") return <AthleteWorkspace locale={locale} />;
  return (
    <>
      <section className="empty-card">
        <div>◌</div>
        <h3>فضای کاری شما آماده است</h3>
        <p>
          {user.role === "coach"
            ? "برای شروع، اولین باشگاه خود را ثبت کنید."
            : "باشگاه‌هایی که در آن‌ها عضو می‌شوید اینجا نمایش داده خواهند شد."}
        </p>
      </section>
    </>
  );
}
function AthleteWorkspace({ locale }: { locale: string }) {
  const [gyms, setGyms] = useState<Gym[]>([]),
    [error, setError] = useState("");
  useEffect(() => {
    const token = localStorage.getItem("clubmax_access_token") || "";
    fetch(`${API}/v1/gyms/catalog`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.message);
        setGyms((data || []).filter((gym: Gym) => gym.joined));
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت باشگاه‌ها انجام نشد.")
      );
  }, []);
  return (
    <div className="athlete-workspace">
      <section className="athlete-section">
        <div className="card-head">
          <div>
            <span>عضویت‌های فعال</span>
            <h3>باشگاه‌های من</h3>
          </div>
          <Link href={`/${locale}/explore-gyms`}>مشاهده همه باشگاه‌ها</Link>
        </div>
        {error && <div className="alert">{error}</div>}
        <div className="athlete-gym-grid">
          {gyms.map((gym) => (
            <article className={gym.monthly_payment_due ? "has-tuition-due" : ""} key={gym.id}>
              <div className="gym-cover-wrap">
                <GymCover gymId={gym.id} />
                {gym.monthly_payment_due && (
                  <span className="tuition-due-badge">
                    {gym.unpaid_months.toLocaleString("fa-IR")} ماه بدهی شهریه
                  </span>
                )}
              </div>
              <div>
                <span>
                  {gym.sport} · {gym.sport_style}
                </span>
                <h3>{gym.name}</h3>
                <p>
                  {gym.province}، {gym.city}
                </p>
                <small>
                  روزهای فعالیت: {gym.active_weekdays.map((day) => days[day]).join("، ")}
                </small>
                <small>
                  ساعت {gym.start_time.slice(0, 5)} تا {gym.end_time.slice(0, 5)}
                </small>
                {gym.monthly_payment_due && (
                  <strong className="tuition-due-amount">
                    بدهی: {gym.outstanding_amount.toLocaleString("fa-IR")} ریال
                  </strong>
                )}
                <Link href={`/${locale}/explore-gyms/${gym.id}`}>مشاهده باشگاه</Link>
              </div>
            </article>
          ))}
          {!gyms.length && !error && (
            <div className="athlete-empty">
              <h3>هنوز عضو باشگاهی نیستید</h3>
              <p>باشگاه مناسب خود را پیدا کنید و ثبت‌نام را انجام دهید.</p>
              <Link href={`/${locale}/explore-gyms`}>انتخاب باشگاه</Link>
            </div>
          )}
        </div>
      </section>
      <section className="useful-links">
        <span>دسترسی سریع</span>
        <h3>لینک‌های مفید</h3>
        <a href="https://athlete.ifsm.ir/login" target="_blank" rel="noopener noreferrer">
          <div>
            <b>بیمه ورزشی</b>
            <small>ورود به سامانه خدمات درمانی فدراسیون پزشکی ورزشی</small>
          </div>
          <i>↗</i>
        </a>
      </section>
    </div>
  );
}
function Metric({ title, value }: { title: string; value: string | number }) {
  return (
    <article className="admin-stat">
      <span>{title}</span>
      <strong>{value}</strong>
      <small>به‌روزرسانی لحظه‌ای</small>
    </article>
  );
}
