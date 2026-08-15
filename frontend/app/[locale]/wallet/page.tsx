"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";

const API = process.env.NEXT_PUBLIC_API_URL!;
type Entry = {
  id: string;
  type: string;
  amount: number;
  balance_after: number;
  description: string;
  reference_type: string;
  created_at: string;
};
const money = (value: number) => `${Math.abs(Number(value)).toLocaleString("fa-IR")} ریال`;

export default function Wallet() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter(),
    [balance, setBalance] = useState(0),
    [entries, setEntries] = useState<Entry[]>([]),
    [error, setError] = useState("");
  useEffect(() => {
    const token = localStorage.getItem("gymax_access_token") || "",
      headers = { Authorization: `Bearer ${token}`, "Accept-Language": "fa" };
    Promise.all([
      fetch(`${API}/v1/wallet`, { headers }),
      fetch(`${API}/v1/wallet/ledger`, { headers }),
    ])
      .then(async ([wallet, ledger]) => {
        if (wallet.status === 401 || wallet.status === 403) {
          router.replace(`/${locale}/login`);
          return;
        }
        const walletData = await wallet.json(),
          ledgerData = await ledger.json();
        if (!wallet.ok) throw new Error(walletData.message);
        if (!ledger.ok) throw new Error(ledgerData.message);
        setBalance(Number(walletData.balance || 0));
        setEntries(ledgerData || []);
      })
      .catch((reason) =>
        setError(reason instanceof Error ? reason.message : "دریافت کیف پول انجام نشد.")
      );
  }, [locale, router]);
  return (
    <main className="catalog-page">
      <header>
        <Link href={`/${locale}/dashboard`}>بازگشت به داشبورد ←</Link>
      </header>
      <section className="wallet-head">
        <span>موجودی قابل برداشت</span>
        <strong>{balance.toLocaleString("fa-IR")} ریال</strong>
      </section>
      <section className="wallet-ledger wallet-ledger-card">
        <div className="wallet-title">
          <div>
            <span>دفتر مالی مربی</span>
            <h1>تراکنش‌های کیف پول</h1>
          </div>
        </div>
        {error && <div className="alert">{error}</div>}
        <div className="wallet-table">
          <div className="wallet-row wallet-row-title">
            <span>نوع</span>
            <span>شرح</span>
            <span>تاریخ</span>
            <span>مبلغ</span>
            <span>مانده</span>
          </div>
          {entries.map((entry) => (
            <div className="wallet-row" key={entry.id}>
              <b className={entry.amount >= 0 ? "deposit" : "withdrawal"}>
                {entry.amount >= 0 ? "واریز" : "برداشت"}
              </b>
              <span>{entry.description}</span>
              <time>{new Date(entry.created_at).toLocaleString("fa-IR")}</time>
              <strong className={entry.amount >= 0 ? "credit" : "debit"}>
                {entry.amount >= 0 ? "+" : "−"}
                {money(entry.amount)}
              </strong>
              <strong>{Number(entry.balance_after).toLocaleString("fa-IR")} ریال</strong>
            </div>
          ))}
          {!entries.length && !error && <p className="wallet-empty">هنوز تراکنشی ثبت نشده است.</p>}
        </div>
      </section>
    </main>
  );
}
