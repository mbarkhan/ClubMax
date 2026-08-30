"use client";
import Link from "next/link";
import SiteLogo from "../../components/SiteLogo";
import { FormEvent, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import LocationFields from "../../components/LocationFields";
import JalaliDateInput from "../../components/JalaliDateInput";
import FileField from "../../components/FileField";
import PasswordField from "../../components/PasswordField";
import IranianIBANInput from "../../components/IranianIBANInput";
import "../../auth.css";
const API = process.env.NEXT_PUBLIC_API_URL!;
const education = [
    ["below_diploma", "زیردیپلم"],
    ["diploma", "دیپلم"],
    ["associate", "فوق‌دیپلم"],
    ["bachelor", "لیسانس"],
    ["master", "فوق‌لیسانس"],
    ["doctorate", "دکتری"],
  ],
  blood = ["A+", "A-", "B+", "B-", "AB+", "AB-", "O+", "O-"],
  relations = ["پدر", "مادر", "برادر", "خواهر", "همسر", "فرزند", "سایر اقوام"];
function validNationalID(value: string) {
  if (!/^\d{10}$/.test(value) || /^(\d)\1{9}$/.test(value)) return false;
  let sum = 0;
  for (let index = 0; index < 9; index++) sum += Number(value[index]) * (10 - index);
  let remainder = sum % 11;
  if (remainder >= 2) remainder = 11 - remainder;
  return remainder === Number(value[9]);
}
function validIranianIBAN(value: string) {
  const normalized = value.replace(/[\s-]/g, "").toUpperCase();
  if (!/^IR\d{24}$/.test(normalized)) return false;
  const rearranged = normalized.slice(4) + "1827" + normalized.slice(2, 4);
  let remainder = 0;
  for (const digit of rearranged) remainder = (remainder * 10 + Number(digit)) % 97;
  return remainder === 1;
}
function Field({
  label,
  name,
  required = false,
  children,
}: {
  label: string;
  name: string;
  required?: boolean;
  children?: React.ReactNode;
}) {
  return (
    <label>
      {label}
      {required && <b className="required">*</b>}
      {children || <input name={name} required={required} />}
    </label>
  );
}
export default function Signup() {
  const { locale } = useParams<{ locale: string }>(),
    router = useRouter();
  const [role, setRole] = useState<"coach" | "athlete" | null>(null),
    [medical, setMedical] = useState(false),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(false);
  async function upload(token: string, kind: string, file: File) {
    const data = new FormData();
    data.append("file", file);
    const response = await fetch(`${API}/v1/users/me/documents/${kind}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: data,
    });
    if (!response.ok) {
      const result = await response.json();
      throw new Error(result.message || "بارگذاری مدارک کامل نشد.");
    }
  }
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    const formElement = event.currentTarget;
    formElement
      .querySelectorAll("[aria-invalid=true]")
      .forEach((item) => item.removeAttribute("aria-invalid"));
    const form = new FormData(formElement),
      value = (key: string) => String(form.get(key) || "").trim(),
      optional = (key: string) => value(key) || null,
      fail = (fieldName: string, message: string): never => {
        const field = formElement.elements.namedItem(fieldName);
        if (field instanceof HTMLElement) {
          field.setAttribute("aria-invalid", "true");
          field.focus();
          field.scrollIntoView({ behavior: "smooth", block: "center" });
        }
        throw new Error(message);
      };
    try {
      if (!role) throw new Error("انتخاب نقش مربی یا ورزشکار الزامی است.");
      const ibanDigits = value("bank_iban").replace(/\D/g, ""),
        bankIBAN = ibanDigits ? `IR${ibanDigits}` : "";
      if (value("password") !== value("password_confirmation"))
        fail("password_confirmation", "رمز عبور و تکرار آن یکسان نیستند.");
      if (!validNationalID(value("national_id")))
        fail("national_id", "کد ملی واردشده معتبر نیست. لطفاً کد ملی صحیح را وارد کنید.");
      if (bankIBAN && !validIranianIBAN(bankIBAN))
        fail("bank_iban", "شماره شبای واردشده معتبر نیست. هر ۲۴ رقم بعد از IR را کامل وارد کنید.");
      if (!/^09\d{9}$/.test(value("phone")))
        fail("phone", "شماره موبایل معتبر نیست. شماره باید با 09 آغاز شود و ۱۱ رقم داشته باشد.");
      if (role === "athlete" && !/^09\d{9}$/.test(value("emergency_phone")))
        fail("emergency_phone", "شماره تماس اضطراری معتبر نیست.");
      const coachingFiles = form
        .getAll("coaching_documents")
        .filter((item) => item instanceof File && item.size) as File[];
      if (role === "coach" && coachingFiles.length > 5)
        throw new Error("حداکثر ۵ فایل مدرک مربیگری قابل انتخاب است.");
      const body = {
        role,
        phone: value("phone"),
        password: value("password"),
        bank_iban: bankIBAN || null,
        first_name: value("first_name"),
        last_name: value("last_name"),
        father_name: optional("father_name"),
        national_id: value("national_id"),
        birth_date: value("birth_date"),
        province: optional("province"),
        city: optional("city"),
        address: optional("address"),
        education: value("education"),
        field_of_study: optional("field_of_study"),
        gender: value("gender"),
        marital_status: value("marital_status"),
        description: optional("description"),
        has_medical_history: role === "athlete" ? medical : null,
        medical_history: role === "athlete" ? optional("medical_history") : null,
        blood_type: role === "athlete" ? optional("blood_type") : null,
        emergency_phone: role === "athlete" ? value("emergency_phone") : null,
        emergency_relation: role === "athlete" ? value("emergency_relation") : null,
        sports_insurance_expires_at:
          role === "athlete" ? optional("sports_insurance_expires_at") : null,
      };
      const response = await fetch(`${API}/v1/auth/register`, {
          method: "POST",
          headers: { "Content-Type": "application/json", "Accept-Language": locale },
          body: JSON.stringify(body),
        }),
        result = await response.json();
      if (!response.ok) {
        if (result.field) {
          const field = formElement.elements.namedItem(result.field);
          if (field instanceof HTMLElement) {
            field.setAttribute("aria-invalid", "true");
            field.focus();
            field.scrollIntoView({ behavior: "smooth", block: "center" });
          }
        }
        throw new Error(result.message || "ثبت‌نام انجام نشد.");
      }
      const groups: [string, string][] = [
        ["profile_image", "profile"],
        ["coaching_documents", "coaching"],
        ["insurance_document", "insurance"],
        ["other_documents", "other"],
      ];
      for (const [name, kind] of groups)
        for (const item of form.getAll(name))
          if (item instanceof File && item.size) await upload(result.access_token, kind, item);
      localStorage.setItem("clubmax_access_token", result.access_token);
      localStorage.setItem("clubmax_user", JSON.stringify(result.user));
      router.replace(`/${locale}/dashboard`);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "خطا در ارتباط با سرور");
    } finally {
      setLoading(false);
    }
  }
  return (
    <main className="signup-page">
      <header>
        <SiteLogo href={`/${locale}`} priority />
        <span>
          قبلاً ثبت‌نام کرده‌اید؟ <Link href={`/${locale}/login`}>ورود</Link>
        </span>
      </header>
      <form className="signup-card" onSubmit={submit}>
        <div className="signup-title">
          <span className="eyebrow">ساخت حساب جدید</span>
          <h1>به جمع کلاب ماکس بپیوندید</h1>
          <p>اطلاعات ستاره‌دار برای ساخت حساب ضروری‌اند.</p>
        </div>
        <div className="role-cards">
          <button
            type="button"
            className={role === "coach" ? "active" : ""}
            onClick={() => setRole("coach")}
          >
            <i>م</i>
            <b>مربی هستم</b>
            <small>مدیریت باشگاه و ورزشکاران</small>
          </button>
          <button
            type="button"
            className={role === "athlete" ? "active" : ""}
            onClick={() => setRole("athlete")}
          >
            <i>و</i>
            <b>ورزشکار هستم</b>
            <small>عضویت و پیگیری فعالیت‌ها</small>
          </button>
        </div>
        {!role && (
          <p className="role-required-hint">انتخاب یکی از نقش‌های مربی یا ورزشکار الزامی است.</p>
        )}
        <h3>اطلاعات هویتی</h3>
        <div className="form-grid">
          <Field label="نام" name="first_name" required />
          <Field label="نام خانوادگی" name="last_name" required />
          <Field label="نام پدر" name="father_name" />
          <Field label="کد ملی" name="national_id" required>
            <input
              name="national_id"
              required
              inputMode="numeric"
              dir="ltr"
              minLength={10}
              maxLength={10}
              pattern="[0-9]{10}"
            />
          </Field>
          <JalaliDateInput name="birth_date" required minYear={1320} maxYear={1405} />
          <Field label="جنسیت" name="gender" required>
            <select name="gender" required defaultValue="">
              <option value="">انتخاب کنید</option>
              <option value="male">مرد</option>
              <option value="female">زن</option>
            </select>
          </Field>
          <Field label="تحصیلات" name="education" required>
            <select name="education" required defaultValue="">
              <option value="">انتخاب کنید</option>
              {education.map(([key, title]) => (
                <option key={key} value={key}>
                  {title}
                </option>
              ))}
            </select>
          </Field>
          <Field label="رشته تحصیلی" name="field_of_study" />
          <Field label="وضعیت تأهل" name="marital_status">
            <select name="marital_status" defaultValue="unknown">
              <option value="unknown">نامشخص</option>
              <option value="single">مجرد</option>
              <option value="married">متأهل</option>
            </select>
          </Field>
        </div>
        <h3>آدرس، تماس و اطلاعات بانکی</h3>
        <div className="form-grid">
          <LocationFields />
          <Field label="نشانی" name="address" />
          <Field label="شماره موبایل" name="phone" required>
            <input
              name="phone"
              required
              inputMode="numeric"
              dir="ltr"
              minLength={11}
              maxLength={11}
              pattern="09[0-9]{9}"
              placeholder="09123456789"
            />
          </Field>
          <Field label="شماره شبای بانکی ایران" name="bank_iban">
            <IranianIBANInput />
          </Field>
          <div className="signup-password-fields">
            <PasswordField name="password" label="رمز عبور" />
            <PasswordField name="password_confirmation" label="تکرار رمز عبور" />
          </div>
        </div>
        {role === "athlete" && (
          <>
            <h3>سلامت ورزشکار</h3>
            <div className="form-grid">
              <Field label="سابقه بیماری" name="has_medical_history" required>
                <select
                  value={String(medical)}
                  onChange={(e) => setMedical(e.target.value === "true")}
                >
                  <option value="false">خیر</option>
                  <option value="true">بله</option>
                </select>
              </Field>
              {medical && <Field label="شرح سابقه بیماری" name="medical_history" required />}
              <Field label="گروه خونی (اختیاری)" name="blood_type">
                <select name="blood_type" defaultValue="">
                  <option value="">انتخاب نشده</option>
                  {blood.map((x) => (
                    <option key={x}>{x}</option>
                  ))}
                </select>
              </Field>
              <Field label="شماره تماس اضطراری" name="emergency_phone" required>
                <input
                  name="emergency_phone"
                  required
                  inputMode="numeric"
                  dir="ltr"
                  maxLength={11}
                />
              </Field>
              <Field label="نسبت تماس اضطراری" name="emergency_relation" required>
                <select name="emergency_relation" required defaultValue="">
                  <option value="">انتخاب کنید</option>
                  {relations.map((x) => (
                    <option key={x}>{x}</option>
                  ))}
                </select>
              </Field>
              <JalaliDateInput
                name="sports_insurance_expires_at"
                label="تاریخ اعتبار بیمه ورزشی"
                minYear={1400}
                maxYear={1420}
              />
            </div>
          </>
        )}
        <h3>تصاویر و مدارک</h3>
        <div className="form-grid file-grid">
          <FileField
            name="profile_image"
            label="عکس پروفایل"
            maxFiles={1}
            accept="image/jpeg,image/png,image/webp"
          />
          {role === "coach" && (
            <FileField name="coaching_documents" label="مدارک مربیگری" maxFiles={5} required />
          )}
          {role === "athlete" && (
            <FileField name="insurance_document" label="تصویر کارت بیمه ورزشی" maxFiles={1} />
          )}
          <FileField name="other_documents" label="سایر مدارک" maxFiles={10} />
        </div>
        <Field label="توضیحات" name="description">
          <textarea name="description" rows={3} />
        </Field>
        <label className="accuracy-confirmation">
          <input type="checkbox" name="information_accuracy_confirmed" required />
          <span>
            اینجانب صحت و کامل بودن اطلاعات و مدارک واردشده را تأیید می‌کنم و مسئولیت هرگونه مغایرت،
            نقص یا ارائه اطلاعات نادرست را بر عهده می‌گیرم.
          </span>
        </label>
        {error && <div className="alert">{error}</div>}
        <button className="submit signup-submit" disabled={loading}>
          {loading ? "در حال ثبت اطلاعات..." : "ساخت حساب "}
        </button>
      </form>
    </main>
  );
}
