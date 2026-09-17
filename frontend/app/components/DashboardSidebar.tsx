"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import SiteLogo from "./SiteLogo";
import "../dashboard-sidebar.css";

const API = "";
type SidebarUser = {
  role: "admin" | "coach" | "athlete";
  first_name: string;
  last_name: string;
  face_image_id?: string;
};
type ItemKey =
  | "dashboard"
  | "users"
  | "gyms"
  | "finance"
  | "payments"
  | "withdrawals"
  | "backups"
  | "wallet"
  | "bmi"
  | "tickets"
  | "profile";
const roleNames = { admin: "مدیر سایت", coach: "مربی", athlete: "ورزشکار" };

function NavIcon({
  name,
}: {
  name:
    | "dashboard"
    | "users"
    | "gyms"
    | "wallet"
    | "explore"
    | "payments"
    | "messages"
    | "profile"
    | "backup"
    | "tools";
}) {
  const paths = {
    dashboard: (
      <>
        <rect x="3" y="3" width="7" height="7" />
        <rect x="14" y="3" width="7" height="7" />
        <rect x="3" y="14" width="7" height="7" />
        <rect x="14" y="14" width="7" height="7" />
      </>
    ),
    users: (
      <>
        <circle cx="12" cy="8" r="3.5" />
        <path d="M4.5 21c.7-4 3.2-6 7.5-6s6.8 2 7.5 6" />
      </>
    ),
    gyms: (
      <>
        <path d="m3 10 9-7 9 7v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1Z" />
        <path d="M9 21v-6h6v6M8 11h.01M16 11h.01" />
      </>
    ),
    wallet: (
      <>
        <path d="M4 7h15a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V6a3 3 0 0 1 3-3h12" />
        <path d="M16 14h5M16 14h.01" />
      </>
    ),
    explore: (
      <>
        <circle cx="11" cy="11" r="6.5" />
        <path d="m16 16 5 5M8 11h6M11 8v6" />
      </>
    ),
    payments: (
      <>
        <path d="M6 3h9l4 4v14H6a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z" />
        <path d="M15 3v5h5M8 13h8M8 17h5" />
      </>
    ),
    messages: (
      <>
        <path d="M20 15a3 3 0 0 1-3 3H9l-5 3v-3.5A3 3 0 0 1 2 15V6a3 3 0 0 1 3-3h12a3 3 0 0 1 3 3Z" />
        <path d="M7 9h10M7 13h6" />
      </>
    ),
    profile: (
      <>
        <circle cx="12" cy="8" r="4" />
        <path d="M4 21c.8-4.1 3.5-6 8-6s7.2 1.9 8 6" />
      </>
    ),
    backup: (
      <>
        <ellipse cx="12" cy="5" rx="7" ry="3" />
        <path d="M5 5v6c0 1.7 3.1 3 7 3 .7 0 1.4 0 2-.1M5 11v6c0 1.7 3.1 3 7 3" />
        <path d="M18 12v8M15 17l3 3 3-3" />
      </>
    ),
    tools: (
      <>
        <path d="M14.7 6.3a4 4 0 0 0-5-3.8l2.1 2.1-2.8 2.8-2.1-2.1a4 4 0 0 0 5.3 5.3L20 18.5a1.8 1.8 0 0 1-2.5 2.5l-7.8-7.8a4 4 0 0 0-5.3-5.3L6.5 10l2.8-2.8Z" />
      </>
    ),
  };
  return (
    <svg className={`nav-icon nav-icon-${name}`} viewBox="0 0 24 24" aria-hidden="true">
      {paths[name]}
    </svg>
  );
}

export default function DashboardSidebar({
  user,
  locale,
  active,
}: {
  user: SidebarUser;
  locale: string;
  active: ItemKey;
}) {
  const router = useRouter(),
    [unreadTickets, setUnreadTickets] = useState(0),
    [mobileOpen, setMobileOpen] = useState(false),
    initials = user.first_name[0] + (user.last_name?.[0] || ""),
    profileImage = user.face_image_id ? `${API}/media/${user.face_image_id}` : "";
  const refreshUnreadTickets = useCallback(async () => {
    const token = localStorage.getItem("clubmax_access_token");
    if (!token) return;
    try {
      const response = await fetch(`${API}/v1/tickets/unread-count`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (response.ok) {
        const data = await response.json();
        setUnreadTickets(Number(data.count) || 0);
      }
    } catch {
      // A transient badge refresh failure should not affect dashboard navigation.
    }
  }, []);
  useEffect(() => {
    refreshUnreadTickets();
    const interval = window.setInterval(refreshUnreadTickets, 30000);
    window.addEventListener("focus", refreshUnreadTickets);
    window.addEventListener("clubmax:tickets-read", refreshUnreadTickets);
    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", refreshUnreadTickets);
      window.removeEventListener("clubmax:tickets-read", refreshUnreadTickets);
    };
  }, [refreshUnreadTickets]);
  useEffect(() => {
    document.body.classList.toggle("mobile-nav-open", mobileOpen);
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setMobileOpen(false);
    };
    window.addEventListener("keydown", closeOnEscape);
    return () => {
      document.body.classList.remove("mobile-nav-open");
      window.removeEventListener("keydown", closeOnEscape);
    };
  }, [mobileOpen]);
  const item = (
    key: ItemKey,
    href: string,
    label: string,
    icon: Parameters<typeof NavIcon>[0]["name"]
  ) => (
    <Link className={active === key ? "active" : ""} href={href}>
      <NavIcon name={icon} />
      <span>{label}</span>
      {key === "tickets" && unreadTickets > 0 && (
        <b className="nav-unread-badge" aria-label={`${unreadTickets} تیکت خوانده‌نشده`}>
          {unreadTickets > 99 ? "+۹۹" : unreadTickets.toLocaleString("fa-IR")}
        </b>
      )}
    </Link>
  );
  const plain = (href: string, label: string, icon: Parameters<typeof NavIcon>[0]["name"]) => (
    <Link href={href}>
      <NavIcon name={icon} />
      <span>{label}</span>
    </Link>
  );
  const mobileItem = (
    key: ItemKey | "explore",
    href: string,
    label: string,
    icon: Parameters<typeof NavIcon>[0]["name"]
  ) => (
    <Link
      className={active === key ? "active" : ""}
      href={href}
      onClick={() => setMobileOpen(false)}
    >
      <NavIcon name={icon} />
      <span>{label}</span>
      {key === "tickets" && unreadTickets > 0 && (
        <b className="mobile-nav-badge">
          {unreadTickets > 9 ? "+۹" : unreadTickets.toLocaleString("fa-IR")}
        </b>
      )}
    </Link>
  );
  const logout = () => {
    localStorage.clear();
    router.replace(`/${locale}/login`);
  };
  return (
    <aside>
      <SiteLogo href={`/${locale}`} priority />
      <button
        type="button"
        className="mobile-menu-toggle"
        onClick={() => setMobileOpen(true)}
        aria-label="بازکردن منوی کاربری"
        aria-expanded={mobileOpen}
      >
        <span className="mobile-user-avatar">
          {profileImage ? <img src={profileImage} alt="" /> : initials}
        </span>
        <span className="mobile-user-copy">
          <b>{user.first_name}</b>
          <small>{roleNames[user.role]}</small>
        </span>
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d="M5 7h14M5 12h14M5 17h14" />
        </svg>
      </button>
      <div className="identity">
        <div className="profile-thumbnail">
          {profileImage ? (
            <img src={profileImage} alt={`عکس ${user.first_name} ${user.last_name}`} />
          ) : (
            initials
          )}
        </div>
        <b>
          {user.first_name} {user.last_name}
        </b>
        <span>{roleNames[user.role]}</span>
      </div>
      <nav>
        {item("dashboard", `/${locale}/dashboard`, "داشبورد", "dashboard")}
        {user.role === "admin" &&
          item("users", `/${locale}/admin/users`, "مدیریت کاربران", "users")}
        {user.role === "admin" && item("gyms", `/${locale}/admin/gyms`, "مدیریت باشگاه‌ها", "gyms")}
        {user.role === "admin" &&
          item("finance", `/${locale}/admin/finance`, "مدیریت مالی", "wallet")}
        {user.role === "admin" &&
          item("backups", `/${locale}/admin/backups`, "پشتیبان‌گیری", "backup")}
        {user.role === "coach" && item("gyms", `/${locale}/gyms`, "باشگاه‌های من", "gyms")}
        {user.role === "coach" && item("wallet", `/${locale}/wallet`, "کیف پول و برداشت", "wallet")}
        {user.role === "athlete" && plain(`/${locale}/explore-gyms`, "باشگاه‌ها", "explore")}
        {user.role === "athlete" &&
          item("payments", `/${locale}/payments`, "پرداخت‌های من", "payments")}
        {user.role === "athlete" && (
          <div className="sidebar-nav-group">
            <div className="sidebar-nav-group-title">
              <NavIcon name="tools" />
              <span>ابزار</span>
            </div>
            {item("bmi", `/${locale}/tools/bmi`, "شاخص توده بدنی", "tools")}
          </div>
        )}
        {item(
          "tickets",
          `/${locale}/tickets`,
          user.role === "admin" ? "مدیریت تیکت‌ها" : "پشتیبانی و تیکت‌ها",
          "messages"
        )}
        {item("profile", `/${locale}/profile`, "پروفایل من", "profile")}
      </nav>
      <button className="logout" onClick={logout}>
        خروج از حساب
      </button>
      <nav className="mobile-bottom-nav" aria-label="دسترسی سریع موبایل">
        {mobileItem("dashboard", `/${locale}/dashboard`, "خانه", "dashboard")}
        {user.role === "admin" && mobileItem("users", `/${locale}/admin/users`, "کاربران", "users")}
        {user.role === "admin" && mobileItem("gyms", `/${locale}/admin/gyms`, "باشگاه‌ها", "gyms")}
        {user.role === "admin" &&
          mobileItem("finance", `/${locale}/admin/finance`, "مالی", "wallet")}
        {user.role === "coach" && mobileItem("gyms", `/${locale}/gyms`, "باشگاه‌ها", "gyms")}
        {user.role === "coach" && mobileItem("wallet", `/${locale}/wallet`, "کیف پول", "wallet")}
        {user.role === "coach" &&
          mobileItem("tickets", `/${locale}/tickets`, "پشتیبانی", "messages")}
        {user.role === "athlete" &&
          mobileItem("explore", `/${locale}/explore-gyms`, "باشگاه‌ها", "explore")}
        {user.role === "athlete" && mobileItem("bmi", `/${locale}/tools/bmi`, "BMI", "tools")}
        {user.role === "athlete" &&
          mobileItem("payments", `/${locale}/payments`, "پرداخت‌ها", "payments")}
        <button
          type="button"
          className={mobileOpen ? "active" : ""}
          onClick={() => setMobileOpen(true)}
        >
          <svg className="nav-icon nav-icon-more" viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="5" cy="12" r="1.5" />
            <circle cx="12" cy="12" r="1.5" />
            <circle cx="19" cy="12" r="1.5" />
          </svg>
          <span>بیشتر</span>
        </button>
      </nav>
      <button
        type="button"
        className={`mobile-drawer-backdrop${mobileOpen ? " open" : ""}`}
        aria-label="بستن منو"
        tabIndex={mobileOpen ? 0 : -1}
        onClick={() => setMobileOpen(false)}
      />
      <section
        className={`mobile-nav-drawer${mobileOpen ? " open" : ""}`}
        role="dialog"
        aria-modal="true"
        aria-label="منوی کامل کاربری"
      >
        <div className="mobile-drawer-handle" />
        <header>
          <div className="mobile-drawer-identity">
            <span className="mobile-user-avatar">
              {profileImage ? <img src={profileImage} alt="" /> : initials}
            </span>
            <div>
              <b>
                {user.first_name} {user.last_name}
              </b>
              <small>{roleNames[user.role]}</small>
            </div>
          </div>
          <button type="button" onClick={() => setMobileOpen(false)} aria-label="بستن منو">
            ×
          </button>
        </header>
        <div className="mobile-drawer-grid">
          {mobileItem("dashboard", `/${locale}/dashboard`, "داشبورد", "dashboard")}
          {user.role === "admin" &&
            mobileItem("users", `/${locale}/admin/users`, "مدیریت کاربران", "users")}
          {user.role === "admin" &&
            mobileItem("gyms", `/${locale}/admin/gyms`, "مدیریت باشگاه‌ها", "gyms")}
          {user.role === "admin" &&
            mobileItem("finance", `/${locale}/admin/finance`, "مدیریت مالی", "wallet")}
          {user.role === "admin" &&
            mobileItem("backups", `/${locale}/admin/backups`, "پشتیبان‌گیری", "backup")}
          {user.role === "coach" && mobileItem("gyms", `/${locale}/gyms`, "باشگاه‌های من", "gyms")}
          {user.role === "coach" &&
            mobileItem("wallet", `/${locale}/wallet`, "کیف پول و برداشت", "wallet")}
          {user.role === "athlete" &&
            mobileItem("explore", `/${locale}/explore-gyms`, "جستجوی باشگاه", "explore")}
          {user.role === "athlete" &&
            mobileItem("payments", `/${locale}/payments`, "پرداخت‌های من", "payments")}
          {user.role === "athlete" &&
            mobileItem("bmi", `/${locale}/tools/bmi`, "شاخص توده بدنی", "tools")}
          {mobileItem(
            "tickets",
            `/${locale}/tickets`,
            user.role === "admin" ? "مدیریت تیکت‌ها" : "پشتیبانی و تیکت",
            "messages"
          )}
          {mobileItem("profile", `/${locale}/profile`, "پروفایل من", "profile")}
          <Link href={`/${locale}`} onClick={() => setMobileOpen(false)}>
            <NavIcon name="explore" />
            <span>صفحه اصلی سایت</span>
          </Link>
        </div>
        <button type="button" className="mobile-logout" onClick={logout}>
          خروج از حساب کاربری
        </button>
      </section>
    </aside>
  );
}
