"use client";
import Link from "next/link";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import "../../../../catalog.css";
import "../../../../athlete-attendance.css";

const API = "";
type Record = { attendance_date: string; present: boolean };
export default function MyAttendance({ params }: { params: Promise<{ locale: string; id: string }> }) {
  const { locale, id } = use(params), router = useRouter();
  const [records, setRecords] = useState<Record[]>([]), [error, setError] = useState("");
  useEffect(() => { fetch(`${API}/v1/gyms/${id}/my-attendance`, { headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` } }).then(async (response) => { if (response.status === 401 || response.status === 403) { router.replace(`/${locale}/login`); return; } const data = await response.json(); if (!response.ok) throw new Error(data.message); setRecords(data || []); }).catch((reason) => setError(reason instanceof Error ? reason.message : "دریافت سوابق حضور و غیاب انجام نشد.")); }, [id, locale, router]);
  return <main className="catalog-page athlete-attendance"><header><Link href={`/${locale}/gyms`}>بازگشت به باشگاه‌های من ←</Link></header><section><span>سوابق من</span><h1>حضور و غیاب من</h1>{error && <div className="alert">{error}</div>}<div>{records.map((record) => <article className={record.present ? "present" : "absent"} key={record.attendance_date}><time>{new Date(`${record.attendance_date.slice(0, 10)}T00:00:00`).toLocaleDateString("fa-IR", { dateStyle: "full" })}</time><b>{record.present ? "حاضر" : "غایب"}</b></article>)}{!records.length && !error && <p>هنوز حضور و غیابی برای این باشگاه ثبت نشده است.</p>}</div></section></main>;
}
