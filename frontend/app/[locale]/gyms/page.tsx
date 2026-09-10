"use client";
import Link from "next/link";
import SiteLogo from "../../components/SiteLogo";
import { FormEvent, useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import LocationFields from "../../components/LocationFields";
import GymImageFields from "../../components/GymImageFields";
import GymCover from "../../components/GymCover";
import { Time24Field, TuitionAmountField } from "../../components/GymInputs";
import "../../auth.css";
const API = "",
  days = ["شنبه", "یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه"];
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
  tuition_mode: string;
  tuition_amount: number;
  capacity: number;
};
export default function Gyms() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter();
  const [items, setItems] = useState<Gym[]>([]),
    [role, setRole] = useState(""),
    [error, setError] = useState(""),
    [open, setOpen] = useState(false),
    [sport, setSport] = useState("");
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  async function load() {
    const r = await fetch(`${API}/v1/gyms`, { headers: { Authorization: `Bearer ${token()}` } });
    if (r.status === 401) {
      router.replace(`/${locale}/login`);
      return;
    }
    const data = await r.json();
    if (r.ok) setItems(data || []);
    else setError(data.message);
  }
  useEffect(() => {
    setRole(JSON.parse(localStorage.getItem("clubmax_user") || "{}").role || "");
    load();
  }, []);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError("");
    const form = e.currentTarget,
      f = new FormData(form),
      images = f.getAll("images").filter((x) => x instanceof File && x.size) as File[];
    if (images.length > 3) {
      setError("حداکثر سه تصویر قابل انتخاب است.");
      return;
    }
    const body = {
      name: f.get("name"),
      province: f.get("province"),
      city: f.get("city"),
      address: f.get("address"),
      phone: f.get("phone") || null,
      sport: f.get("sport"),
      sport_style: f.get("sport_style"),
      active_weekdays: f.getAll("weekdays").map(Number),
      start_time: f.get("start_time"),
      end_time: f.get("end_time"),
      tuition_mode: f.get("tuition_mode"),
      tuition_amount: Number(f.get("tuition_amount")),
      capacity: Number(f.get("capacity") || 0),
    };
    const r = await fetch(`${API}/v1/gyms`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token()}`,
          "Content-Type": "application/json",
          "Accept-Language": "fa",
        },
        body: JSON.stringify(body),
      }),
      result = await r.json();
    if (!r.ok) {
      setError(result.message || "ثبت باشگاه انجام نشد.");
      return;
    }
    for (const image of images) {
      const upload = new FormData();
      upload.append("image", image);
      const sent = await fetch(`${API}/v1/gyms/${result.id}/images`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token()}` },
        body: upload,
      });
      if (!sent.ok) {
        setError("باشگاه ثبت شد، اما بارگذاری یکی از تصاویر انجام نشد.");
        break;
      }
    }
    form.reset();
    setSport("");
    setOpen(false);
    load();
  }
  async function remove(id: string) {
    if (!confirm("این باشگاه حذف شود؟")) return;
    const r = await fetch(`${API}/v1/gyms/${id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token()}` },
    });
    if (r.ok) setItems((v) => v.filter((x) => x.id !== id));
    else setError("حذف باشگاه انجام نشد.");
  }
  return (
    <main className="signup-page gyms-page">
      <header>
        <SiteLogo href={`/${locale}`} priority />
        <Link href={`/${locale}/dashboard`}>بازگشت به داشبورد ←</Link>
      </header>
      <section className="gyms-head">
        <div>
          <span className="eyebrow">مدیریت باشگاه</span>
          <h1>باشگاه‌های من</h1>
          <p>
            {role === "coach"
              ? "باشگاه‌ها، زمان فعالیت و شهریه را مدیریت کنید."
              : "باشگاه‌هایی که عضو آن‌ها هستید."}
          </p>
        </div>
        {role === "coach" && (
          <button className="submit" onClick={() => setOpen((v) => !v)}>
            {open ? "بستن فرم" : "＋ ثبت باشگاه"}
          </button>
        )}
      </section>
      {error && <div className="alert">{error}</div>}
      {open && (
        <form className="signup-card gym-form" onSubmit={submit}>
          <h3>مشخصات باشگاه</h3>
          <div className="form-grid">
            <label>
              نام باشگاه <b className="required">*</b>
              <input name="name" required />
            </label>
            <label>
              شماره تماس
              <input name="phone" dir="ltr" />
            </label>
            <LocationFields required />
            <label>
              آدرس کامل <b className="required">*</b>
              <input name="address" required />
            </label>
            <label>
              رشته ورزشی <b className="required">*</b>
              <input
                name="sport"
                required
                list="sport-suggestions"
                value={sport}
                onChange={(event) => setSport(event.target.value)}
                autoComplete="off"
              />
              <datalist id="sport-suggestions">
                {[...new Set(items.map((item) => item.sport))].map((value) => (
                  <option key={value} value={value} />
                ))}
              </datalist>
            </label>
            <label>
              سبک رشته ورزشی <b className="required">*</b>
              <input name="sport_style" required list="style-suggestions" autoComplete="off" />
              <datalist id="style-suggestions">
                {[
                  ...new Set(
                    items
                      .filter((item) => !sport || item.sport === sport)
                      .map((item) => item.sport_style)
                  ),
                ].map((value) => (
                  <option key={value} value={value} />
                ))}
              </datalist>
            </label>
            <label>
              نحوه محاسبه شهریه
              <select name="tuition_mode">
                <option value="monthly">ماهیانه</option>
                <option value="per_session">جلسه‌ای</option>
                <option value="course">دوره‌ای</option>
              </select>
            </label>
            <TuitionAmountField />
            <label>
              ظرفیت ورزشکاران
              <input name="capacity" required type="number" min="0" step="1" defaultValue="0" />
              <small>عدد صفر یعنی ظرفیت نامحدود</small>
            </label>
            <div className="gym-time-row">
              <Time24Field name="start_time" label="ساعت شروع" />
              <Time24Field name="end_time" label="ساعت پایان" />
            </div>
          </div>
          <h3 className="gym-images-title">تصاویر باشگاه</h3>
          <GymImageFields />
          <div className="days-field">
            <b>روزهای فعالیت *</b>
            <div>
              {days.map((day, index) => (
                <label key={day}>
                  <input type="checkbox" name="weekdays" value={index} />
                  <span>{day}</span>
                </label>
              ))}
            </div>
          </div>
          <button className="submit signup-submit">ثبت باشگاه</button>
        </form>
      )}
      {!open && (
        <section className="gym-grid">
          {items.length ? (
            items.map((g) => (
              <article className="gym-card" key={g.id}>
                <GymCover gymId={g.id} />
                {role === "coach" && (
                  <button
                    type="button"
                    className="gym-card-delete"
                    onClick={() => remove(g.id)}
                    aria-label={`حذف ${g.name}`}
                    title="حذف باشگاه"
                  >
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                      <path d="M4 7h16M9 7V4h6v3m3 0-1 13H7L6 7m4 4v5m4-5v5" />
                    </svg>
                  </button>
                )}
                <Link className="gym-card-main" href={`/${locale}/gyms/${g.id}/athletes`}>
                  <div>
                    <span>
                      {g.sport} · {g.sport_style}
                    </span>
                    <h2>{g.name}</h2>
                    <p>
                      {g.province}، {g.city}
                    </p>
                    <small>
                      {g.active_weekdays.map((d) => days[d]).join("، ")} ·{" "}
                      {g.start_time.slice(0, 5)} تا {g.end_time.slice(0, 5)}
                    </small>
                    <strong>{g.tuition_amount.toLocaleString("fa-IR")} ریال</strong>
                  </div>
                </Link>
                {role === "coach" && (
                  <div className="gym-card-actions">
                    <Link
                      href={`/${locale}/gyms/${g.id}/edit`}
                      aria-label={`ویرایش ${g.name}`}
                      title="ویرایش باشگاه"
                    >
                      ✎
                    </Link>
                  </div>
                )}
              </article>
            ))
          ) : (
            <div className="empty-card">
              <h3>هنوز باشگاهی ثبت نشده است</h3>
              <p>
                {role === "coach"
                  ? "برای شروع روی ثبت باشگاه بزنید."
                  : "پس از عضویت، باشگاه اینجا نمایش داده می‌شود."}
              </p>
            </div>
          )}
        </section>
      )}
    </main>
  );
}
