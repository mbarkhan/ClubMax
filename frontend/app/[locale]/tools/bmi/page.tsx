"use client";

import { FormEvent, use, useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import DashboardSidebar from "../../../components/DashboardSidebar";
import BMITrendChart, { bmiStatus, BodyMeasurement } from "../../../components/BMITrendChart";
import "../../../bmi.css";

const API = "";
type User = {
  role: "athlete";
  first_name: string;
  last_name: string;
  face_image_id?: string;
};
type BMIOverview = {
  measurements: BodyMeasurement[];
  latest?: BodyMeasurement;
  can_record: boolean;
  next_allowed_at?: string;
};

export default function BMIPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params);
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [overview, setOverview] = useState<BMIOverview | null>(null);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [saving, setSaving] = useState(false);
  const load = useCallback(async () => {
    const token = localStorage.getItem("clubmax_access_token") || "";
    const headers = { Authorization: `Bearer ${token}`, "Accept-Language": "fa" };
    try {
      const [profileResponse, bmiResponse] = await Promise.all([
        fetch(`${API}/v1/users/me`, { headers }),
        fetch(`${API}/v1/athlete/bmi`, { headers, cache: "no-store" }),
      ]);
      if (
        [profileResponse.status, bmiResponse.status].some(
          (status) => status === 401 || status === 403
        )
      ) {
        router.replace(`/${locale}/login`);
        return;
      }
      const profile = await profileResponse.json();
      const bmi = await bmiResponse.json();
      if (!profileResponse.ok) throw new Error(profile.message);
      if (profile.role !== "athlete") {
        router.replace(`/${locale}/dashboard`);
        return;
      }
      if (!bmiResponse.ok) throw new Error(bmi.message);
      setUser(profile);
      setOverview(bmi);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "دریافت اطلاعات BMI انجام نشد.");
    }
  }, [locale, router]);
  useEffect(() => {
    void load();
  }, [load]);
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const values = new FormData(form);
    setSaving(true);
    setError("");
    setNotice("");
    try {
      const response = await fetch(`${API}/v1/athlete/bmi/measurements`, {
          method: "POST",
          headers: {
            Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
            "Content-Type": "application/json",
            "Accept-Language": "fa",
          },
          body: JSON.stringify({
            height_cm: Number(values.get("height_cm")),
            weight_kg: Number(values.get("weight_kg")),
          }),
        }),
        data = await response.json();
      if (!response.ok) throw new Error(data.message || "ثبت اندازه‌گیری انجام نشد.");
      form.reset();
      setNotice("اندازه‌گیری جدید ثبت و نمودار به‌روزرسانی شد.");
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ثبت اندازه‌گیری انجام نشد.");
    } finally {
      setSaving(false);
    }
  }
  const change = useMemo(() => {
    if (!overview || overview.measurements.length < 2) return null;
    return overview.measurements[0].bmi - overview.measurements[1].bmi;
  }, [overview]);
  if (!user || !overview) {
    return (
      <div className="screen-loader">
        <span>{error || "در حال آماده‌سازی ابزار BMI..."}</span>
      </div>
    );
  }
  const latest = overview.latest;
  const currentStatus = latest ? bmiStatus[latest.status] || bmiStatus.age_specific : null;
  return (
    <div className="shell athlete-shell">
      <DashboardSidebar user={user} locale={locale} active="bmi" />
      <main className="dash bmi-page">
        <header>
          <div>
            <span>ابزارهای سلامت</span>
            <h1>شاخص توده بدنی</h1>
          </div>
        </header>
        <section className="bmi-hero">
          <div>
            <h2>قد و وزن خود را منظم ثبت کنید.</h2>
          </div>
          <div
            className="bmi-current"
            style={{ "--bmi-color": currentStatus?.color || "#84928b" } as React.CSSProperties}
          >
            <small>آخرین BMI</small>
            <strong>{latest ? latest.bmi.toLocaleString("fa-IR") : "—"}</strong>
            <span>{currentStatus?.label || "هنوز ثبت نشده"}</span>
          </div>
        </section>
        {error && <div className="alert">{error}</div>}
        {notice && <div className="bmi-notice">{notice}</div>}
        <section className="bmi-summary">
          <article>
            <span>آخرین قد</span>
            <strong>
              {latest ? `${latest.height_cm.toLocaleString("fa-IR")} سانتی‌متر` : "—"}
            </strong>
          </article>
          <article>
            <span>آخرین وزن</span>
            <strong>{latest ? `${latest.weight_kg.toLocaleString("fa-IR")} کیلوگرم` : "—"}</strong>
          </article>
          <article>
            <span>تغییر نسبت به ثبت قبل</span>
            <strong className={change === null ? "" : change > 0 ? "bmi-up" : "bmi-down"}>
              {change === null
                ? "—"
                : `${change > 0 ? "+" : ""}${change.toLocaleString("fa-IR", { maximumFractionDigits: 1 })}`}
            </strong>
          </article>
        </section>
        <div className="bmi-main-grid">
          <section className="bmi-panel bmi-trend-panel">
            <div className="bmi-panel-head">
              <div>
                <span>روند تغییرات</span>
                <h2>نمودار BMI</h2>
              </div>
              <small>۲۴ رکورد اخیر</small>
            </div>
            <BMITrendChart measurements={overview.measurements} />
          </section>
          <section className="bmi-panel bmi-entry-panel">
            <div className="bmi-panel-head">
              <div>
                <span>اندازه‌گیری جدید</span>
                <h2>ثبت قد و وزن</h2>
              </div>
            </div>
            <form key={latest?.id || "new"} onSubmit={submit}>
              <label>
                قد <span>سانتی‌متر</span>
                <input
                  name="height_cm"
                  type="number"
                  min="40"
                  max="250"
                  step="0.1"
                  inputMode="decimal"
                  required
                  disabled={!overview.can_record || saving}
                  placeholder="مثلاً ۱۷۵"
                  defaultValue={latest?.height_cm}
                />
              </label>
              <label>
                وزن <span>کیلوگرم</span>
                <input
                  name="weight_kg"
                  type="number"
                  min="5"
                  max="500"
                  step="0.1"
                  inputMode="decimal"
                  required
                  disabled={!overview.can_record || saving}
                  placeholder="مثلاً ۷۲٫۵"
                />
              </label>
              {overview.can_record ? (
                <p className="bmi-ready">امکان ثبت اندازه‌گیری جدید فعال است.</p>
              ) : (
                <p className="bmi-wait">
                  ثبت بعدی از{" "}
                  {new Date(overview.next_allowed_at!).toLocaleString("fa-IR-u-ca-persian", {
                    dateStyle: "long",
                    timeStyle: "short",
                  })}{" "}
                  امکان‌پذیر است.
                </p>
              )}
              <button disabled={!overview.can_record || saving}>
                {saving ? "در حال ثبت..." : "ثبت و محاسبه BMI"}
              </button>
            </form>
          </section>
        </div>
        <section className="bmi-scale" aria-label="راهنمای بازه‌های BMI">
          {latest?.reference === "cdc_bmi_for_age" ? (
            <>
              <div className="scale-blue">
                <b>کمبود وزن</b>
                <span>کمتر از صدک ۵</span>
              </div>
              <div className="scale-green">
                <b>محدوده سالم</b>
                <span>صدک ۵ تا کمتر از ۸۵</span>
              </div>
              <div className="scale-yellow">
                <b>اضافه‌وزن</b>
                <span>صدک ۸۵ تا کمتر از ۹۵</span>
              </div>
              <div className="scale-orange">
                <b>چاقی</b>
                <span>صدک ۹۵ و بیشتر</span>
              </div>
              <div className="scale-red">
                <b>چاقی شدید</b>
                <span>۱۲۰٪ صدک ۹۵ یا BMI حداقل ۳۵</span>
              </div>
            </>
          ) : (
            <>
              <div className="scale-blue">
                <b>کمبود وزن</b>
                <span>کمتر از ۱۸٫۵</span>
              </div>
              <div className="scale-green">
                <b>محدوده سالم</b>
                <span>۱۸٫۵ تا ۲۴٫۹</span>
              </div>
              <div className="scale-yellow">
                <b>اضافه‌وزن</b>
                <span>۲۵ تا ۲۹٫۹</span>
              </div>
              <div className="scale-orange">
                <b>چاقی درجه ۱</b>
                <span>۳۰ تا ۳۴٫۹</span>
              </div>
              <div className="scale-red">
                <b>چاقی درجه ۲ و ۳</b>
                <span>۳۵ و بیشتر</span>
              </div>
            </>
          )}
        </section>
        <section className="bmi-history bmi-panel">
          <div className="bmi-panel-head">
            <div>
              <span>تاریخچه اندازه‌گیری</span>
              <h2>قد، وزن و نتیجه محاسبه‌شده</h2>
            </div>
          </div>
          <div className="bmi-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>تاریخ ثبت</th>
                  <th>سن</th>
                  <th>قد</th>
                  <th>وزن</th>
                  <th>BMI</th>
                  <th>وضعیت</th>
                </tr>
              </thead>
              <tbody>
                {overview.measurements.map((item) => {
                  const status = bmiStatus[item.status] || bmiStatus.age_specific;
                  return (
                    <tr key={item.id}>
                      <td>
                        {new Date(item.measured_at).toLocaleString("fa-IR-u-ca-persian", {
                          dateStyle: "medium",
                          timeStyle: "short",
                        })}
                      </td>
                      <td>{item.age_years.toLocaleString("fa-IR")} سال</td>
                      <td>{item.height_cm.toLocaleString("fa-IR")} سانتی‌متر</td>
                      <td>{item.weight_kg.toLocaleString("fa-IR")} کیلوگرم</td>
                      <td>
                        <strong>{item.bmi.toLocaleString("fa-IR")}</strong>
                        {item.bmi_percentile !== undefined && (
                          <small>صدک {item.bmi_percentile.toLocaleString("fa-IR")}</small>
                        )}
                      </td>
                      <td>
                        <span
                          className="bmi-status"
                          style={{ color: status.color, backgroundColor: `${status.color}16` }}
                        >
                          {status.label}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {!overview.measurements.length && (
              <div className="bmi-history-empty">هنوز اندازه‌گیری ثبت نشده است.</div>
            )}
          </div>
        </section>
        <p className="bmi-disclaimer">
          BMI فقط یک ابزار غربالگری است و جایگزین ارزیابی پزشک یا متخصص تغذیه نیست. برای سنین ۲ تا
          ۱۹ سال، نتیجه با صدک ماهانه و جنسیت بر اساس جدول رسمی CDC تفسیر می‌شود؛ از ۲۰ سالگی
          بازه‌های ثابت بزرگسالان به‌کار می‌روند. برای جنسیت «سایر» در سن زیر ۲۰ سال، تفسیر بالینی
          لازم است.
        </p>
      </main>
    </div>
  );
}
