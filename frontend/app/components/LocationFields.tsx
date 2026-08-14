"use client";

import {useCallback, useEffect, useState} from "react";

const API = process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:8081";

async function loadLocations(province: string, signal: AbortSignal) {
  const query = province ? `?province=${encodeURIComponent(province)}` : "";
  const response = await fetch(`${API}/v1/meta/locations${query}`, {signal});
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json() as Promise<{provinces?: string[]; counties?: string[]}>;
}

export default function LocationFields({required = false, initialProvince = "", initialCounty = ""}: {required?: boolean; initialProvince?: string; initialCounty?: string}) {
  const [provinces, setProvinces] = useState<string[]>([]);
  const [province, setProvince] = useState(initialProvince);
  const [county, setCounty] = useState(initialCounty);
  const [counties, setCounties] = useState<string[]>([]);
  const [loadingProvinces, setLoadingProvinces] = useState(true);
  const [loadingCounties, setLoadingCounties] = useState(false);
  const [error, setError] = useState("");
  const [retry, setRetry] = useState(0);
  const retryLoad = useCallback(() => setRetry(value => value + 1), []);

  useEffect(() => {
    const controller = new AbortController();
    setLoadingProvinces(true);
    setError("");
    loadLocations("", controller.signal)
      .then(data => setProvinces(data.provinces ?? []))
      .catch(reason => {
        if (reason instanceof DOMException && reason.name === "AbortError") return;
        setError("دریافت فهرست استان‌ها انجام نشد. از روشن بودن بک‌اند مطمئن شوید.");
      })
      .finally(() => setLoadingProvinces(false));
    return () => controller.abort();
  }, [retry]);

  useEffect(() => {
    setCounty(province === initialProvince ? initialCounty : "");
    setCounties([]);
    if (!province) return;
    const controller = new AbortController();
    setLoadingCounties(true);
    setError("");
    loadLocations(province, controller.signal)
      .then(data => setCounties(data.counties ?? []))
      .catch(reason => {
        if (reason instanceof DOMException && reason.name === "AbortError") return;
        setError("دریافت فهرست شهرستان‌ها انجام نشد.");
      })
      .finally(() => setLoadingCounties(false));
    return () => controller.abort();
  }, [province]);

  return <>
    <label>استان{required && <b className="required">*</b>}
      <select name="province" required={required} value={province} disabled={loadingProvinces} onChange={event => setProvince(event.target.value)}>
        <option value="">{loadingProvinces ? "در حال دریافت استان‌ها..." : "انتخاب استان"}</option>
        {provinces.map(item => <option key={item} value={item}>{item}</option>)}
      </select>
      {error && <span className="field-error">{error} <button type="button" onClick={retryLoad}>تلاش دوباره</button></span>}
    </label>
    <label>شهرستان{required && <b className="required">*</b>}
      <select name="city" required={required} disabled={!province || loadingCounties} value={county} onChange={event => setCounty(event.target.value)}>
        <option value="">{loadingCounties ? "در حال دریافت شهرستان‌ها..." : province ? "انتخاب شهرستان" : "ابتدا استان را انتخاب کنید"}</option>
        {counties.map(item => <option key={item} value={item}>{item}</option>)}
      </select>
    </label>
  </>;
}
