"use client";
import Link from "next/link";
import { use, useCallback, useEffect, useMemo, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = "";
const money = (n: number) => `${Number(n).toLocaleString("fa-IR")} ریال`;
type Status = "under_review" | "queued" | "completed" | "rejected";
type Withdrawal = {
  id: string;
  coach_name: string;
  gross_amount: number;
  fee_amount: number;
  net_amount: number;
  iban: string;
  status: Status;
  bank_tracking_code?: string;
  rejection_reason?: string;
  created_at: string;
  reviewed_at?: string;
  scheduled_completion_at?: string;
  completed_at?: string;
};
const labels: Record<Status, string> = {
  under_review: "در حال بررسی",
  queued: "در صف پرداخت",
  completed: "تکمیل‌شده",
  rejected: "ردشده",
};
export default function WithdrawalsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [items, setItems] = useState<Withdrawal[]>([]),
    [error, setError] = useState(""),
    [working, setWorking] = useState(""),
    [copiedIBAN, setCopiedIBAN] = useState("");
  const load = useCallback(
    () =>
      fetch(`${API}/v1/withdrawals`, {
        headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
      })
        .then(async (r) => {
          const d = await r.json();
          if (!r.ok) throw new Error(d.message);
          setItems(d || []);
        })
        .catch((e) => setError(e.message)),
    []
  );
  useEffect(() => {
    load();
    const timer = setInterval(load, 30000);
    return () => clearInterval(timer);
  }, [load]);
  const review = async (item: Withdrawal, status: "queued" | "rejected") => {
    const value = prompt(
      status === "queued"
        ? "پس از انجام انتقال بانکی، کد پیگیری بانک را وارد کنید:"
        : "علت رد درخواست را وارد کنید:"
    );
    if (!value?.trim()) return;
    setWorking(item.id);
    const body =
      status === "queued"
        ? { status, bank_tracking_code: value.trim() }
        : { status, rejection_reason: value.trim() };
    const r = await fetch(`${API}/v1/withdrawals/${item.id}/review`, {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
        "Content-Type": "application/json",
        "Accept-Language": "fa",
      },
      body: JSON.stringify(body),
    });
    const d = await r.json();
    setWorking("");
    if (!r.ok) {
      setError(d.message);
      return;
    }
    load();
  };
  const pending = useMemo(() => items.filter((x) => x.status === "under_review").length, [items]);
  const copyIBAN = async (item: Withdrawal) => {
    try {
      await navigator.clipboard.writeText(item.iban);
    } catch {
      const field = document.createElement("textarea");
      field.value = item.iban;
      field.style.position = "fixed";
      field.style.opacity = "0";
      document.body.appendChild(field);
      field.select();
      document.execCommand("copy");
      field.remove();
    }
    setCopiedIBAN(item.id);
    window.setTimeout(() => setCopiedIBAN((id) => (id === item.id ? "" : id)), 1800);
  };
  return (
    <AdminShell
      locale={locale}
      active="finance"
      kicker="کارتابل تسویه مربیان"
      title="درخواست‌های برداشت"
    >
      <nav className="admin-tabs">
        <Link href={`/${locale}/admin/finance`}>تنظیمات مالی</Link>
        <Link href={`/${locale}/admin/payments`}>پرداخت شهریه</Link>
        <Link className="active" href={`/${locale}/admin/withdrawals`}>
          درخواست‌های برداشت
        </Link>
      </nav>
      <div className="admin-grid">
        <article className="admin-stat">
          <span>نیازمند بررسی</span>
          <strong>{pending.toLocaleString("fa-IR")}</strong>
          <small>درخواست جدید</small>
        </article>
        <article className="admin-stat">
          <span>در صف پرداخت</span>
          <strong>
            {items.filter((x) => x.status === "queued").length.toLocaleString("fa-IR")}
          </strong>
          <small>دارای کد پیگیری بانک</small>
        </article>
        <article className="admin-stat">
          <span>تکمیل‌شده</span>
          <strong>
            {items.filter((x) => x.status === "completed").length.toLocaleString("fa-IR")}
          </strong>
          <small>نهایی‌شده در ساعت ۲۳:۳۰</small>
        </article>
      </div>
      {error && (
        <div className="alert" style={{ marginTop: 14 }}>
          {error}
        </div>
      )}
      <section className="admin-panel" style={{ marginTop: 16 }}>
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>مربی / شبا</th>
                <th>درخواست</th>
                <th>کارمزد</th>
                <th>خالص واریزی</th>
                <th>وضعیت</th>
                <th>پیگیری / علت</th>
                <th>زمان‌بندی</th>
                <th>عملیات</th>
              </tr>
            </thead>
            <tbody>
              {items.map((x) => (
                <tr key={x.id}>
                  <td>
                    <strong>{x.coach_name}</strong>
                    <div className="iban-copy-row">
                      <small className="ltr">{x.iban}</small>
                      <button
                        type="button"
                        className="copy-iban-button"
                        onClick={() => copyIBAN(x)}
                        aria-label={`کپی شماره شبای ${x.coach_name}`}
                      >
                        <svg viewBox="0 0 24 24" aria-hidden="true">
                          <rect x="8" y="8" width="11" height="11" rx="2" />
                          <path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2" />
                        </svg>
                        <span>{copiedIBAN === x.id ? "کپی شد" : "کپی"}</span>
                      </button>
                    </div>
                  </td>
                  <td>
                    <strong className="money">{money(x.gross_amount)}</strong>
                    <small>{new Date(x.created_at).toLocaleString("fa-IR")}</small>
                  </td>
                  <td className="money">{money(x.fee_amount)}</td>
                  <td className="money">
                    <strong>{money(x.net_amount)}</strong>
                  </td>
                  <td>
                    <span
                      className={`admin-badge ${x.status === "under_review" ? "warn" : x.status === "rejected" ? "danger" : x.status === "queued" ? "info" : ""}`}
                    >
                      {labels[x.status]}
                    </span>
                  </td>
                  <td>
                    {x.bank_tracking_code ? (
                      <span className="ltr">{x.bank_tracking_code}</span>
                    ) : (
                      x.rejection_reason || "—"
                    )}
                  </td>
                  <td>
                    {x.completed_at
                      ? new Date(x.completed_at).toLocaleString("fa-IR")
                      : x.scheduled_completion_at
                        ? `تکمیل: ${new Date(x.scheduled_completion_at).toLocaleString("fa-IR")}`
                        : "—"}
                  </td>
                  <td>
                    <div className="admin-actions">
                      {x.status === "under_review" && (
                        <>
                          <button disabled={working === x.id} onClick={() => review(x, "queued")}>
                            ثبت پرداخت بانکی
                          </button>
                          <button
                            disabled={working === x.id}
                            className="danger"
                            onClick={() => review(x, "rejected")}
                          >
                            رد
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!items.length && !error && <div className="admin-empty">درخواستی ثبت نشده است.</div>}
      </section>
      <div className="admin-note">
        ثبت کد پیگیری به معنی انجام انتقال بانکی توسط مدیر است و درخواست را وارد «صف پرداخت» می‌کند.
        سامانه در ساعت ۲۳:۳۰ تهران تراکنش‌های سررسیدشده را نهایی کرده و هم‌زمان برای مربی پیام
        می‌فرستد.
      </div>
    </AdminShell>
  );
}
