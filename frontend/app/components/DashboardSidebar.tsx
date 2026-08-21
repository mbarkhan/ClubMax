"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import SiteLogo from "./SiteLogo";
import "../dashboard-sidebar.css";

const API = process.env.NEXT_PUBLIC_API_URL!;
type SidebarUser = {
  role: "admin" | "coach" | "athlete";
  first_name: string;
  last_name: string;
  face_image_id?: string;
};
type ItemKey =
  "dashboard" | "users" | "gyms" | "finance" | "payments" | "withdrawals" | "wallet" | "profile";
const roleNames = { admin: "مدیر سایت", coach: "مربی", athlete: "ورزشکار" };

function NavIcon({
  name,
}: {
  name: "dashboard" | "users" | "gyms" | "wallet" | "explore" | "payments" | "messages" | "profile";
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
    initials = user.first_name[0] + (user.last_name?.[0] || ""),
    profileImage = user.face_image_id ? `${API}/media/${user.face_image_id}` : "";
  const item = (
    key: ItemKey,
    href: string,
    label: string,
    icon: Parameters<typeof NavIcon>[0]["name"]
  ) => (
    <Link className={active === key ? "active" : ""} href={href}>
      <NavIcon name={icon} />
      <span>{label}</span>
    </Link>
  );
  const plain = (href: string, label: string, icon: Parameters<typeof NavIcon>[0]["name"]) => (
    <Link href={href}>
      <NavIcon name={icon} />
      <span>{label}</span>
    </Link>
  );
  return (
    <aside>
      <SiteLogo href={`/${locale}`} priority />
      <div className="mobile-dashboard-links">
        <Link className={active === "dashboard" ? "active" : ""} href={`/${locale}/dashboard`}>
          داشبورد
        </Link>
        {user.role === "admin" && (
          <>
            <Link className={active === "users" ? "active" : ""} href={`/${locale}/admin/users`}>
              کاربران
            </Link>
            <Link className={active === "gyms" ? "active" : ""} href={`/${locale}/admin/gyms`}>
              باشگاه‌ها
            </Link>
            <Link
              className={active === "finance" ? "active" : ""}
              href={`/${locale}/admin/finance`}
            >
              مالی
            </Link>
          </>
        )}
        {user.role === "coach" && (
          <>
            <Link className={active === "gyms" ? "active" : ""} href={`/${locale}/gyms`}>
              باشگاه‌ها
            </Link>
            <Link className={active === "wallet" ? "active" : ""} href={`/${locale}/wallet`}>
              کیف پول
            </Link>
          </>
        )}
        {user.role === "athlete" && (
          <Link className={active === "payments" ? "active" : ""} href={`/${locale}/payments`}>
            پرداخت‌ها
          </Link>
        )}
        <Link className={active === "profile" ? "active" : ""} href={`/${locale}/profile`}>
          پروفایل
        </Link>
      </div>
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
        {user.role === "coach" && item("gyms", `/${locale}/gyms`, "باشگاه‌های من", "gyms")}
        {user.role === "coach" && item("wallet", `/${locale}/wallet`, "کیف پول و برداشت", "wallet")}
        {user.role === "athlete" && plain(`/${locale}/explore-gyms`, "باشگاه‌ها", "explore")}
        {user.role === "athlete" &&
          item("payments", `/${locale}/payments`, "پرداخت‌های من", "payments")}
        {plain(`/${locale}/dashboard`, "پیام‌ها", "messages")}
        {item("profile", `/${locale}/profile`, "پروفایل من", "profile")}
      </nav>
      <button
        className="logout"
        onClick={() => {
          localStorage.clear();
          router.replace(`/${locale}/login`);
        }}
      >
        خروج از حساب
      </button>
    </aside>
  );
}
