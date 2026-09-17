"use client";

import Link from "next/link";
import { FormEvent, use, useCallback, useEffect, useState } from "react";
import DashboardSidebar from "../../../components/DashboardSidebar";
import "../../../admin.css";
import "../../../tickets.css";
const API = "";
type Role = "admin" | "coach" | "athlete";
type User = {
  id: string;
  role: Role;
  first_name: string;
  last_name: string;
  face_image_id?: string;
};
type Detail = {
  ticket: {
    id: string;
    number: number;
    owner_name: string;
    owner_role: Role;
    owner_phone?: string;
    category: string;
    subject: string;
    status: string;
    created_at: string;
  };
  messages: {
    id: string;
    sender_id: string;
    sender_name: string;
    sender_role: Role;
    body: string;
    created_at: string;
  }[];
};
const statusNames: Record<string, string> = {
  waiting_admin: "در انتظار پاسخ پشتیبانی",
  waiting_member: "در انتظار پاسخ عضو",
  closed: "بسته‌شده",
};
const categoryNames: Record<string, string> = {
  payment: "پرداخت و امور مالی",
  account: "حساب کاربری",
  gym: "باشگاه",
  technical: "مشکل فنی",
  other: "سایر موارد",
};
export default function TicketDetail({
  params,
}: {
  params: Promise<{ locale: string; id: string }>;
}) {
  const { locale, id } = use(params),
    [user, setUser] = useState<User | null>(null),
    [detail, setDetail] = useState<Detail | null>(null),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  const load = useCallback(async () => {
    const [ur, tr] = await Promise.all([
        fetch(`${API}/v1/users/me`, { headers: { Authorization: `Bearer ${token()}` } }),
        fetch(`${API}/v1/tickets/${id}`, { headers: { Authorization: `Bearer ${token()}` } }),
      ]),
      u = await ur.json(),
      t = await tr.json();
    if (!ur.ok) throw new Error(u.message);
    if (!tr.ok) throw new Error(t.message);
    setUser(u);
    setDetail(t);
    window.dispatchEvent(new Event("clubmax:tickets-read"));
  }, [id]);
  useEffect(() => {
    load().catch((e) => setError(e.message));
  }, [load]);
  async function reply(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const form = e.currentTarget,
      body = new FormData(form).get("body");
    try {
      const r = await fetch(`${API}/v1/tickets/${id}/messages`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token()}`, "Content-Type": "application/json" },
          body: JSON.stringify({ body }),
        }),
        d = await r.json();
      if (!r.ok) throw new Error(d.message);
      form.reset();
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ارسال پاسخ انجام نشد.");
    } finally {
      setBusy(false);
    }
  }
  async function status(value: "closed" | "waiting_admin") {
    setBusy(true);
    try {
      const r = await fetch(`${API}/v1/tickets/${id}/status`, {
          method: "PUT",
          headers: { Authorization: `Bearer ${token()}`, "Content-Type": "application/json" },
          body: JSON.stringify({ status: value }),
        }),
        d = await r.json();
      if (!r.ok) throw new Error(d.message);
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "تغییر وضعیت انجام نشد.");
    } finally {
      setBusy(false);
    }
  }
  if (!user || !detail)
    return (
      <div className="screen-loader">
        <i>C</i>
        <span>{error || "در حال دریافت گفت‌وگو..."}</span>
      </div>
    );
  const t = detail.ticket,
    isAdmin = user.role === "admin";
  return (
    <div className="shell">
      <DashboardSidebar user={user} locale={locale} active="tickets" />
      <main className="dash ticket-thread-page">
        <Link className="ticket-back" href={`/${locale}/tickets`}>
          → بازگشت به فهرست تیکت‌ها
        </Link>
        <header className="thread-header">
          <div>
            <span>تیکت #{t.number.toLocaleString("fa-IR", { useGrouping: false })}</span>
            <h1>{t.subject}</h1>
            <div className="thread-labels">
              <b className={`ticket-status ${t.status}`}>{statusNames[t.status]}</b>
              <b>{categoryNames[t.category]}</b>
            </div>
          </div>
          <button
            className="ticket-secondary"
            disabled={busy}
            onClick={() => status(t.status === "closed" ? "waiting_admin" : "closed")}
          >
            {t.status === "closed" ? "بازگشایی تیکت" : "بستن تیکت"}
          </button>
        </header>
        {isAdmin && (
          <section className="ticket-owner">
            <strong>{t.owner_name}</strong>
            <span>{t.owner_role === "coach" ? "مربی" : "ورزشکار"}</span>
            {t.owner_phone && <span dir="ltr">{t.owner_phone}</span>}
          </section>
        )}
        {error && <div className="alert">{error}</div>}
        <section className="thread-messages">
          {detail.messages.map((m) => (
            <article
              className={m.sender_role === "admin" ? "support-message" : "member-message"}
              key={m.id}
            >
              <header>
                <div className="thread-avatar">
                  {m.sender_role === "admin" ? "پ" : m.sender_name.charAt(0)}
                </div>
                <div>
                  <strong>
                    {m.sender_role === "admin" ? "پشتیبانی کلاب‌ماکس" : m.sender_name}
                  </strong>
                  <time>{new Date(m.created_at).toLocaleString("fa-IR")}</time>
                </div>
              </header>
              <p>{m.body}</p>
            </article>
          ))}
        </section>
        {t.status !== "closed" ? (
          <form className="thread-reply" onSubmit={reply}>
            <label>{isAdmin ? "پاسخ پشتیبانی" : "پاسخ شما"}</label>
            <textarea
              name="body"
              required
              maxLength={5000}
              rows={5}
              placeholder="پاسخ خود را واضح و کامل بنویسید..."
            />
            <div>
              <small>پس از ارسال، طرف مقابل از پاسخ جدید مطلع می‌شود.</small>
              <button className="ticket-primary" disabled={busy}>
                {busy ? "در حال ارسال..." : "ارسال پاسخ"}
              </button>
            </div>
          </form>
        ) : (
          <div className="ticket-closed-note">
            این تیکت بسته شده است. برای ادامه گفت‌وگو ابتدا آن را بازگشایی کنید.
          </div>
        )}
      </main>
    </div>
  );
}
