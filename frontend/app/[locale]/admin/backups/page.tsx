"use client";

import { use, useCallback, useEffect, useState } from "react";
import AdminShell from "../../../components/AdminShell";
import "../../../admin.css";

const API = process.env.NEXT_PUBLIC_API_URL!;

type BackupFile = { name: string; size: number; created_at: string };
type BackupStatus = {
  timezone: string;
  schedule: string[];
  schedule_minutes: number[];
  retention: number;
  next_run: string;
  files: BackupFile[];
};

const formatSize = (bytes: number) => {
  if (bytes < 1024) return `${bytes.toLocaleString("fa-IR")} بایت`;
  if (bytes < 1024 * 1024)
    return `${(bytes / 1024).toLocaleString("fa-IR", { maximumFractionDigits: 1 })} کیلوبایت`;
  return `${(bytes / 1024 / 1024).toLocaleString("fa-IR", { maximumFractionDigits: 1 })} مگابایت`;
};

const formatDate = (value: string) =>
  new Intl.DateTimeFormat("fa-IR-u-ca-persian", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Tehran",
  }).format(new Date(value));

export default function BackupsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = use(params);
  const [status, setStatus] = useState<BackupStatus | null>(null);
  const [error, setError] = useState("");
  const [downloading, setDownloading] = useState(false);
  const [restoreFile, setRestoreFile] = useState<File | null>(null);
  const [restoring, setRestoring] = useState(false);
  const [success, setSuccess] = useState("");
  const [schedule, setSchedule] = useState(["06:00", "14:00", "23:00"]);
  const [retention, setRetention] = useState(42);
  const [savingPolicy, setSavingPolicy] = useState(false);
  const [includeFiles, setIncludeFiles] = useState(true);

  const load = useCallback(async () => {
    const response = await fetch(`${API}/v1/admin/backups`, {
      headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
    });
    const data = await response.json();
    if (!response.ok) throw new Error(data.message || "دریافت وضعیت پشتیبان‌گیری انجام نشد.");
    setStatus(data);
    setSchedule(data.schedule || []);
    setRetention(data.retention);
  }, []);

  useEffect(() => {
    load().catch((reason) =>
      setError(reason instanceof Error ? reason.message : "دریافت اطلاعات انجام نشد.")
    );
  }, [load]);

  const download = async () => {
    setDownloading(true);
    setError("");
    try {
      const response = await fetch(
        `${API}/v1/admin/backups/download?include_files=${includeFiles}`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
          },
        }
      );
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.message || "تهیه پشتیبان کامل انجام نشد.");
      }
      const disposition = response.headers.get("Content-Disposition") || "";
      const match = disposition.match(/filename="?([^";]+)"?/i);
      const name = match?.[1] || "clubmax-backup.tar.gz";
      const url = URL.createObjectURL(await response.blob());
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = name;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      URL.revokeObjectURL(url);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "تهیه پشتیبان کامل انجام نشد.");
    } finally {
      setDownloading(false);
    }
  };

  const savePolicy = async () => {
    if (!schedule.length || schedule.some((value) => !/^([01]\d|2[0-3]):[0-5]\d$/.test(value))) {
      setError("حداقل یک زمان معتبر برای پشتیبان‌گیری وارد کنید.");
      return;
    }
    if (!Number.isInteger(retention) || retention < 1 || retention > 1000) {
      setError("تعداد نسخه‌های نگهداری باید بین ۱ تا ۱۰۰۰ باشد.");
      return;
    }
    const minutes = schedule.map((value) => {
      const [hour, minute] = value.split(":").map(Number);
      return hour * 60 + minute;
    });
    setSavingPolicy(true);
    setError("");
    setSuccess("");
    try {
      const response = await fetch(`${API}/v1/admin/backups/policy`, {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ schedule_minutes: minutes, retention }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.message || "ذخیره تنظیمات انجام نشد.");
      setSuccess("تنظیمات پشتیبان‌گیری خودکار ذخیره شد.");
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "ذخیره تنظیمات انجام نشد.");
    } finally {
      setSavingPolicy(false);
    }
  };

  const restore = async () => {
    if (!restoreFile) {
      setError("ابتدا فایل پشتیبان کامل را انتخاب کنید.");
      return;
    }
    if (
      !window.confirm(
        "تمام اطلاعات فعلی پایگاه داده و فایل‌های آپلودشده جایگزین می‌شوند. از ادامه مطمئن هستید؟"
      )
    )
      return;
    setRestoring(true);
    setError("");
    setSuccess("");
    try {
      const body = new FormData();
      body.append("backup", restoreFile);
      const response = await fetch(`${API}/v1/admin/backups/restore`, {
        method: "POST",
        headers: { Authorization: `Bearer ${localStorage.getItem("clubmax_access_token") || ""}` },
        body,
      });
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.message || "بازیابی پشتیبان انجام نشد.");
      setSuccess(data.message || "اطلاعات با موفقیت بازیابی شدند.");
      setRestoreFile(null);
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "بازیابی پشتیبان انجام نشد.");
    } finally {
      setRestoring(false);
    }
  };

  return (
    <AdminShell locale={locale} active="backups" kicker="حفاظت از اطلاعات" title="پشتیبان‌گیری">
      {error && <div className="alert backup-alert">{error}</div>}
      {success && <div className="success-message backup-alert">{success}</div>}
      <div className="admin-grid backup-stats">
        <article className="admin-stat">
          <span>زمان اجرای خودکار</span>
          <strong>{status?.schedule.join("، ") || "—"}</strong>
          <small>هر روز به وقت ایران</small>
        </article>
        <article className="admin-stat">
          <span>اجرای بعدی</span>
          <strong className="backup-date">{status ? formatDate(status.next_run) : "—"}</strong>
          <small>منطقه زمانی Asia/Tehran</small>
        </article>
        <article className="admin-stat">
          <span> نگهداری</span>
          <strong>{status ? status.retention.toLocaleString("fa-IR") : "—"}</strong>
          <small>آخرین فایل پشتیبان پایگاه داده</small>
        </article>
      </div>

      <section className="admin-panel backup-policy-panel">
        <div className="admin-section-title">
          <div>
            <h2>تنظیم اجرای خودکار</h2>
            <p>زمان‌ها بر اساس ساعت رسمی ایران محاسبه می‌شوند.</p>
          </div>
          <button className="primary-button" disabled={savingPolicy} onClick={savePolicy}>
            {savingPolicy ? "در حال ذخیره..." : "ذخیره تنظیمات"}
          </button>
        </div>
        <div className="backup-policy-form">
          <div className="backup-time-settings">
            <span className="backup-field-label">زمان‌های اجرا</span>
            <div className="backup-time-list">
              {schedule.map((value, index) => (
                <div className="backup-time-row" key={`${index}-${value}`}>
                  <input
                    type="time"
                    value={value}
                    onChange={(event) =>
                      setSchedule(
                        schedule.map((item, itemIndex) =>
                          itemIndex === index ? event.target.value : item
                        )
                      )
                    }
                  />
                  <button
                    type="button"
                    aria-label="حذف زمان"
                    disabled={schedule.length === 1}
                    onClick={() =>
                      setSchedule(schedule.filter((_, itemIndex) => itemIndex !== index))
                    }
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                className="backup-add-time"
                type="button"
                disabled={schedule.length >= 24}
                onClick={() => setSchedule([...schedule, "12:00"])}
              >
                + افزودن زمان
              </button>
            </div>
          </div>
          <label className="backup-retention-field">
            <span className="backup-field-label">تعداد نسخه‌های قابل نگهداری</span>
            <input
              type="number"
              min="1"
              max="1000"
              value={retention}
              onChange={(event) => setRetention(Number(event.target.value))}
            />
            <small>با کاهش این مقدار، نسخه‌های قدیمی‌تر حذف می‌شوند.</small>
          </label>
        </div>
      </section>

      <section className="admin-panel backup-download-panel">
        <div>
          <h2>تهیه فایل پشتیبان قابل دانلود</h2>
          <p>
            پایگاه داده همیشه در آرشیو قرار می‌گیرد. در صورت نیاز می‌توانید همه تصاویر و مدارک
            آپلودشده را نیز به آن اضافه کنید.
          </p>
          <label className="backup-include-files">
            <input
              type="checkbox"
              checked={includeFiles}
              onChange={(event) => setIncludeFiles(event.target.checked)}
            />
            <span>پشتیبان‌گیری از کلیه فایل‌های آپلودشده</span>
          </label>
          <small>این فایل حاوی اطلاعات حساس کاربران است؛ آن را فقط در محل امن نگهداری کنید.</small>
        </div>
        <button className="primary-button" disabled={downloading} onClick={download}>
          {downloading ? "در حال تهیه پشتیبان..." : "تهیه و دانلود پشتیبان"}
        </button>
      </section>

      <section className="admin-panel backup-restore-panel">
        <div className="backup-restore-copy">
          <span className="admin-badge danger">عملیات حساس</span>
          <h2>بازیابی پشتیبان کامل</h2>
          <p>
            فایل <b>clubmax-*.tar.gz</b> را انتخاب کنید. اگر آرشیو شامل فایل‌های آپلودی باشد، تصاویر
            و مدارک فعلی جایگزین می‌شوند؛ در پشتیبان فقط‌دیتابیس، فایل‌های فعلی بدون تغییر باقی
            می‌مانند.
          </p>
        </div>
        <div className="backup-restore-actions">
          <label className="backup-file-picker">
            <span>{restoreFile ? restoreFile.name : "انتخاب فایل پشتیبان"}</span>
            <input
              type="file"
              accept=".tar.gz,application/gzip"
              disabled={restoring}
              onChange={(event) => setRestoreFile(event.target.files?.[0] || null)}
            />
          </label>
          <button className="restore-button" disabled={!restoreFile || restoring} onClick={restore}>
            {restoring ? "در حال بازیابی..." : "بازیابی کامل اطلاعات"}
          </button>
        </div>
      </section>

      <section className="admin-panel backup-history">
        <div className="admin-section-title">
          <h2>آخرین پشتیبان‌های خودکار پایگاه داده</h2>
          <span>{status?.files.length.toLocaleString("fa-IR") || "۰"} فایل اخیر</span>
        </div>
        {!status ? (
          <p className="backup-empty">در حال دریافت اطلاعات...</p>
        ) : status.files.length === 0 ? (
          <p className="backup-empty">هنوز پشتیبان خودکاری ایجاد نشده است.</p>
        ) : (
          <div className="admin-table-wrap">
            <table className="admin-table backup-table">
              <thead>
                <tr>
                  <th>نام فایل</th>
                  <th>زمان ایجاد</th>
                  <th>حجم</th>
                  <th>وضعیت</th>
                </tr>
              </thead>
              <tbody>
                {status.files.map((file) => (
                  <tr key={file.name}>
                    <td className="ltr">{file.name}</td>
                    <td>{formatDate(file.created_at)}</td>
                    <td>{formatSize(file.size)}</td>
                    <td>
                      <span className="admin-badge">ذخیره‌شده</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </AdminShell>
  );
}
