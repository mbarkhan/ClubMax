"use client";

import {useState} from "react";

export default function PasswordField({name, label, autoComplete = "new-password"}: {name: string; label: string; autoComplete?: string}) {
  const [visible, setVisible] = useState(false);
  return <label>{label}<b className="required">*</b>
    <span className="password-input">
      <input name={name} required type={visible ? "text" : "password"} dir="ltr" minLength={8} autoComplete={autoComplete}/>
      <button type="button" onClick={() => setVisible(value => !value)} aria-label={visible ? "پنهان کردن رمز عبور" : "نمایش رمز عبور"} title={visible ? "پنهان کردن رمز عبور" : "نمایش رمز عبور"} aria-pressed={visible}>
        {visible
          ? <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 3l18 18M10.6 10.7a2 2 0 002.7 2.7M9.9 4.2A10.7 10.7 0 0112 4c5.5 0 9 5.2 9 5.2a13 13 0 01-2.1 2.6M6.6 6.6A15.5 15.5 0 003 9.2s3.5 5.2 9 5.2a9.7 9.7 0 004-.8"/></svg>
          : <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 12s3.5-5.2 9-5.2S21 12 21 12s-3.5 5.2-9 5.2S3 12 3 12z"/><circle cx="12" cy="12" r="2.5"/></svg>}
      </button>
    </span>
  </label>;
}
