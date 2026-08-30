"use client";
import { use, useCallback, useEffect, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = process.env.NEXT_PUBLIC_API_URL!;
const money = (n: number) => `${Number(n).toLocaleString("fa-IR")} ریال`;
type Gym = {
  id: string;
  name: string;
  province: string;
  city: string;
  sport: string;
  sport_style: string;
  coach_name: string;
  coach_phone?: string;
  active_athletes: number;
  payment_count: number;
  paid_tuition: number;
  tuition_amount: number;
  capacity: number;
  last_payment_at?: string;
};
export default function GymsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [gyms, setGyms] = useState<Gym[]>([]),
    [query, setQuery] = useState(""),
    [page, setPage] = useState(0),
    [error, setError] = useState("");
  const load = useCallback(
    () =>
      fetch(`${API}/v1/admin/gyms?limit=100&offset=${page * 100}&q=${encodeURIComponent(query)}`, {
        headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
      })
        .then(async (r) => {
          const d = await r.json();
          if (!r.ok) throw new Error(d.message);
          setGyms(d || []);
        })
        .catch((e) => setError(e.message)),
    [query, page]
  );
  useEffect(() => {
    const t = setTimeout(load, 250);
    return () => clearTimeout(t);
  }, [load]);
  const remove = async (g: Gym) => {
    if (!confirm(`باشگاه «${g.name}» و تمام عضویت‌ها و سوابق وابسته حذف شود؟`)) return;
    const r = await fetch(`${API}/v1/gyms/${g.id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
    });
    if (!r.ok) {
      setError((await r.json()).message);
      return;
    }
    load();
  };
  return (
    <AdminShell locale={locale} active="gyms" kicker="نظارت باشگاه‌ها" title="همه باشگاه‌ها">
      <div className="admin-toolbar">
        <input
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setPage(0);
          }}
          placeholder="جستجو با نام باشگاه، مربی یا رشته..."
        />
      </div>
      {error && <div className="alert">{error}</div>}
      <section className="admin-panel">
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>باشگاه</th>
                <th>مربی</th>
                <th>اعضای فعال</th>
                <th>شهریه</th>
                <th>وضعیت مالی</th>
                <th>آخرین پرداخت</th>
                <th>عملیات</th>
              </tr>
            </thead>
            <tbody>
              {gyms.map((g) => (
                <tr key={g.id}>
                  <td>
                    <strong>{g.name}</strong>
                    <small>
                      {g.sport} · {g.sport_style}
                      <br />
                      {g.province}، {g.city}
                    </small>
                  </td>
                  <td>
                    <strong>{g.coach_name}</strong>
                    <small dir="ltr">{g.coach_phone || "—"}</small>
                  </td>
                  <td>
                    {g.active_athletes.toLocaleString("fa-IR")}
                    {g.capacity > 0 && <small>از ظرفیت {g.capacity.toLocaleString("fa-IR")}</small>}
                  </td>
                  <td className="money">{money(g.tuition_amount)}</td>
                  <td>
                    <strong>{money(g.paid_tuition)}</strong>
                    <small>{g.payment_count.toLocaleString("fa-IR")} تراکنش</small>
                  </td>
                  <td>
                    {g.last_payment_at ? new Date(g.last_payment_at).toLocaleString("fa-IR") : "—"}
                  </td>
                  <td>
                    <div className="admin-actions">
                      <button className="danger" onClick={() => remove(g)}>
                        حذف باشگاه
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!gyms.length && !error && <div className="admin-empty">باشگاهی ثبت نشده است.</div>}
        <div className="admin-pagination">
          <button disabled={page === 0} onClick={() => setPage((value) => value - 1)}>
            صفحه قبل
          </button>
          <span>صفحه {(page + 1).toLocaleString("fa-IR")}</span>
          <button disabled={gyms.length < 100} onClick={() => setPage((value) => value + 1)}>
            صفحه بعد
          </button>
        </div>
      </section>
    </AdminShell>
  );
}
