"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

const API = "";

type SignedInUser = {
  first_name: string;
  last_name: string;
  role: "admin" | "coach" | "athlete";
};

export default function HomeAccountActions({ locale }: { locale: string }) {
  const [user, setUser] = useState<SignedInUser | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("clubmax_access_token");
    if (!token) return;

    const cached = localStorage.getItem("clubmax_user");
    if (cached) {
      try {
        setUser(JSON.parse(cached));
      } catch {
        localStorage.removeItem("clubmax_user");
      }
    }

    fetch(`${API}/v1/users/me`, { headers: { Authorization: `Bearer ${token}` } })
      .then(async (response) => {
        if (!response.ok) throw new Error("session expired");
        const currentUser = await response.json();
        localStorage.setItem("clubmax_user", JSON.stringify(currentUser));
        setUser(currentUser);
      })
      .catch(() => {
        localStorage.removeItem("clubmax_access_token");
        localStorage.removeItem("clubmax_user");
        setUser(null);
      });
  }, []);

  if (!user) {
    return (
      <div className="home-auth-actions">
        <Link className="login-link" href={`/${locale}/login`}>
          ورود
        </Link>
        <Link className="join-link" href={`/${locale}/signup`}>
          ثبت‌نام رایگان
        </Link>
      </div>
    );
  }

  const name = `${user.first_name} ${user.last_name}`.trim();
  return (
    <Link className="home-account" href={`/${locale}/dashboard`}>
      <span className="home-account-avatar" aria-hidden="true">
        {user.first_name?.slice(0, 1) || "ک"}
      </span>
      <span>
        <small>خوش آمدید، {name || "کاربر کلاب‌ماکس"}</small>
        <b>{user.role === "admin" ? "پنل مدیریت" : "پنل کاربری"}</b>
      </span>
      <span className="home-account-arrow" aria-hidden="true">←</span>
    </Link>
  );
}
