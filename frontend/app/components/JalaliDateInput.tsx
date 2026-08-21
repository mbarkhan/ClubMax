"use client";

import { useMemo, useState } from "react";
import { createPortal } from "react-dom";
import { jalaaliMonthLength, toGregorian, toJalaali } from "jalaali-js";

const months = [
  "فروردین",
  "اردیبهشت",
  "خرداد",
  "تیر",
  "مرداد",
  "شهریور",
  "مهر",
  "آبان",
  "آذر",
  "دی",
  "بهمن",
  "اسفند",
];
const weekdays = ["ش", "ی", "د", "س", "چ", "پ", "ج"];
type JalaliValue = { y: number; m: number; d: number };

function parseInitial(value: string, minYear: number, maxYear: number): JalaliValue | null {
  if (!value) return null;
  const parts = value.slice(0, 10).split("-").map(Number);
  if (parts.length !== 3 || parts.some((part) => !Number.isFinite(part)) || parts[0] < 1000)
    return null;
  try {
    const date = toJalaali(parts[0], parts[1], parts[2]);
    return date.jy >= minYear && date.jy <= maxYear ? { y: date.jy, m: date.jm, d: date.jd } : null;
  } catch {
    return null;
  }
}

export default function JalaliDateInput({
  name,
  label = "تاریخ شمسی",
  required = false,
  minYear = 1320,
  maxYear = 1450,
  initialValue = "",
}: {
  name: string;
  label?: string;
  required?: boolean;
  minYear?: number;
  maxYear?: number;
  initialValue?: string;
}) {
  const today = useMemo(() => {
    const now = new Date();
    return toJalaali(now.getFullYear(), now.getMonth() + 1, now.getDate());
  }, []);
  const initial = useMemo(
    () => parseInitial(initialValue, minYear, maxYear),
    [initialValue, minYear, maxYear]
  );
  const [selected, setSelected] = useState<JalaliValue | null>(initial);
  const [viewYear, setViewYear] = useState(
    initial?.y ?? Math.min(maxYear, Math.max(minYear, today.jy))
  );
  const [viewMonth, setViewMonth] = useState(initial?.m ?? today.jm);
  const [open, setOpen] = useState(false);
  const monthLength = jalaaliMonthLength(viewYear, viewMonth);
  const firstGregorian = toGregorian(viewYear, viewMonth, 1);
  const firstWeekday =
    (new Date(firstGregorian.gy, firstGregorian.gm - 1, firstGregorian.gd).getDay() + 1) % 7;
  const cells = Array.from({ length: firstWeekday + monthLength }, (_, index) =>
    index < firstWeekday ? null : index - firstWeekday + 1
  );
  const iso = selected
    ? (() => {
        const date = toGregorian(selected.y, selected.m, selected.d);
        return `${date.gy}-${String(date.gm).padStart(2, "0")}-${String(date.gd).padStart(2, "0")}T00:00:00Z`;
      })()
    : "";
  const display = selected
    ? `${selected.y.toLocaleString("fa-IR", { useGrouping: false })}/${selected.m.toLocaleString("fa-IR", { minimumIntegerDigits: 2 })}/${selected.d.toLocaleString("fa-IR", { minimumIntegerDigits: 2 })}`
    : "";
  const moveMonth = (direction: number) => {
    let year = viewYear;
    let month = viewMonth + direction;
    if (month === 13) {
      month = 1;
      year += 1;
    }
    if (month === 0) {
      month = 12;
      year -= 1;
    }
    if (year >= minYear && year <= maxYear) {
      setViewYear(year);
      setViewMonth(month);
    }
  };
  return (
    <div className="jalali-field">
      <span>
        {label}
        {required && <b className="required">*</b>}
      </span>
      <div className="jalali-input-wrap">
        <input
          readOnly
          required={required}
          value={display}
          onClick={() => setOpen(true)}
          onFocus={() => setOpen(true)}
          placeholder="انتخاب تاریخ شمسی"
          aria-label={label}
        />
        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          aria-label="باز کردن تقویم"
        >
          ▦
        </button>
      </div>
      <input type="hidden" name={name} value={iso} />
      {open &&
        createPortal(
          <>
            <button
              className="jalali-dismiss"
              type="button"
              onClick={() => setOpen(false)}
              aria-label="بستن تقویم"
            />
            <div className="jalali-calendar" role="dialog" aria-label={`تقویم ${label}`}>
              <header>
                <button
                  type="button"
                  onClick={() => moveMonth(-1)}
                  disabled={viewYear === minYear && viewMonth === 1}
                >
                  ‹
                </button>
                <div>
                  <strong>{months[viewMonth - 1]}</strong>
                  <select
                    value={viewYear}
                    onChange={(event) => setViewYear(Number(event.target.value))}
                    aria-label="سال"
                  >
                    {Array.from(
                      { length: maxYear - minYear + 1 },
                      (_, index) => maxYear - index
                    ).map((year) => (
                      <option key={year}>{year}</option>
                    ))}
                  </select>
                </div>
                <button
                  type="button"
                  onClick={() => moveMonth(1)}
                  disabled={viewYear === maxYear && viewMonth === 12}
                >
                  ›
                </button>
              </header>
              <div className="jalali-weekdays">
                {weekdays.map((day) => (
                  <span key={day}>{day}</span>
                ))}
              </div>
              <div className="jalali-days">
                {cells.map((day, index) =>
                  day ? (
                    <button
                      type="button"
                      key={day}
                      className={
                        selected?.y === viewYear && selected.m === viewMonth && selected.d === day
                          ? "selected"
                          : ""
                      }
                      onClick={() => {
                        setSelected({ y: viewYear, m: viewMonth, d: day });
                        setOpen(false);
                      }}
                    >
                      {day.toLocaleString("fa-IR")}
                    </button>
                  ) : (
                    <i key={`empty-${index}`} />
                  )
                )}
              </div>
              <footer>
                <button
                  type="button"
                  onClick={() => {
                    setSelected(null);
                    setOpen(false);
                  }}
                >
                  پاک کردن
                </button>
                {today.jy >= minYear && today.jy <= maxYear && (
                  <button
                    type="button"
                    onClick={() => {
                      setSelected({ y: today.jy, m: today.jm, d: today.jd });
                      setViewYear(today.jy);
                      setViewMonth(today.jm);
                      setOpen(false);
                    }}
                  >
                    امروز
                  </button>
                )}
              </footer>
            </div>
          </>,
          document.body
        )}
    </div>
  );
}
