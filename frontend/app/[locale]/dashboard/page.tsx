"use client";
import Link from "next/link";
import GymCover from "../../components/GymCover";
import DashboardSidebar from "../../components/DashboardSidebar";
import BMITrendChart, { bmiStatus, BodyMeasurement } from "../../components/BMITrendChart";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import "../../admin.css";
import "../../bmi.css";
const API = "";
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
type JoinRequest = {
  id: string;
  gym_id: string;
  gym_name: string;
  athlete_id: string;
  athlete_name: string;
  athlete_phone?: string;
  status: "pending" | "approved" | "rejected";
  rejection_reason?: string;
  created_at: string;
};
type BMIOverview = {
  measurements: BodyMeasurement[];
  latest?: BodyMeasurement;
  can_record: boolean;
  next_allowed_at?: string;
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
        <AdminModule
          icon="ذ"
          title="پشتیبان‌گیری"
          text="مشاهده وضعیت پشتیبان‌های زمان‌بندی‌شده و دریافت آرشیو کامل پایگاه داده و فایل‌ها."
          href={`/${locale}/admin/backups`}
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
  return <CoachWorkspace locale={locale} />;
}
function CoachWorkspace({ locale }: { locale: string }) {
  const [requests, setRequests] = useState<JoinRequest[]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [busy, setBusy] = useState(""),
    [rejecting, setRejecting] = useState(""),
    [reasons, setReasons] = useState<Record<string, string>>({});
  const load = () => {
    const token = localStorage.getItem("clubmax_access_token") || "";
    fetch(`${API}/v1/gym-join-requests?status=pending`, {
      headers: { Authorization: `Bearer ${token}`, "Accept-Language": "fa" },
    })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok) throw new Error(data.message);
        setRequests(data || []);
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت درخواست‌ها انجام نشد.")
      );
  };
  useEffect(load, []);
  async function review(request: JoinRequest, status: "approved" | "rejected") {
    const reason = reasons[request.id]?.trim();
    if (status === "rejected" && !reason) {
      setError("برای رد درخواست، علت رد را بنویسید.");
      return;
    }
    setBusy(request.id);
    setError("");
    setNotice("");
    try {
      const response = await fetch(`${API}/v1/gym-join-requests/${request.id}/review`, {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
            "Content-Type": "application/json",
            "Accept-Language": "fa",
          },
          body: JSON.stringify({
            status,
            rejection_reason: status === "rejected" ? reason : undefined,
          }),
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "بررسی درخواست انجام نشد.");
      setRequests((items) => items.filter((item) => item.id !== request.id));
      setRejecting("");
      setNotice(
        status === "approved"
          ? `عضویت ${request.athlete_name} تأیید شد.`
          : `درخواست ${request.athlete_name} رد شد.`
      );
    } catch (reasonValue) {
      setError(reasonValue instanceof Error ? reasonValue.message : "بررسی درخواست انجام نشد.");
    } finally {
      setBusy("");
    }
  }
  return (
    <div className="coach-workspace">
      <section className="join-request-panel">
        <div className="card-head">
          <div>
            <span>کارتابل عضویت</span>
            <h3>درخواست‌های در انتظار بررسی</h3>
          </div>
          <b>{requests.length.toLocaleString("fa-IR")} درخواست</b>
        </div>
        {error && <div className="alert">{error}</div>}
        {notice && <div className="join-request-notice">{notice}</div>}
        <div className="join-request-list">
          {requests.map((request) => (
            <article className="join-request-card" key={request.id}>
              <div className="join-request-person">
                <span>{request.athlete_name.slice(0, 1)}</span>
                <div>
                  <h4>{request.athlete_name}</h4>
                  <p>{request.athlete_phone || "شماره تماس ثبت نشده"}</p>
                </div>
              </div>
              <div className="join-request-gym">
                <small>باشگاه درخواستی</small>
                <strong>{request.gym_name}</strong>
                <time>{new Date(request.created_at).toLocaleDateString("fa-IR")}</time>
              </div>
              {rejecting === request.id ? (
                <div className="join-reject-form">
                  <label htmlFor={`reject-${request.id}`}>علت رد درخواست</label>
                  <textarea
                    id={`reject-${request.id}`}
                    maxLength={500}
                    value={reasons[request.id] || ""}
                    onChange={(event) =>
                      setReasons((values) => ({ ...values, [request.id]: event.target.value }))
                    }
                    placeholder="علت را برای ورزشکار توضیح دهید..."
                  />
                  <div>
                    <button
                      type="button"
                      className="reject-confirm"
                      disabled={busy === request.id}
                      onClick={() => review(request, "rejected")}
                    >
                      ثبت رد درخواست
                    </button>
                    <button
                      type="button"
                      className="cancel-review"
                      onClick={() => setRejecting("")}
                    >
                      انصراف
                    </button>
                  </div>
                </div>
              ) : (
                <div className="join-request-actions">
                  <button
                    type="button"
                    className="approve-request"
                    disabled={Boolean(busy)}
                    onClick={() => review(request, "approved")}
                  >
                    {busy === request.id ? "در حال ثبت..." : "تأیید عضویت"}
                  </button>
                  <button
                    type="button"
                    className="reject-request"
                    disabled={Boolean(busy)}
                    onClick={() => setRejecting(request.id)}
                  >
                    رد درخواست
                  </button>
                </div>
              )}
            </article>
          ))}
          {!requests.length && !error && (
            <div className="join-request-empty">
              <strong>درخواست تازه‌ای ندارید</strong>
              <span>درخواست‌های عضویت ورزشکاران اینجا نمایش داده می‌شوند.</span>
            </div>
          )}
        </div>
      </section>
      <section className="coach-quick-link">
        <div>
          <span>مدیریت باشگاه</span>
          <h3>باشگاه‌ها، اعضا و شهریه‌ها</h3>
        </div>
        <Link href={`/${locale}/gyms`}>ورود به باشگاه‌های من ←</Link>
      </section>
    </div>
  );
}
function AthleteWorkspace({ locale }: { locale: string }) {
  const [gyms, setGyms] = useState<Gym[]>([]),
    [requests, setRequests] = useState<JoinRequest[]>([]),
    [bmi, setBMI] = useState<BMIOverview | null>(null),
    [error, setError] = useState("");
  useEffect(() => {
    const token = localStorage.getItem("clubmax_access_token") || "";
    Promise.all([
      fetch(`${API}/v1/gyms/catalog`, { headers: { Authorization: `Bearer ${token}` } }),
      fetch(`${API}/v1/gym-join-requests`, { headers: { Authorization: `Bearer ${token}` } }),
      fetch(`${API}/v1/athlete/bmi`, { headers: { Authorization: `Bearer ${token}` } }),
    ])
      .then(async ([gymResponse, requestResponse, bmiResponse]) => {
        const gymData = await gymResponse.json(),
          requestData = await requestResponse.json(),
          bmiData = await bmiResponse.json();
        if (!gymResponse.ok) throw new Error(gymData.message);
        if (!requestResponse.ok) throw new Error(requestData.message);
        if (!bmiResponse.ok) throw new Error(bmiData.message);
        setGyms((gymData || []).filter((gym: Gym) => gym.joined));
        setRequests(
          (requestData || []).filter((request: JoinRequest) => request.status !== "approved")
        );
        setBMI(bmiData);
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت باشگاه‌ها انجام نشد.")
      );
  }, []);
  return (
    <div className="athlete-workspace">
      {bmi && (
        <section className="dashboard-bmi-card">
          <div className="card-head">
            <div>
              <span>پایش سلامت</span>
              <h3>روند شاخص توده بدنی</h3>
            </div>
            <div>
              {bmi.latest && (
                <strong
                  className="dashboard-bmi-value"
                  style={{ color: (bmiStatus[bmi.latest.status] || bmiStatus.age_specific).color }}
                >
                  {bmi.latest.bmi.toLocaleString("fa-IR")}
                </strong>
              )}
              <Link href={`/${locale}/tools/bmi`}>
                {bmi.measurements.length ? "مشاهده جزئیات" : "ثبت اولین اندازه‌گیری"}
              </Link>
            </div>
          </div>
          <BMITrendChart measurements={bmi.measurements} compact />
        </section>
      )}
      {requests.length > 0 && (
        <section className="athlete-requests">
          <div className="card-head">
            <div>
              <span>پیگیری عضویت</span>
              <h3>درخواست‌های باشگاه</h3>
            </div>
          </div>
          <div className="athlete-request-list">
            {requests.map((request) => (
              <Link href={`/${locale}/explore-gyms/${request.gym_id}`} key={request.id}>
                <div>
                  <strong>{request.gym_name}</strong>
                  {request.status === "rejected" && request.rejection_reason && (
                    <small>{request.rejection_reason}</small>
                  )}
                </div>
                <span className={`join-status ${request.status}`}>
                  {request.status === "pending" ? "در انتظار بررسی" : "رد شده"}
                </span>
              </Link>
            ))}
          </div>
        </section>
      )}
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
              <p>باشگاه مناسب خود را پیدا کنید و درخواست عضویت بفرستید.</p>
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
