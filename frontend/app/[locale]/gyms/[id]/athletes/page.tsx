"use client";
import Link from "next/link";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import "../../../../catalog.css";
import "../../../../roster.css";
import "../../../../roster-overrides.css";
import "../../../../print.css";
import "../../../../attendance-report.css";
import "../../../../member-print.css";
const API = "";
type Athlete = { id: string; first_name: string; last_name: string; phone?: string; age: number; tuition_status: "settled" | "debtor"; unpaid_months: number; outstanding_amount: number; registered_at: string; last_weight_kg?: number };
type Gym = { name: string };
export default function GymAthletes({ params }: { params: Promise<{ locale: string; id: string }> }) {
  const { locale, id } = use(params), router = useRouter();
  const [items, setItems] = useState<Athlete[]>([]), [gym, setGym] = useState<Gym | null>(null), [error, setError] = useState("");
  useEffect(() => { const headers = { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` }; Promise.all([fetch(`${API}/v1/gyms/${id}/athletes`, { headers }), fetch(`${API}/v1/gyms/${id}`, { headers })]).then(async ([members, gymResponse]) => { if ([members.status, gymResponse.status].some((status) => status === 401 || status === 403)) { router.replace(`/${locale}/login`); return; } const memberData = await members.json(), gymData = await gymResponse.json(); if (!members.ok) throw new Error(memberData.message); if (!gymResponse.ok) throw new Error(gymData.message); setItems(memberData || []); setGym(gymData); }).catch((reason) => setError(reason instanceof Error ? reason.message : "دریافت ورزشکاران انجام نشد.")); }, [id, locale, router]);
  return <main className="catalog-page"><header className="no-print"><Link href={`/${locale}/gyms`}>بازگشت به باشگاه‌های من ←</Link></header><section className="wallet-ledger roster-card"><div className="roster-head roster-page-head"><div><span>اعضای باشگاه</span><h1>{gym ? `اعضای ${gym.name}` : "اعضای باشگاه"}</h1></div><div className="roster-head-actions"><strong>{items.length.toLocaleString("fa-IR")} ورزشکار</strong><Link className="no-print" href={`/${locale}/gyms/${id}/attendance/history`}>دفتر حضور/غیاب</Link></div></div>{error && <div className="alert">{error}</div>}<div className="roster-table"><div className="roster-row roster-title roster-row-members"><span>نام و نام خانوادگی</span><span>شماره موبایل</span><span>سن</span><span>وزن</span><span>وضعیت شهریه</span><span>تاریخ ثبت‌نام</span></div>{items.map((item) => <div className="roster-row roster-row-members" key={item.id}><b>{item.first_name} {item.last_name}</b><span dir="ltr">{item.phone || "—"}</span><span>{item.age ? `${item.age.toLocaleString("fa-IR")} سال` : "—"}</span><span>{item.last_weight_kg == null ? "—" : `${item.last_weight_kg.toLocaleString("fa-IR")} کیلوگرم`}</span><Link className={`tuition-cell ${item.tuition_status}`} href={`/${locale}/gyms/${id}/athletes/${item.id}/payments`}>{item.tuition_status === "debtor" ? <><b>بدهکار</b><small>{item.outstanding_amount.toLocaleString("fa-IR")} ریال · {item.unpaid_months.toLocaleString("fa-IR")} ماه</small></> : <><b>تسویه‌شده</b><small>مشاهده پرداخت‌ها</small></>}</Link><time>{new Date(item.registered_at).toLocaleDateString("fa-IR")}</time></div>)}{!items.length && !error && <div className="roster-empty">هنوز ورزشکاری در این باشگاه ثبت‌نام نکرده است.</div>}</div><div className="member-print-action no-print"><button onClick={() => window.print()}>چاپ فهرست اعضا</button></div></section></main>;
}
