"use client";
import Link from "next/link";
import { use, useCallback, useEffect, useMemo, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = "";
const money = (n: number) => `${Number(n).toLocaleString("fa-IR")} ریال`;
type Payment = {
  id: string;
  athlete_name: string;
  athlete_phone?: string;
  gym_name: string;
  coach_name: string;
  tuition_amount: number;
  platform_fee: number;
  total_amount: number;
  tracking_id: string;
  payment_kind: "registration" | "monthly";
  billing_year?: number;
  billing_month?: number;
  created_at: string;
};
export default function AdminPayments({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [items, setItems] = useState<Payment[]>([]),
    [query, setQuery] = useState(""),
    [page, setPage] = useState(0),
    [error, setError] = useState("");
  const load = useCallback(
    () =>
      fetch(
        `${API}/v1/admin/payments?limit=100&offset=${page * 100}&q=${encodeURIComponent(query)}`,
        {
          headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
        }
      )
        .then(async (r) => {
          const d = await r.json();
          if (!r.ok) throw new Error(d.message);
          setItems(d || []);
        })
        .catch((e) => setError(e.message)),
    [query, page]
  );
  useEffect(() => {
    const t = setTimeout(load, 250);
    return () => clearTimeout(t);
  }, [load]);
  const total = useMemo(() => items.reduce((s, x) => s + Number(x.total_amount), 0), [items]);
  return (
    <AdminShell locale={locale} active="finance" kicker="پیگیری تراکنش" title="پرداخت‌های شهریه">
      <nav className="admin-tabs">
        <Link href={`/${locale}/admin/finance`}>تنظیمات مالی</Link>
        <Link className="active" href={`/${locale}/admin/payments`}>
          پرداخت شهریه
        </Link>
        <Link href={`/${locale}/admin/withdrawals`}>درخواست‌های برداشت</Link>
      </nav>
      <div className="admin-grid">
        <article className="admin-stat">
          <span>تعداد نتایج</span>
          <strong>{items.length.toLocaleString("fa-IR")}</strong>
          <small>تراکنش موفق</small>
        </article>
        <article className="admin-stat">
          <span>مجموع پرداختی</span>
          <strong>{money(total)}</strong>
          <small>شهریه و کارمزد</small>
        </article>
      </div>
      <div className="admin-toolbar" style={{ marginTop: 16 }}>
        <input
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setPage(0);
          }}
          placeholder="جستجو با کد پیگیری، ورزشکار یا باشگاه..."
        />
      </div>
      {error && <div className="alert">{error}</div>}
      <section className="admin-panel">
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>پرداخت‌کننده</th>
                <th>باشگاه / مربی</th>
                <th>نوع</th>
                <th>مبلغ شهریه</th>
                <th>کارمزد</th>
                <th>مبلغ کل</th>
                <th>کد پیگیری</th>
                <th>زمان</th>
              </tr>
            </thead>
            <tbody>
              {items.map((p) => (
                <tr key={p.id}>
                  <td>
                    <strong>{p.athlete_name}</strong>
                    <small dir="ltr">{p.athlete_phone || "—"}</small>
                  </td>
                  <td>
                    <strong>{p.gym_name}</strong>
                    <small>مربی: {p.coach_name}</small>
                  </td>
                  <td>
                    {p.payment_kind === "monthly" ? "شهریه ماهانه" : "ثبت‌نام"}
                    {(p.billing_year || 0) > 0 && (
                      <small>
                        {p.billing_year}/{p.billing_month}
                      </small>
                    )}
                  </td>
                  <td className="money">{money(p.tuition_amount)}</td>
                  <td className="money">{money(p.platform_fee)}</td>
                  <td className="money">
                    <strong>{money(p.total_amount)}</strong>
                  </td>
                  <td className="ltr">{p.tracking_id}</td>
                  <td>{new Date(p.created_at).toLocaleString("fa-IR")}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!items.length && !error && <div className="admin-empty">پرداختی یافت نشد.</div>}
        <div className="admin-pagination">
          <button disabled={page === 0} onClick={() => setPage((value) => value - 1)}>
            صفحه قبل
          </button>
          <span>صفحه {(page + 1).toLocaleString("fa-IR")}</span>
          <button disabled={items.length < 100} onClick={() => setPage((value) => value + 1)}>
            صفحه بعد
          </button>
        </div>
      </section>
    </AdminShell>
  );
}
