"use client";
import Link from "next/link";
import { use, useEffect, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = "";
type Policy = { min_amount: number; max_amount: number; fee_bps: number; updated_at: string };
type Overview = {
  gross_payments: number;
  platform_revenue: number;
  withdrawal_fees: number;
  pending_withdrawals: number;
};
const money = (n: number) => `${Number(n).toLocaleString("fa-IR")} ریال`;
export default function FinancePage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [policy, setPolicy] = useState<Policy | null>(null),
    [overview, setOverview] = useState<Overview | null>(null),
    [message, setMessage] = useState(""),
    [error, setError] = useState("");
  useEffect(() => {
    const headers = {
      Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
    };
    Promise.all([
      fetch(`${API}/v1/finance/policy`, { headers }),
      fetch(`${API}/v1/admin/overview`, { headers }),
    ])
      .then(async ([p, o]) => {
        const pd = await p.json(),
          od = await o.json();
        if (!p.ok) throw new Error(pd.message);
        if (!o.ok) throw new Error(od.message);
        setPolicy(pd);
        setOverview(od);
      })
      .catch((e) => setError(e.message));
  }, []);
  const save = async () => {
    if (!policy) return;
    setMessage("");
    const r = await fetch(`${API}/v1/finance/policy`, {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(policy),
    });
    const d = await r.json();
    if (!r.ok) {
      setError(d.message);
      return;
    }
    setMessage("تنظیمات مالی ذخیره شد.");
  };
  return (
    <AdminShell locale={locale} active="finance" kicker="کنترل درآمد و تسویه" title="مدیریت مالی">
      <nav className="admin-tabs">
        <Link className="active" href={`/${locale}/admin/finance`}>
          تنظیمات مالی
        </Link>
        <Link href={`/${locale}/admin/payments`}>پرداخت شهریه</Link>
        <Link href={`/${locale}/admin/withdrawals`}>درخواست‌های برداشت</Link>
      </nav>
      <div className="admin-grid">
        <article className="admin-stat">
          <span>گردش پرداخت‌ها</span>
          <strong>{money(overview?.gross_payments || 0)}</strong>
          <small>مجموع پرداخت ورزشکاران</small>
        </article>
        <article className="admin-stat">
          <span>درآمد پرداخت</span>
          <strong>{money(overview?.platform_revenue || 0)}</strong>
          <small>کارمزدهای ثبت‌شده</small>
        </article>
        <article className="admin-stat">
          <span>درآمد برداشت</span>
          <strong>{money(overview?.withdrawal_fees || 0)}</strong>
          <small>کارمزد برداشت تکمیل‌شده</small>
        </article>
        <article className="admin-stat">
          <span>کارتابل باز</span>
          <strong>{(overview?.pending_withdrawals || 0).toLocaleString("fa-IR")}</strong>
          <small>در حال بررسی یا صف پرداخت</small>
        </article>
      </div>
      <section className="admin-panel" style={{ marginTop: 16 }}>
        <div className="admin-section-title">
          <h2>برداشت مربیان</h2>
        </div>
        {message && <div className="success-message">{message}</div>}
        {error && <div className="alert">{error}</div>}
        {policy && (
          <div className="admin-form">
            <label>
              <span>حداقل مبلغ درخواست (ریال)</span>
              <input
                type="number"
                min="1"
                value={policy.min_amount}
                onChange={(e) => setPolicy({ ...policy, min_amount: Number(e.target.value) })}
              />
            </label>
            <label>
              <span>حداکثر مبلغ درخواست (ریال)</span>
              <input
                type="number"
                min="1"
                value={policy.max_amount}
                onChange={(e) => setPolicy({ ...policy, max_amount: Number(e.target.value) })}
              />
            </label>
            <label>
              <span>درصد کارمزد سایت</span>
              <input
                type="number"
                min="0"
                max="100"
                step="0.01"
                value={policy.fee_bps / 100}
                onChange={(e) =>
                  setPolicy({ ...policy, fee_bps: Math.round(Number(e.target.value) * 100) })
                }
              />
            </label>
            <div className="form-actions">
              <button className="primary-button" onClick={save}>
                ذخیره تنظیمات
              </button>
            </div>
          </div>
        )}
        <div className="admin-note">
          حداکثر مجاز هر مربی، کمترین مقدار میان سقف تعیین‌شده در این صفحه و موجودی قابل برداشت کیف
          پول اوست. کارمزد هنگام ثبت درخواست محاسبه و مبلغ ناخالص همان لحظه به‌صورت تراکنشی رزرو
          می‌شود.
        </div>
      </section>
    </AdminShell>
  );
}
