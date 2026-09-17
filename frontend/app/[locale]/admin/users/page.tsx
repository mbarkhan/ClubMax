"use client";
import { use, useCallback, useEffect, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";
const API = "";
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
function ActionIcon({ name }: { name: "promote" | "reset" | "delete" }) {
  if (name === "promote") return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3 20 7v5c0 4.4-3.4 7.6-8 9-4.6-1.4-8-4.6-8-9V7l8-4Z"/><path d="M12 15V8m0 0-3 3m3-3 3 3"/></svg>;
  if (name === "reset") return <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="10" width="16" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/></svg>;
  return <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M10 11v6m4-6v6M9 7l1-3h4l1 3m-9 0 1 14h10l1-14"/></svg>;
}
export default function UsersPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params),
    [users, setUsers] = useState<User[]>([]),
    [query, setQuery] = useState(""),
    [page, setPage] = useState(0),
    [error, setError] = useState(""),
    [selected, setSelected] = useState<User | null>(null),
    [username, setUsername] = useState(""),
    [password, setPassword] = useState(""),
    [resetTarget, setResetTarget] = useState<User | null>(null),
    [resetPassword, setResetPassword] = useState(""),
    [resetConfirmation, setResetConfirmation] = useState(""),
    [notice, setNotice] = useState("");
  const load = useCallback(() => {
    const token = localStorage.getItem("clubmax_access_token") || "";
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
      headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
    });
    if (!r.ok) {
      const d = await r.json();
      setError(d.message);
      return;
    }
    load();
  };
  const resetPasswordByAdmin = async () => {
    if (!resetTarget) return;
    setError("");
    setNotice("");
    if (resetPassword.length < 8) {
      setError("رمز جدید باید حداقل ۸ کاراکتر باشد.");
      return;
    }
    if (resetPassword !== resetConfirmation) {
      setError("رمز جدید و تکرار آن یکسان نیستند.");
      return;
    }
    const response = await fetch(`${API}/v1/users/${resetTarget.id}/password-reset`, {
        method: "PUT",
        headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`, "Content-Type": "application/json" },
        body: JSON.stringify({ password: resetPassword }),
      }),
      data = await response.json();
    if (!response.ok) {
      setError(data.message || "بازنشانی رمز انجام نشد.");
      return;
    }
    setResetTarget(null);
    setResetPassword("");
    setResetConfirmation("");
    setNotice(data.message);
  };
  const promote = async () => {
    if (!selected) return;
    const r = await fetch(`${API}/v1/users/${selected.id}/promote`, {
      method: "PUT",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
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
      {notice && <div className="admin-notice">{notice}</div>}
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
                        <button className="admin-icon-button promote" onClick={() => setSelected(user)} aria-label="تبدیل به مدیر" title="تبدیل به مدیر">
                          <ActionIcon name="promote" />
                        </button>
                      )}
                      {user.username === "admin" ? (
                        <span className="admin-badge warn">محافظت‌شده</span>
                      ) : (
                        <>
                          <button className="admin-icon-button reset" onClick={() => setResetTarget(user)} aria-label="بازنشانی رمز" title="بازنشانی رمز">
                            <ActionIcon name="reset" />
                          </button>
                          <button className="admin-icon-button danger" onClick={() => remove(user)} aria-label="حذف کاربر" title="حذف کاربر">
                            <ActionIcon name="delete" />
                          </button>
                        </>
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
      {resetTarget && (
        <div className="dialog-backdrop">
          <div className="admin-dialog">
            <h3>بازنشانی رمز عبور</h3>
            <p>رمز دلخواه جدید برای {resetTarget.first_name} {resetTarget.last_name} را وارد کنید. این رمز فقط یک‌بار در اختیار شماست.</p>
            <label>
              رمز عبور جدید
              <input dir="ltr" type="password" minLength={8} value={resetPassword} onChange={(e) => setResetPassword(e.target.value)} autoComplete="new-password" />
            </label>
            <label>
              تکرار رمز عبور جدید
              <input dir="ltr" type="password" minLength={8} value={resetConfirmation} onChange={(e) => setResetConfirmation(e.target.value)} autoComplete="new-password" />
            </label>
            <div className="admin-dialog-actions">
              <button className="primary-button" onClick={resetPasswordByAdmin}>بازنشانی رمز</button>
              <button className="secondary-button" onClick={() => { setResetTarget(null); setResetPassword(""); setResetConfirmation(""); }}>انصراف</button>
            </div>
          </div>
        </div>
      )}
    </AdminShell>
  );
}
