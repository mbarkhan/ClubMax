"use client";

import Link from "next/link";
import { use, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import "../../../../catalog.css";

const API = "";
type Gym = {
  id: string;
  name: string;
  coach_name: string;
  sport: string;
  sport_style: string;
  province: string;
  city: string;
  tuition_amount: number;
  capacity: number;
  capacity_remaining: number | null;
  joined: boolean;
  registration_paid: boolean;
};
export default function Checkout({ params }: { params: Promise<{ locale: string; id: string }> }) {
  const { locale, id } = use(params),
    router = useRouter();
  const [gym, setGym] = useState<Gym | null>(null),
    [fee, setFee] = useState(0),
    [error, setError] = useState(""),
    [paying, setPaying] = useState(false);
  const token = () => localStorage.getItem("clubmax_access_token") || "";
  useEffect(() => {
    Promise.all([
      fetch(`${API}/v1/gyms/catalog/${id}`, { headers: { Authorization: `Bearer ${token()}` } }),
      fetch(`${API}/v1/registration/fee`, { headers: { Authorization: `Bearer ${token()}` } }),
    ])
      .then(async ([g, p]) => {
        if (g.status === 401 || g.status === 403) {
          router.replace(`/${locale}/login`);
          return;
        }
        const gymData = await g.json(),
          policy = await p.json();
        if (!g.ok) throw new Error(gymData.message);
        if (!gymData.joined || gymData.registration_paid) {
          router.replace(`/${locale}/explore-gyms/${id}`);
          return;
        }
        setGym(gymData);
        setFee(Number(policy.fee_amount || 0));
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت فاکتور انجام نشد.")
      );
  }, [id, locale, router]);
  async function pay() {
    setPaying(true);
    setError("");
    try {
      const response = await fetch(`${API}/v1/gyms/${id}/pay`, {
          method: "POST",
          headers: { Authorization: `Bearer ${token()}`, "Content-Type": "application/json" },
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "ایجاد پرداخت انجام نشد.");
      window.location.assign(data.payment_url);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ایجاد پرداخت انجام نشد.");
      setPaying(false);
    }
  }
  if (!gym)
    return (
      <main className="catalog-page">
        <div className="screen-loader">
          <span>{error || "در حال آماده‌سازی فاکتور..."}</span>
        </div>
      </main>
    );
  return (
    <main className="catalog-page">
      <header>
        <Link href={`/${locale}/explore-gyms/${id}`}>بازگشت به مشخصات باشگاه ←</Link>
      </header>
      <section className="checkout-layout">
        <div className="checkout-gym">
          <span>پرداخت امن با زیبال</span>
          <h1>{gym.name}</h1>
          <p>
            {gym.sport} · {gym.sport_style}
          </p>
          <p>مربی: {gym.coach_name}</p>
          <small>
            {gym.province}، {gym.city}
          </small>
          <div className="test-gateway-note">
            پس از انتخاب پرداخت، برای تکمیل تراکنش به درگاه امن زیبال منتقل می‌شوید.
          </div>
        </div>
        <aside>
          <h2>جزئیات فاکتور</h2>
          <div>
            <span>شهریه باشگاه</span>
            <strong>{gym.tuition_amount.toLocaleString("fa-IR")} ریال</strong>
          </div>
          <div>
            <span>کارمزد ثبت‌نام</span>
            <strong>{fee.toLocaleString("fa-IR")} ریال</strong>
          </div>
          <div className="invoice-total">
            <span>جمع پرداخت</span>
            <strong>{(gym.tuition_amount + fee).toLocaleString("fa-IR")} ریال</strong>
          </div>
          {error && <div className="alert">{error}</div>}
          <button onClick={pay} disabled={paying}>
            {paying ? "در حال انتقال به درگاه..." : "پرداخت آنلاین"}
          </button>
        </aside>
      </section>
    </main>
  );
}
