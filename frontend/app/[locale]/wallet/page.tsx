"use client";
import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import DashboardSidebar from "../../components/DashboardSidebar";
import "../../admin.css";
const API = process.env.NEXT_PUBLIC_API_URL!;
const money = (v: number) => `${Math.abs(Number(v)).toLocaleString("fa-IR")} ریال`;
type User = {
  id: string;
  role: "coach";
  first_name: string;
  last_name: string;
  face_image_id?: string;
  bank_iban?: string;
};
type Entry = {
  id: string;
  type: string;
  amount: number;
  balance_after: number;
  description: string;
  created_at: string;
};
type Policy = { min_amount: number; max_amount: number; fee_bps: number };
type Status = "under_review" | "queued" | "completed" | "rejected";
type Withdrawal = {
  id: string;
  gross_amount: number;
  fee_amount: number;
  net_amount: number;
  iban: string;
  status: Status;
  bank_tracking_code?: string;
  rejection_reason?: string;
  created_at: string;
  scheduled_completion_at?: string;
  completed_at?: string;
};
const labels: Record<Status, string> = {
  under_review: "در حال بررسی",
  queued: "در صف پرداخت",
  completed: "تکمیل‌شده",
  rejected: "ردشده",
};
export default function Wallet() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter(),
    [user, setUser] = useState<User | null>(null),
    [balance, setBalance] = useState(0),
    [entries, setEntries] = useState<Entry[]>([]),
    [policy, setPolicy] = useState<Policy | null>(null),
    [requests, setRequests] = useState<Withdrawal[]>([]),
    [amount, setAmount] = useState(""),
    [error, setError] = useState(""),
    [message, setMessage] = useState(""),
    [working, setWorking] = useState(false);
  const headers = {
    Authorization: `Bearer ${typeof window !== "undefined" ? localStorage.getItem("gymax_access_token") || "" : ""}`,
    "Accept-Language": "fa",
  };
  const load = () =>
    Promise.all([
      fetch(`${API}/v1/users/me`, { headers }),
      fetch(`${API}/v1/wallet`, { headers }),
      fetch(`${API}/v1/wallet/ledger`, { headers }),
      fetch(`${API}/v1/finance/policy`, { headers }),
      fetch(`${API}/v1/withdrawals`, { headers }),
    ])
      .then(async (responses) => {
        const data = await Promise.all(responses.map((r) => r.json()));
        if (responses.some((r) => r.status === 401 || r.status === 403)) {
          router.replace(`/${locale}/login`);
          return;
        }
        const bad = responses.findIndex((r) => !r.ok);
        if (bad >= 0) throw new Error(data[bad].message);
        if (data[0].role !== "coach") {
          router.replace(`/${locale}/dashboard`);
          return;
        }
        setUser(data[0]);
        setBalance(Number(data[1].balance || 0));
        setEntries(data[2] || []);
        setPolicy(data[3]);
        setRequests(data[4] || []);
      })
      .catch((e) => setError(e instanceof Error ? e.message : "دریافت اطلاعات مالی انجام نشد."));
  useEffect(() => {
    load();
    const timer = setInterval(load, 30000);
    return () => clearInterval(timer);
  }, []);
  const max = useMemo(() => Math.min(balance, policy?.max_amount ?? balance), [balance, policy]);
  const fee = useMemo(
    () => Math.floor(((Number(amount) || 0) * (policy?.fee_bps || 0)) / 10000),
    [amount, policy]
  );
  const request = async () => {
    setError("");
    setMessage("");
    setWorking(true);
    const r = await fetch(`${API}/v1/withdrawals`, {
      method: "POST",
      headers: { ...headers, "Content-Type": "application/json" },
      body: JSON.stringify({ amount: Number(amount) }),
    });
    const d = await r.json();
    setWorking(false);
    if (!r.ok) {
      setError(d.message);
      return;
    }
    setAmount("");
    setMessage("درخواست برداشت ثبت شد و اکنون در حال بررسی است.");
    load();
  };
  if (!user)
    return (
      <div className="screen-loader">
        <i>G</i>
        <span>{error || "در حال دریافت کیف پول..."}</span>
      </div>
    );
  return (
    <div className="shell">
      <DashboardSidebar user={user} locale={locale} active="wallet" />
      <main className="dash admin-page">
        <header>
          <div>
            <span>امور مالی مربی</span>
            <h1>کیف پول و برداشت</h1>
          </div>
        </header>
        <div className="admin-grid">
          <article className="admin-stat">
            <span>موجودی قابل برداشت</span>
            <strong>{money(balance)}</strong>
            <small>پس از رزرو درخواست‌های باز</small>
          </article>
          <article className="admin-stat">
            <span>سقف فعلی درخواست</span>
            <strong>{money(max)}</strong>
            <small>کمترینِ موجودی و سقف سامانه</small>
          </article>
        </div>
        <section className="admin-panel" style={{ marginTop: 16 }}>
          <div className="admin-section-title">
            <h2>درخواست برداشت جدید</h2>
          </div>
          {message && <div className="success-message">{message}</div>}
          {error && <div className="alert">{error}</div>}{" "}
          {!user.bank_iban ? (
            <div className="admin-note">
              برای برداشت ابتدا شماره شبای معتبر خود را در صفحه پروفایل ثبت کنید.
            </div>
          ) : (
            <div className="admin-form">
              <label>
                <span>مبلغ برداشت (ریال)</span>
                <input
                  type="number"
                  min={policy?.min_amount || 1}
                  max={max}
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  placeholder="مبلغ را وارد کنید"
                />
              </label>
              <label>
                <span>شماره شبای مقصد</span>
                <input className="ltr" value={user.bank_iban} disabled />
              </label>
              <label>
                <span>خالص واریزی پس از کارمزد</span>
                <input value={money(Math.max(0, (Number(amount) || 0) - fee))} disabled />
              </label>
              <div className="form-actions">
                <button
                  className="primary-button"
                  disabled={
                    working ||
                    !amount ||
                    Number(amount) < (policy?.min_amount || 0) ||
                    Number(amount) > max
                  }
                  onClick={request}
                >
                  {working ? "در حال ثبت..." : "ثبت درخواست برداشت"}
                </button>
              </div>
            </div>
          )}
        </section>
        <section className="admin-panel" style={{ marginTop: 16 }}>
          <div className="admin-section-title">
            <h2>سوابق درخواست برداشت</h2>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>زمان درخواست</th>
                  <th>مبلغ</th>
                  <th>کارمزد</th>
                  <th>خالص واریزی</th>
                  <th>وضعیت</th>
                  <th>پیگیری</th>
                </tr>
              </thead>
              <tbody>
                {requests.map((x) => (
                  <tr key={x.id}>
                    <td>{new Date(x.created_at).toLocaleString("fa-IR")}</td>
                    <td>{money(x.gross_amount)}</td>
                    <td>{money(x.fee_amount)}</td>
                    <td>
                      <strong>{money(x.net_amount)}</strong>
                    </td>
                    <td>
                      <span
                        className={`admin-badge ${x.status === "under_review" ? "warn" : x.status === "queued" ? "info" : x.status === "rejected" ? "danger" : ""}`}
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
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!requests.length && (
            <div className="admin-empty">هنوز درخواست برداشتی ثبت نشده است.</div>
          )}
        </section>
        <section className="admin-panel" style={{ marginTop: 16 }}>
          <div className="admin-section-title">
            <h2> تراکنش‌های کیف پول</h2>
          </div>
          <div className="admin-table-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>شرح</th>
                  <th>تاریخ</th>
                  <th>مبلغ</th>
                  <th>مانده</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => (
                  <tr key={e.id}>
                    <td>{e.description}</td>
                    <td>{new Date(e.created_at).toLocaleString("fa-IR")}</td>
                    <td className={e.amount >= 0 ? "credit" : "debit"}>
                      {e.amount >= 0 ? "+" : "−"}
                      {money(e.amount)}
                    </td>
                    <td>{money(e.balance_after)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!entries.length && <div className="admin-empty">هنوز تراکنشی ثبت نشده است.</div>}
        </section>
      </main>
    </div>
  );
}
