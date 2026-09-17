"use client";

import {useState} from "react";

const ones = ["", "یک", "دو", "سه", "چهار", "پنج", "شش", "هفت", "هشت", "نه"];
const teens = ["ده", "یازده", "دوازده", "سیزده", "چهارده", "پانزده", "شانزده", "هفده", "هجده", "نوزده"];
const tens = ["", "", "بیست", "سی", "چهل", "پنجاه", "شصت", "هفتاد", "هشتاد", "نود"];
const hundreds = ["", "صد", "دویست", "سیصد", "چهارصد", "پانصد", "ششصد", "هفتصد", "هشتصد", "نهصد"];
const scales = ["", "هزار", "میلیون", "میلیارد", "تریلیون"];

function underThousand(number: number) {
  const parts: string[] = [];
  if (number >= 100) { parts.push(hundreds[Math.floor(number / 100)]); number %= 100; }
  if (number >= 20) { parts.push(tens[Math.floor(number / 10)]); number %= 10; if (number) parts.push(ones[number]); }
  else if (number >= 10) parts.push(teens[number - 10]);
  else if (number) parts.push(ones[number]);
  return parts.join(" و ");
}

function numberToPersianWords(number: number) {
  if (!number) return "صفر ریال";
  const parts: string[] = [];
  let scale = 0;
  while (number > 0 && scale < scales.length) {
    const group = number % 1000;
    if (group) parts.unshift(`${underThousand(group)}${scales[scale] ? ` ${scales[scale]}` : ""}`);
    number = Math.floor(number / 1000);
    scale++;
  }
  return `${parts.join(" و ")} ریال`;
}

export function TuitionAmountField({initialValue = 0}: {initialValue?: number}) {
  const [amount, setAmount] = useState(initialValue ? String(initialValue) : "");
  const numeric = Number(amount || 0);
  const words = numberToPersianWords(numeric);
  return <label>مبلغ شهریه (ریال)
    <span className="money-input-wrap">
      <input required inputMode="numeric" dir="ltr" value={amount ? numeric.toLocaleString("en-US") : ""} onChange={event => setAmount(event.target.value.replace(/\D/g, "").replace(/^0+(?=\d)/, ""))} placeholder="1,000,000" aria-describedby="tuition-words"/>
      {amount && <span className="money-tooltip" title={words} aria-label={words}>ⓘ</span>}
    </span>
    <input type="hidden" name="tuition_amount" value={amount}/>
    {amount && <small id="tuition-words" className="money-words">{words}</small>}
  </label>;
}

export function Time24Field({name, label, initialValue = ""}: {name: string; label: string; initialValue?: string}) {
  const [hour, setHour] = useState(initialValue.slice(0, 2));
  const [minute, setMinute] = useState(initialValue.slice(3, 5));
  const two = (value: number) => String(value).padStart(2, "0");
  return <label>{label} <b className="required">*</b>
    <span className="time-24-fields" dir="ltr">
      <select required value={hour} onChange={event => setHour(event.target.value)} aria-label={`${label}، ساعت`}><option value="">ساعت</option>{Array.from({length:24},(_,i)=><option key={i} value={two(i)}>{two(i)}</option>)}</select>
      <b>:</b>
      <select required value={minute} onChange={event => setMinute(event.target.value)} aria-label={`${label}، دقیقه`}><option value="">دقیقه</option>{Array.from({length:60},(_,i)=><option key={i} value={two(i)}>{two(i)}</option>)}</select>
    </span>
    <input type="hidden" name={name} value={hour && minute ? `${hour}:${minute}` : ""}/>
  </label>;
}
