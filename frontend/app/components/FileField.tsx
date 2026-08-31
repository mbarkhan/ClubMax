"use client";

import { ChangeEvent, useEffect, useState } from "react";

type Preview = { name: string; url?: string; kind: "image" | "pdf" | "file" };

export default function FileField({
  name,
  label,
  required = false,
  accept = "image/jpeg,image/png,image/webp,application/pdf",
  maxSizeKB = 500,
  className = "",
}: {
  name: string;
  label: string;
  maxFiles: number;
  required?: boolean;
  accept?: string;
  maxSizeKB?: number;
  className?: string;
}) {
  const [previews, setPreviews] = useState<Preview[]>([]);
  useEffect(
    () => () => previews.forEach((item) => item.url && URL.revokeObjectURL(item.url)),
    [previews]
  );

  function change(event: ChangeEvent<HTMLInputElement>) {
    const files = Array.from(event.target.files ?? []);
    if (files.some((file) => file.size > maxSizeKB * 1024)) {
      event.target.value = "";
      setPreviews([]);
      alert(`فایل با حجم حداکثر ${maxSizeKB.toLocaleString("fa-IR")} کیلوبایت انتخاب کنید.`);
      return;
    }
    setPreviews(
      files.map((file) => ({
        name: file.name,
        kind: file.type.startsWith("image/")
          ? "image"
          : file.type === "application/pdf"
            ? "pdf"
            : "file",
        url: file.type.startsWith("image/") ? URL.createObjectURL(file) : undefined,
      }))
    );
  }

  return (
    <label className={`file-field ${className}`.trim()}>
      <span>
        {label}
        {required && <b className="required">*</b>}
      </span>
      <input name={name} type="file" accept={accept} required={required} onChange={change} />
      <div className="file-prompt">
        <strong>انتخاب فایل</strong>
        <small>
          {previews.length
            ? `${previews.length} فایل انتخاب شد`
            : `حداکثر ${maxSizeKB.toLocaleString("fa-IR")} کیلوبایت`}
        </small>
      </div>
      {previews.length > 0 && (
        <div className="file-previews" aria-label="پیش‌نمایش فایل‌های انتخابی">
          {previews.map((item, index) => (
            <span className="file-preview" key={`${item.name}-${index}`} title={item.name}>
              {item.kind === "image" ? (
                <img src={item.url} alt={item.name} />
              ) : (
                <b>{item.kind === "pdf" ? "PDF" : "FILE"}</b>
              )}
            </span>
          ))}
        </div>
      )}
    </label>
  );
}
