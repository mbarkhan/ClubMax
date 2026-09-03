"use client";

import Link from "next/link";
import { FormEvent, use, useCallback, useEffect, useState } from "react";
import DashboardSidebar from "../../components/DashboardSidebar";
import "../../admin.css";
import "../../tickets.css";

const API = process.env.NEXT_PUBLIC_API_URL!;
type Role = "admin" | "coach" | "athlete";
type User = {
  id: string;
  role: Role;
  first_name: string;
  last_name: string;
  face_image_id?: string;
};
type Ticket = {
  id: string;
  number: number;
  owner_name: string;
  owner_role: Role;
  category: string;
  subject: string;
  status: string;
  unread: boolean;
  last_message_at: string;
};
const categories: Record<string, string> = {
  payment: "پرداخت و امور مالی",
  account: "حساب کاربری",
  gym: "باشگاه",
  technical: "مشکل فنی",
  other: "سایر موارد",
};
const statuses: Record<string, string> = {
  waiting_admin: "در انتظار پاسخ پشتیبانی",
  waiting_member: "در انتظار پاسخ عضو",
  closed: "بسته‌شده",
};

export default function TicketsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params);
  const [user, setUser] = useState<User | null>(null),
    [items, setItems] = useState<Ticket[]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [busy, setBusy] = useState(false);
  const [query, setQuery] = useState(""),
    [status, setStatus] = useState(""),
    [category, setCategory] = useState(""),
    [showForm, setShowForm] = useState(false);
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  const load = useCallback(async (current: User, q = "", s = "", c = "") => {
    const search = new URLSearchParams({ limit: "100" });
    if (q) search.set("q", q);
    if (s) search.set("status", s);
    if (c) search.set("category", c);
    const r = await fetch(`${API}/v1/tickets?${search}`, {
        headers: { Authorization: `Bearer ${token()}` },
      }),
      d = await r.json();
    if (!r.ok) throw new Error(d.message || "دریافت تیکت‌ها انجام نشد.");
    setItems(d || []);
  }, []);
  useEffect(() => {
    fetch(`${API}/v1/users/me`, { headers: { Authorization: `Bearer ${token()}` } })
      .then(async (r) => {
        const d = await r.json();
        if (!r.ok) throw new Error(d.message);
        setUser(d);
        await load(d);
      })
      .catch((e) => setError(e.message || "خطا در احراز هویت"));
  }, [load]);
  useEffect(() => {
    if (!user) return;
    const timer = setTimeout(
      () => load(user, query, status, category).catch((e) => setError(e.message)),
      250
    );
    return () => clearTimeout(timer);
  }, [user, query, status, category, load]);
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const form = new FormData(e.currentTarget);
    const payload = Object.fromEntries(form.entries());
    try {
      const r = await fetch(`${API}/v1/tickets`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token()}`, "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        }),
        d = await r.json();
      if (!r.ok) throw new Error(d.message);
      e.currentTarget.reset();
      setShowForm(false);
      setNotice(
        `تیکت شماره ${Number(d.number).toLocaleString("fa-IR", { useGrouping: false })} ثبت شد.`
      );
      if (user) await load(user);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ثبت تیکت انجام نشد.");
    } finally {
      setBusy(false);
    }
  }
  if (!user)
    return (
      <div className="screen-loader">
        <i>C</i>
        <span>{error || "در حال آماده‌سازی پشتیبانی..."}</span>
      </div>
    );
  const isAdmin = user.role === "admin";
  return (
    <div className="shell">
      <DashboardSidebar user={user} locale={locale} active="tickets" />
      <main className="dash tickets-page">
        <header className="tickets-header">
          <div>
            <span>{isAdmin ? "مرکز پشتیبانی اعضا" : "ارتباط مستقیم با پشتیبانی"}</span>
            <h1>{isAdmin ? "مدیریت تیکت‌ها" : "پشتیبانی و تیکت‌ها"}</h1>
            <p>
              {isAdmin
                ? "درخواست‌های مربیان و ورزشکاران را بررسی و پاسخ دهید."
                : "مشکل خود، به‌ویژه مسائل پرداخت، را ثبت و تا حل کامل پیگیری کنید."}
            </p>
          </div>
          {!isAdmin && (
            <button className="ticket-primary" onClick={() => setShowForm((v) => !v)}>
              {showForm ? "انصراف" : "ثبت تیکت جدید"}
            </button>
          )}
        </header>
        {notice && <div className="ticket-notice">{notice}</div>}
        {error && <div className="alert">{error}</div>}
        {!isAdmin && showForm && (
          <form className="ticket-compose" onSubmit={submit}>
            <div>
              <label>موضوع تیکت</label>
              <input name="subject" maxLength={200} required />
            </div>
            <div>
              <label>دسته‌بندی</label>
              <select name="category" defaultValue="payment">
                {Object.entries(categories).map(([v, l]) => (
                  <option key={v} value={v}>
                    {l}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label>شرح کامل مشکل</label>
              <textarea
                name="body"
                maxLength={5000}
                required
                rows={6}
                placeholder="جزئیات، تاریخ، مبلغ و کد پیگیری مرتبط را بنویسید..."
              />
            </div>
            <button className="ticket-primary" disabled={busy}>
              {busy ? "در حال ثبت..." : "ارسال برای پشتیبانی"}
            </button>
          </form>
        )}
        <section className="ticket-toolbar">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={isAdmin ? "جست‌وجوی شماره، موضوع یا نام عضو..." : "جست‌وجوی تیکت..."}
          />
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="">همه وضعیت‌ها</option>
            {Object.entries(statuses).map(([v, l]) => (
              <option key={v} value={v}>
                {l}
              </option>
            ))}
          </select>
          <select value={category} onChange={(e) => setCategory(e.target.value)}>
            <option value="">همه دسته‌ها</option>
            {Object.entries(categories).map(([v, l]) => (
              <option key={v} value={v}>
                {l}
              </option>
            ))}
          </select>
        </section>
        <section className="ticket-list">
          {items.map((t) => (
            <Link
              href={`/${locale}/tickets/${t.id}`}
              className={`ticket-card ${t.unread ? "unread" : ""}`}
              key={t.id}
            >
              <div className="ticket-card-main">
                <div className="ticket-card-top">
                  <span className="ticket-number">
                    تیکت #{t.number.toLocaleString("fa-IR", { useGrouping: false })}
                  </span>
                  {t.unread && <b className="ticket-new">جدید</b>}
                  <span className={`ticket-status ${t.status}`}>{statuses[t.status]}</span>
                </div>
                <h2>{t.subject}</h2>
                <div className="ticket-meta">
                  {isAdmin && (
                    <span>
                      عضو: {t.owner_name} ({t.owner_role === "coach" ? "مربی" : "ورزشکار"})
                    </span>
                  )}
                  <span>{categories[t.category]}</span>
                  <time>{new Date(t.last_message_at).toLocaleString("fa-IR")}</time>
                </div>
              </div>
              <span className="ticket-arrow">←</span>
            </Link>
          ))}
          {!items.length && !error && (
            <div className="ticket-empty">
              <span>✓</span>
              <h2>{isAdmin ? "صف پشتیبانی خالی است" : "هنوز تیکتی ثبت نکرده‌اید"}</h2>
              <p>
                {isAdmin
                  ? "درخواست جدیدی مطابق فیلتر فعلی وجود ندارد."
                  : "اگر مسئله‌ای داشتید، پشتیبانی کلاب‌ماکس همراه شماست."}
              </p>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
