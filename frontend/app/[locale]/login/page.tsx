"use client";

import Link from "next/link";
import SiteLogo from "../../components/SiteLogo";
import { FormEvent, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import PasswordField from "../../components/PasswordField";
import "../../auth.css";

const API = "";
export default function Login() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter();
  const [error, setError] = useState(""),
    [loading, setLoading] = useState(false);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      const response = await fetch(`${API}/v1/auth/login`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "Accept-Language": locale },
          body: JSON.stringify({ identity: form.get("identity"), password: form.get("password") }),
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "نام کاربری یا رمز عبور صحیح نیست.");
      localStorage.setItem("clubmax_access_token", data.access_token);
      localStorage.setItem("clubmax_user", JSON.stringify(data.user));
      router.replace(`/${locale}/dashboard`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ارتباط با سرور برقرار نشد.");
    } finally {
      setLoading(false);
    }
  }
  return (
    <main className="auth-layout">
      <section className="auth-story">
        <SiteLogo href={`/${locale}`} priority />
        <div>
          <span className="eyebrow" style={{ fontSize: "1rem", fontWeight: "bold", color: "#5aa" }}>
            باشگاه، ساده‌تر از همیشه
          </span>
          <h1>
            مسیر ورزش
            <br />
            در یک فضای امن.
          </h1>
          <p>مدیریت اعضا، برنامه‌ها و پرداخت‌ها؛ همه در یک پلتفرم امن و ساده.</p>
        </div>
        <small>1.0.0</small>
      </section>
      <section className="auth-form-wrap">
        <form className="auth-form" onSubmit={submit}>
          <SiteLogo href={`/${locale}`} className="mobile-logo" priority />
          <span className="eyebrow">ورود به حساب</span>
          <h2>خوش آمدید</h2>
          <p>کد ملی خود را وارد کنید. مدیر سایت با نام کاربری خود وارد می‌شود.</p>
          <label>
            نام کاربری / کد ملی
            <input name="identity" required dir="ltr" autoComplete="username" />
          </label>
          <PasswordField name="password" label="رمز عبور" autoComplete="current-password" />
          {error && <div className="alert">{error}</div>}
          <button className="submit" disabled={loading}>
            {loading ? "در حال ورود..." : "ورود"}
          </button>
          <p className="form-footer">
            حساب ندارید؟ <Link href={`/${locale}/signup`}>ثبت‌نام کنید</Link>
          </p>
        </form>
      </section>
    </main>
  );
}
