"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import SiteLogo from "../../components/SiteLogo";
import GymCover from "../../components/GymCover";
import "../../catalog.css";

const API = process.env.NEXT_PUBLIC_API_URL!;
const days = ["شنبه", "یکشنبه", "دوشنبه", "سه‌شنبه", "چهارشنبه", "پنجشنبه", "جمعه"];
const normalizeSearch = (value: string) =>
  value.trim().toLocaleLowerCase("fa").replace(/ي/g, "ی").replace(/ك/g, "ک");
type Gym = {
  id: string;
  name: string;
  coach_name: string;
  province: string;
  city: string;
  address: string;
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
};

export default function ExploreGyms() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter();
  const [items, setItems] = useState<Gym[]>([]),
    [error, setError] = useState(""),
    [filters, setFilters] = useState({
      q: "",
      coach: "",
      province: "",
      city: "",
      sport: "",
      style: "",
      mode: "",
    });
  useEffect(() => {
    const token = localStorage.getItem("clubmax_access_token") || "";
    fetch(`${API}/v1/gyms/catalog`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async (response) => {
        if (response.status === 401 || response.status === 403) {
          router.replace(`/${locale}/login`);
          return;
        }
        const data = await response.json();
        if (!response.ok) throw new Error(data.message);
        setItems(data || []);
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت باشگاه‌ها انجام نشد.")
      );
  }, [locale, router]);
  const values = (key: keyof Gym) => [
    ...new Set(items.map((item) => String(item[key] ?? "")).filter(Boolean)),
  ];
  const shown = useMemo(
    () =>
      items.filter((item) => {
        const all = normalizeSearch(
          [
            item.name,
            item.coach_name,
            item.province,
            item.city,
            item.address,
            item.sport,
            item.sport_style,
            item.tuition_amount,
            ...item.active_weekdays.map((day) => days[day]),
          ].join(" ")
        );
        return (
          (!filters.q || all.includes(normalizeSearch(filters.q))) &&
          (!filters.coach ||
            normalizeSearch(item.coach_name).includes(normalizeSearch(filters.coach))) &&
          (!filters.province || item.province === filters.province) &&
          (!filters.city || item.city === filters.city) &&
          (!filters.sport ||
            normalizeSearch(item.sport).includes(normalizeSearch(filters.sport))) &&
          (!filters.style ||
            normalizeSearch(item.sport_style).includes(normalizeSearch(filters.style))) &&
          (!filters.mode || item.tuition_mode === filters.mode)
        );
      }),
    [items, filters]
  );
  const change = (key: keyof typeof filters, value: string) =>
    setFilters((current) => ({ ...current, [key]: value }));
  return (
    <main className="catalog-page">
      <header>
        <SiteLogo href={`/${locale}`} priority />
        <Link href={`/${locale}/dashboard`}>بازگشت به داشبورد ←</Link>
      </header>
      <section className="catalog-title">
        <span className="eyebrow">انتخاب باشگاه</span>
        <h1>باشگاه مناسب خود را پیدا کنید</h1>
        <p>مشخصات باشگاه‌ها را مقایسه و پس از انتخاب، ثبت‌نام و پرداخت شهریه را تکمیل کنید.</p>
      </section>
      <section className="catalog-filters">
        <input
          value={filters.q}
          onChange={(event) => change("q", event.target.value)}
          placeholder="جست‌وجو در تمام مشخصات..."
        />
        <input
          value={filters.coach}
          onChange={(event) => change("coach", event.target.value)}
          placeholder="جست‌وجوی نام مربی..."
          list="coach-options"
          autoComplete="off"
        />
        <datalist id="coach-options">
          {values("coach_name").map((name) => (
            <option key={name} value={name} />
          ))}
        </datalist>
        <Filter
          label="استان"
          value={filters.province}
          options={values("province")}
          onChange={(value) => change("province", value)}
        />
        <Filter
          label="شهرستان"
          value={filters.city}
          options={values("city")}
          onChange={(value) => change("city", value)}
        />
        <input
          value={filters.sport}
          onChange={(event) => change("sport", event.target.value)}
          placeholder="رشته ورزشی..."
        />
        <input
          value={filters.style}
          onChange={(event) => change("style", event.target.value)}
          placeholder="سبک رشته..."
        />
        <Filter
          label="نوع شهریه"
          value={filters.mode}
          options={["monthly", "per_session", "course"]}
          titles={{ monthly: "ماهیانه", per_session: "جلسه‌ای", course: "دوره‌ای" }}
          onChange={(value) => change("mode", value)}
        />
      </section>
      {error && <div className="alert">{error}</div>}
      <section className="catalog-grid">
        {shown.map((gym) => (
          <article key={gym.id}>
            <GymCover gymId={gym.id} />
            <div>
              <span>
                {gym.sport} · {gym.sport_style}
              </span>
              <h2>{gym.name}</h2>
              <p>مربی: {gym.coach_name}</p>
              <small>
                {gym.province}، {gym.city}
              </small>
              <small className="activity-days">
                روزهای فعالیت: {gym.active_weekdays.map((day) => days[day]).join("، ")}
              </small>
              <small className={`capacity-status ${gym.capacity_remaining === 0 ? "full" : ""}`}>
                ظرفیت باقی‌مانده:{" "}
                {gym.capacity_remaining === null
                  ? "نامحدود"
                  : `${gym.capacity_remaining.toLocaleString("fa-IR")} نفر`}
              </small>
              <strong>{gym.tuition_amount.toLocaleString("fa-IR")} ریال</strong>
              <Link
                className={gym.capacity_remaining === 0 && !gym.joined ? "disabled-link" : ""}
                href={`/${locale}/explore-gyms/${gym.id}`}
              >
                {gym.joined
                  ? "مشاهده عضویت"
                  : gym.capacity_remaining === 0
                    ? "تکمیل ظرفیت"
                    : "مشاهده و ثبت‌نام"}
              </Link>
            </div>
          </article>
        ))}
        {!shown.length && !error && (
          <div className="catalog-empty">باشگاهی مطابق فیلترها پیدا نشد.</div>
        )}
      </section>
    </main>
  );
}
function Filter({
  label,
  value,
  options,
  titles = {},
  onChange,
}: {
  label: string;
  value: string;
  options: string[];
  titles?: Record<string, string>;
  onChange: (value: string) => void;
}) {
  return (
    <select value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="">{label}: همه</option>
      {options.map((option) => (
        <option key={option} value={option}>
          {titles[option] || option}
        </option>
      ))}
    </select>
  );
}
