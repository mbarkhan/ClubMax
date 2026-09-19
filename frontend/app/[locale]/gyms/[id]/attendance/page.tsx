"use client";

import Link from "next/link";
import { use, useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import JalaliDateInput from "../../../../components/JalaliDateInput";
import "../../../../catalog.css";
import "../../../../roster.css";
import "../../../../attendance.css";
import "../../../../attendance-overrides.css";

const API = "";
type Athlete = { id: string; first_name: string; last_name: string; present: boolean };
const today = () => {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
};

export default function GymAttendance({
  params,
}: {
  params: Promise<{ locale: string; id: string }>;
}) {
  const { locale, id } = use(params),
    router = useRouter();
  const [date, setDate] = useState(today),
    [items, setItems] = useState<Athlete[]>([]),
    [error, setError] = useState(""),
    [pending, setPending] = useState<Set<string>>(new Set());
  const versions = useRef(new Map<string, number>()),
    timers = useRef(new Map<string, ReturnType<typeof setTimeout>>()),
    controllers = useRef(new Map<string, AbortController>()),
    loader = useRef<AbortController | null>(null);
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  const load = useCallback(async () => {
    loader.current?.abort();
    const controller = new AbortController();
    loader.current = controller;
    setError("");
    try {
      const response = await fetch(`${API}/v1/gyms/${id}/attendance?date=${date}`, {
        headers: { Authorization: `Bearer ${token()}` },
        signal: controller.signal,
      });
      if (response.status === 401 || response.status === 403) {
        router.replace(`/${locale}/login`);
        return;
      }
      const data = await response.json();
      if (!response.ok) throw new Error(data.message || "دریافت حضور و غیاب انجام نشد.");
      setItems(data || []);
    } catch (reason) {
      if ((reason as Error).name !== "AbortError")
        setError(reason instanceof Error ? reason.message : "دریافت حضور و غیاب انجام نشد.");
    }
  }, [date, id, locale, router]);
  useEffect(() => {
    load();
    return () => loader.current?.abort();
  }, [load]);
  useEffect(
    () => () => {
      timers.current.forEach(clearTimeout);
      controllers.current.forEach((controller) => controller.abort());
    },
    []
  );
  const toggle = (athlete: Athlete) => {
    const present = !athlete.present,
      version = (versions.current.get(athlete.id) || 0) + 1;
    versions.current.set(athlete.id, version);
    setError("");
    setItems((current) =>
      current.map((item) => (item.id === athlete.id ? { ...item, present } : item))
    );
    setPending((current) => new Set(current).add(athlete.id));
    const oldTimer = timers.current.get(athlete.id);
    if (oldTimer) clearTimeout(oldTimer);
    timers.current.set(
      athlete.id,
      setTimeout(async () => {
        controllers.current.get(athlete.id)?.abort();
        const controller = new AbortController();
        controllers.current.set(athlete.id, controller);
        try {
          const response = await fetch(`${API}/v1/gyms/${id}/attendance/${athlete.id}`, {
            method: "PUT",
            headers: { Authorization: `Bearer ${token()}`, "Content-Type": "application/json" },
            body: JSON.stringify({ date, present }),
            signal: controller.signal,
          });
          const data = await response.json();
          if (!response.ok) throw new Error(data.message || "ثبت حضور و غیاب انجام نشد.");
        } catch (reason) {
          if (
            (reason as Error).name !== "AbortError" &&
            versions.current.get(athlete.id) === version
          ) {
            setItems((current) =>
              current.map((item) =>
                item.id === athlete.id ? { ...item, present: !present } : item
              )
            );
            setError(reason instanceof Error ? reason.message : "ثبت حضور و غیاب انجام نشد.");
          }
        } finally {
          if (versions.current.get(athlete.id) === version)
            setPending((current) => {
              const updated = new Set(current);
              updated.delete(athlete.id);
              return updated;
            });
        }
      }, 300)
    );
  };
  return (
    <main className="catalog-page attendance-page">
      <header>
        <Link href={`/${locale}/gyms`}>بازگشت به باشگاه‌های من ←</Link>
      </header>
      <section className="wallet-ledger roster-card">
        <div className="roster-head attendance-head">
          <div>
            <span>مدیریت حضور و غیاب</span>
            <h1>حضور ورزشکاران</h1>
          </div>
          <JalaliDateInput
            key={date}
            name="attendance-date"
            label="تاریخ"
            initialValue={`${date}T00:00:00Z`}
            minYear={1400}
            maxYear={1450}
            allowClear={false}
            maxDate={today()}
            onChange={setDate}
          />
        </div>
        {error && <div className="alert">{error}</div>}
        <div className="attendance-list">
          {items.map((athlete) => (
            <label
              className={`attendance-row${athlete.present ? " present" : ""}`}
              key={athlete.id}
            >
              <input type="checkbox" checked={athlete.present} onChange={() => toggle(athlete)} />
              <span className="attendance-check" aria-hidden="true">
                ✓
              </span>
              <b>
                {athlete.first_name} {athlete.last_name}
              </b>
              {pending.has(athlete.id) && <small>در حال ثبت…</small>}
            </label>
          ))}
          {!items.length && !error && (
            <div className="roster-empty">هنوز ورزشکاری در این باشگاه ثبت‌نام نکرده است.</div>
          )}
        </div>
      </section>
    </main>
  );
}
