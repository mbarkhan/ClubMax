"use client";

import Link from "next/link";
import { use, useCallback, useEffect, useState } from "react";
import "../../../catalog.css";

const API = "";
const days = ["شنبه", "یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه"];
const modes: Record<string, string> = {
  monthly: "ماهیانه",
  per_session: "جلسه‌ای",
  course: "دوره‌ای",
};
type Gym = {
  id: string;
  name: string;
  coach_name: string;
  province: string;
  city: string;
  address: string;
  phone?: string;
  sport: string;
  sport_style: string;
  active_weekdays: number[];
  start_time: string;
  end_time: string;
  tuition_mode: string;
  tuition_amount: number;
  capacity: number;
  capacity_remaining: number | null;
  joined: boolean;
  registration_paid: boolean;
  join_request_status?: "pending" | "approved" | "rejected";
  join_request_rejection_reason?: string;
  unpaid_months: number;
  outstanding_amount: number;
  monthly_payment_due: boolean;
};
type Image = { id: string; storage_key: string };

export default function GymDetail({ params }: { params: Promise<{ locale: string; id: string }> }) {
  const { locale, id } = use(params);
  const [gym, setGym] = useState<Gym | null>(null),
    [images, setImages] = useState<Image[]>([]),
    [error, setError] = useState(""),
    [paying, setPaying] = useState(false),
    [requesting, setRequesting] = useState(false),
    [authenticated, setAuthenticated] = useState(false),
    [isAthlete, setIsAthlete] = useState(false);
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  const load = useCallback(() => {
    const accessToken = token();
    setAuthenticated(Boolean(accessToken));
    void (async () => {
      try {
        const [detail, pictures] = await Promise.all([
          fetch(`${API}/v1/public/gyms/catalog/${id}`),
          fetch(`${API}/v1/public/gyms/${id}/images`),
        ]);
        let data = await detail.json();
        if (!detail.ok) throw new Error(data.message);
        if (pictures.ok) setImages(await pictures.json());
        if (accessToken) {
          const me = await fetch(`${API}/v1/users/me`, {
            headers: { Authorization: `Bearer ${accessToken}` },
          });
          if (me.ok) {
            const user = await me.json();
            const athlete = user.role === "athlete";
            setIsAthlete(athlete);
            localStorage.setItem("clubmax_user", JSON.stringify(user));
            if (athlete) {
              const athleteDetail = await fetch(`${API}/v1/gyms/catalog/${id}`, {
                headers: { Authorization: `Bearer ${accessToken}` },
              });
              if (athleteDetail.ok) data = await athleteDetail.json();
            }
          }
        }
        setGym(data);
      } catch (reason) {
        setError(reason instanceof Error ? reason.message : "دریافت اطلاعات باشگاه انجام نشد.");
      }
    })();
  }, [id]);
  useEffect(load, [load]);
  async function payMonthlyDues() {
    setPaying(true);
    setError("");
    try {
      const response = await fetch(`${API}/v1/gyms/${id}/monthly-dues/pay`, {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token()}`,
            "Content-Type": "application/json",
            "Accept-Language": "fa",
          },
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "ایجاد پرداخت انجام نشد.");
      window.location.assign(data.payment_url);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "پرداخت شهریه انجام نشد.");
    } finally {
      setPaying(false);
    }
  }
  async function requestMembership() {
    setRequesting(true);
    setError("");
    try {
      const response = await fetch(`${API}/v1/gyms/${id}/join-requests`, {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token()}`,
            "Accept-Language": "fa",
          },
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "ارسال درخواست عضویت انجام نشد.");
      load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ارسال درخواست عضویت انجام نشد.");
    } finally {
      setRequesting(false);
    }
  }
  if (!gym)
    return (
      <main className="catalog-page">
        <div className="screen-loader">
          <span>{error || "در حال دریافت اطلاعات باشگاه..."}</span>
        </div>
      </main>
    );
  return (
    <main className="catalog-page">
      <header>
        <Link href={`/${locale}/explore-gyms`}>بازگشت به باشگاه‌ها ←</Link>
      </header>
      <section className="gym-detail">
        <div className="gym-gallery">
          {images.length ? (
            images.map((image) => (
              <img key={image.id} src={`${API}/media/${image.storage_key}`} alt={gym.name} />
            ))
          ) : (
            <div>تصویری ثبت نشده است</div>
          )}
        </div>
        <div className="gym-detail-info">
          <span>
            {gym.sport} · {gym.sport_style}
          </span>
          <h1>{gym.name}</h1>
          <h3>مربی {gym.coach_name}</h3>
          <dl>
            <dt>نشانی</dt>
            <dd>
              {gym.province}، {gym.city}، {gym.address}
            </dd>
            <dt>روزهای فعالیت</dt>
            <dd>{gym.active_weekdays.map((day) => days[day]).join("، ")}</dd>
            <dt>ساعت</dt>
            <dd>
              {gym.start_time.slice(0, 5)} تا {gym.end_time.slice(0, 5)}
            </dd>
            <dt>نوع شهریه</dt>
            <dd>{modes[gym.tuition_mode]}</dd>
            <dt>مبلغ شهریه</dt>
            <dd>{gym.tuition_amount.toLocaleString("fa-IR")} ریال</dd>
            <dt>ظرفیت باقی‌مانده</dt>
            <dd>
              {gym.capacity_remaining === null
                ? "نامحدود"
                : `${gym.capacity_remaining.toLocaleString("fa-IR")} نفر`}
            </dd>
          </dl>
          {error && <div className="alert membership-alert">{error}</div>}
          {gym.joined && !gym.registration_paid ? (
            <section className="membership-approved">
              <span>درخواست شما تأیید شد</span>
              <h2>عضویت شما در باشگاه ثبت شده است.</h2>
              <p>اکنون می‌توانید شهریه ثبت‌نام را از طریق درگاه امن پرداخت کنید.</p>
              <Link className="pay-link" href={`/${locale}/explore-gyms/${id}/checkout`}>
                پرداخت شهریه ثبت‌نام
              </Link>
            </section>
          ) : gym.joined && gym.monthly_payment_due ? (
            <section className="monthly-dues">
              <span>شهریه پرداخت‌نشده</span>
              <h2>{gym.unpaid_months.toLocaleString("fa-IR")} ماه</h2>
              <div>
                <small>جمع بدهی</small>
                <strong>{gym.outstanding_amount.toLocaleString("fa-IR")} ریال</strong>
              </div>
              <p>پس از انتخاب پرداخت، برای تکمیل تراکنش به درگاه امن زیبال منتقل می‌شوید.</p>
              <button onClick={payMonthlyDues} disabled={paying}>
                {paying ? "در حال انتقال به درگاه..." : "پرداخت تمام شهریه‌های معوق"}
              </button>
            </section>
          ) : gym.joined ? (
            <div className="joined-badge">شهریه شما تا پایان ماه جاری تسویه است.</div>
          ) : gym.join_request_status === "pending" ? (
            <div className="membership-request-status pending">
              <strong>در انتظار بررسی مربی</strong>
              <span>پس از تأیید مربی، امکان پرداخت شهریه برای شما فعال می‌شود.</span>
            </div>
          ) : gym.join_request_status === "rejected" ? (
            <div className="membership-request-status rejected">
              <strong>درخواست قبلی رد شده است</strong>
              {gym.join_request_rejection_reason && (
                <span>{gym.join_request_rejection_reason}</span>
              )}
              {gym.capacity_remaining === 0 ? (
                <span>در حال حاضر ظرفیت باشگاه تکمیل است و امکان درخواست مجدد وجود ندارد.</span>
              ) : (
                <button type="button" onClick={requestMembership} disabled={requesting}>
                  {requesting ? "در حال ارسال..." : "ارسال دوباره درخواست"}
                </button>
              )}
            </div>
          ) : gym.capacity_remaining === 0 ? (
            <div className="capacity-full-message">ظرفیت این باشگاه تکمیل شده است.</div>
          ) : authenticated && isAthlete ? (
            <button
              type="button"
              className="membership-request-button"
              onClick={requestMembership}
              disabled={requesting}
            >
              {requesting ? "در حال ارسال درخواست..." : "ارسال درخواست عضویت"}
            </button>
          ) : authenticated ? (
            <div className="membership-role-message">
              ثبت درخواست عضویت فقط برای حساب ورزشکار فعال است.
            </div>
          ) : (
            <Link className="pay-link" href={`/${locale}/login`}>
              برای ثبت‌نام وارد شوید
            </Link>
          )}
        </div>
      </section>
    </main>
  );
}
