"use client";

import { ReactNode, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardSidebar from "./DashboardSidebar";

const API = process.env.NEXT_PUBLIC_API_URL!;
export type AdminUser = {
  id: string;
  role: "admin";
  first_name: string;
  last_name: string;
  face_image_id?: string;
};

export default function AdminShell({
  locale,
  active,
  title,
  kicker,
  children,
}: {
  locale: string;
  active: "dashboard" | "users" | "gyms" | "finance" | "payments" | "withdrawals";
  title: string;
  kicker: string;
  children: ReactNode;
}) {
  const router = useRouter(),
    [user, setUser] = useState<AdminUser | null>(null),
    [error, setError] = useState("");
  useEffect(() => {
    const token = localStorage.getItem("gymax_access_token") || "";
    if (!token) {
      router.replace(`/${locale}/login`);
      return;
    }
    fetch(`${API}/v1/users/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async (response) => {
        const data = await response.json();
        if (!response.ok || data.role !== "admin") throw new Error("دسترسی مدیر سایت لازم است.");
        setUser(data);
      })
      .catch((reason) => {
        setError(reason instanceof Error ? reason.message : "خطا در احراز هویت");
        setTimeout(() => router.replace(`/${locale}/login`), 800);
      });
  }, [locale, router]);
  if (!user)
    return (
      <div className="screen-loader">
        <i>G</i>
        <span>{error || "در حال آماده‌سازی پنل مدیریت..."}</span>
      </div>
    );
  return (
    <div className="shell admin-shell">
      <DashboardSidebar user={user} locale={locale} active={active} />
      <main className="dash admin-page">
        <header>
          <div>
            <span>{kicker}</span>
            <h1>{title}</h1>
          </div>
        </header>
        {children}
      </main>
    </div>
  );
}
