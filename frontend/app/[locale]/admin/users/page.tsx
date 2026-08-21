"use client";
import { use, useCallback, useEffect, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = process.env.NEXT_PUBLIC_API_URL!;
type User = {
  id: string;
  role: "admin" | "coach" | "athlete" | "operator";
  first_name: string;
  last_name: string;
  phone?: string;
  national_id?: string;
  username?: string;
  active: boolean;
  created_at: string;
};
const roles = {
  admin: "مدیر سایت",
  coach: "مربی",
  athlete: "ورزشکار",
  operator: "متصدی (غیرفعال)",
};
export default function UsersPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [users, setUsers] = useState<User[]>([]),
    [query, setQuery] = useState(""),
    [page, setPage] = useState(0),
    [error, setError] = useState(""),
    [selected, setSelected] = useState<User | null>(null),
    [username, setUsername] = useState(""),
    [password, setPassword] = useState("");
  const load = useCallback(() => {
    const token = localStorage.getItem("gymax_access_token") || "";
    fetch(`${API}/v1/users?limit=100&offset=${page * 100}&q=${encodeURIComponent(query)}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(async (r) => {
        const d = await r.json();
        if (!r.ok) throw new Error(d.message);
        setUsers(d || []);
      })
      .catch((e) => setError(e.message));
  }, [query, page]);
  useEffect(() => {
    const timer = setTimeout(load, 250);
    return () => clearTimeout(timer);
  }, [load]);
  const remove = async (user: User) => {
    if (!confirm(`حساب ${user.first_name} ${user.last_name} و اطلاعات وابسته حذف شود؟`)) return;
    const r = await fetch(`${API}/v1/users/${user.id}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${localStorage.getItem("gymax_access_token") || ""}` },
    });
    if (!r.ok) {
      const d = await r.json();
      setError(d.message);
      return;
    }
    load();
  };
  const promote = async () => {
    if (!selected) return;
    const r = await fetch(`${API}/v1/users/${selected.id}/promote`, {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("gymax_access_token") || ""}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username, password }),
    });
    const d = r.status === 204 ? {} : await r.json();
    if (!r.ok) {
      setError(d.message);
      return;
    }
    setSelected(null);
    setUsername("");
    setPassword("");
    load();
  };
  return (
    <AdminShell locale={locale} active="users" kicker="مدیریت دسترسی" title="کاربران سایت">
      <div className="admin-toolbar">
        <input
          value={query}
          onChange={(e) => {
            setQuery(e.target.value);
            setPage(0);
          }}
          placeholder="جستجو با نام، نام خانوادگی یا موبایل..."
        />
      </div>
      {error && <div className="alert">{error}</div>}
      <section className="admin-panel">
        <div className="admin-table-wrap">
          <table className="admin-table">
            <thead>
              <tr>
                <th>کاربر</th>
                <th>نقش</th>
                <th>شماره موبایل / نام کاربری</th>
                <th>کد ملی</th>
                <th>وضعیت</th>
                <th>تاریخ عضویت</th>
                <th>عملیات</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>
                    <strong>
                      {user.first_name} {user.last_name}
                    </strong>
                  </td>
                  <td>
                    <span
                      className={`admin-badge ${user.role === "admin" ? "warn" : user.role === "athlete" ? "info" : ""}`}
                    >
                      {roles[user.role]}
                    </span>
                  </td>
                  <td className="ltr">{user.phone || user.username || "—"}</td>
                  <td className="ltr">{user.national_id || "—"}</td>
                  <td>{user.active ? "فعال" : "غیرفعال"}</td>
                  <td>{new Date(user.created_at).toLocaleDateString("fa-IR")}</td>
                  <td>
                    <div className="admin-actions">
                      {user.role !== "admin" && (
                        <button onClick={() => setSelected(user)}>تبدیل به مدیر</button>
                      )}
                      {user.username === "admin" ? (
                        <span className="admin-badge warn">محافظت‌شده</span>
                      ) : (
                        <button className="danger" onClick={() => remove(user)}>
                          حذف
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {!users.length && !error && <div className="admin-empty">کاربری یافت نشد.</div>}
        <div className="admin-pagination">
          <button disabled={page === 0} onClick={() => setPage((value) => value - 1)}>
            صفحه قبل
          </button>
          <span>صفحه {(page + 1).toLocaleString("fa-IR")}</span>
          <button disabled={users.length < 100} onClick={() => setPage((value) => value + 1)}>
            صفحه بعد
          </button>
        </div>
      </section>
      {selected && (
        <div className="dialog-backdrop">
          <div className="admin-dialog">
            <h3>
              تبدیل {selected.first_name} {selected.last_name} به مدیر
            </h3>
            <p>پس از تغییر نقش، ورود این کاربر با نام کاربری و رمز مدیر انجام می‌شود.</p>
            <label>
              نام کاربری
              <input
                dir="ltr"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="off"
              />
            </label>
            <label>
              رمز عبور جدید
              <input
                dir="ltr"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </label>
            <div className="admin-dialog-actions">
              <button className="primary-button" onClick={promote}>
                تأیید ارتقا
              </button>
              <button className="secondary-button" onClick={() => setSelected(null)}>
                انصراف
              </button>
            </div>
          </div>
        </div>
      )}
    </AdminShell>
  );
}
