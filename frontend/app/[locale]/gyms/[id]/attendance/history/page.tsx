"use client";

import Link from "next/link";
import { use, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { toGregorian, toJalaali } from "jalaali-js";
import "../../../../../catalog.css";
import "../../../../../roster.css";
import "../../../../../roster-overrides.css";
import "../../../../../print.css";
import "../../../../../attendance-report.css";

const API = "", months = ["فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"];
type Athlete = { id: string; first_name: string; last_name: string };
type Record = { attendance_date: string; athlete_id: string };
const gregorianDate = (year: number, month: number) => { const value = toGregorian(year, month, 1); return `${value.gy}-${String(value.gm).padStart(2, "0")}-${String(value.gd).padStart(2, "0")}`; };

export default function AttendanceHistory({ params }: { params: Promise<{ locale: string; id: string }> }) {
  const { locale, id } = use(params), router = useRouter();
  const now = useMemo(() => { const today = new Date(); return toJalaali(today.getFullYear(), today.getMonth() + 1, today.getDate()); }, []);
  const [year, setYear] = useState(now.jy), [month, setMonth] = useState(now.jm), [athletes, setAthletes] = useState<Athlete[]>([]), [records, setRecords] = useState<Record[]>([]), [query, setQuery] = useState(""), [error, setError] = useState("");
  const moveMonth = (step: number) => { let nextMonth = month + step, nextYear = year; if (nextMonth === 13) { nextMonth = 1; nextYear++; } if (nextMonth === 0) { nextMonth = 12; nextYear--; } setYear(nextYear); setMonth(nextMonth); };
  useEffect(() => {
    const headers = { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` }, from = gregorianDate(year, month), until = gregorianDate(month === 12 ? year + 1 : year, month === 12 ? 1 : month + 1);
    setRecords([]);
    Promise.all([fetch(`${API}/v1/gyms/${id}/athletes`, { headers }), fetch(`${API}/v1/gyms/${id}/attendance/history?from=${from}&until=${until}`, { headers })]).then(async ([members, history]) => {
      if ([members.status, history.status].some((status) => status === 401 || status === 403)) { router.replace(`/${locale}/login`); return; }
      const memberData = await members.json(), historyData = await history.json();
      if (!members.ok) throw new Error(memberData.message); if (!history.ok) throw new Error(historyData.message);
      setAthletes(memberData || []); setRecords(historyData || []); setError("");
    }).catch((reason) => setError(reason instanceof Error ? reason.message : "دریافت دفتر حضور و غیاب انجام نشد."));
  }, [id, locale, month, router, year]);
  const dates = useMemo(() => [...new Set(records.map((record) => record.attendance_date.slice(0, 10)))], [records]);
  const present = useMemo(() => new Set(records.map((record) => `${record.attendance_date.slice(0, 10)}:${record.athlete_id}`)), [records]);
  const shown = useMemo(() => athletes.filter((athlete) => `${athlete.first_name} ${athlete.last_name}`.includes(query.trim())), [athletes, query]);
  const title = `${months[month - 1]} ${year.toLocaleString("fa-IR", { useGrouping: false })}`;
  return <main className="catalog-page attendance-history-page"><header className="no-print"><Link href={`/${locale}/gyms/${id}/athletes`}>بازگشت به اعضای باشگاه ←</Link></header><section className="wallet-ledger roster-card"><div className="roster-head"><div><span>سوابق ثبت‌شده</span><h1>دفتر حضور و غیاب</h1></div><button className="print-button no-print" onClick={() => window.print()}>چاپ</button></div><div className="attendance-toolbar no-print"><div className="month-switcher"><button type="button" onClick={() => moveMonth(-1)}>‹</button><b>{title}</b><button type="button" onClick={() => moveMonth(1)}>›</button></div><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="جستجو بر اساس نام عضو" aria-label="جستجو بر اساس نام عضو" /></div>{error && <div className="alert">{error}</div>}<div className="attendance-matrix"><table><thead><tr><th>نام عضو</th>{dates.map((date) => <th key={date}>{new Date(`${date}T00:00:00`).toLocaleDateString("fa-IR")}</th>)}</tr></thead><tbody>{shown.map((athlete) => <tr key={athlete.id}><th>{athlete.first_name} {athlete.last_name}</th>{dates.map((date) => <td key={date}>{present.has(`${date}:${athlete.id}`) && <span className="attendance-mark">✓</span>}</td>)}</tr>)}</tbody></table>{!dates.length && !error && <div className="roster-empty">در {title} حضور و غیابی ثبت نشده است.</div>}</div></section></main>;
}
