"use client";

import { useState } from "react";

function formatIBAN(value: string) {
  const digits = value.replace(/\D/g, "").slice(0, 24);
  const groups: string[] = [];
  if (digits.length) groups.push(digits.slice(0, 2));
  for (let index = 2; index < Math.min(digits.length, 22); index += 4) {
    groups.push(digits.slice(index, index + 4));
  }
  if (digits.length > 22) groups.push(digits.slice(22, 24));
  return groups.join("-");
}

export default function IranianIBANInput({
  name = "bank_iban",
  initialValue = "",
}: {
  name?: string;
  initialValue?: string;
}) {
  const [value, setValue] = useState(() => formatIBAN(initialValue.replace(/^IR/i, "")));
  return (
    <div className="iban-input" dir="ltr">
      <span aria-hidden="true">IR</span>
      <input
        name={name}
        value={value}
        onChange={(event) => setValue(formatIBAN(event.target.value))}
        inputMode="numeric"
        autoComplete="off"
        maxLength={30}
        placeholder="00-0000-0000-0000-0000-0000-00"
        aria-label="۲۴ رقم شماره شبا"
      />
    </div>
  );
}
