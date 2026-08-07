# GyMax
بک‌اند سامانه مدیریت باشگاه GyMax با Go و PostgreSQL. این نسخه شامل مدیریت کاربران و نقش‌ها، ثبت باشگاه، عضویت ورزشکار و زیرساخت تأیید شماره موبایل است.

GyMax is a gym-management platform built for coaches and athletes.

## ساختار پروژه

```text
GyMax/
├── backend/       API، منطق دامنه، repository و migrationهای PostgreSQL
├── frontend/      رابط کاربری Next.js، فارسی و انگلیسی
├── compose.yaml   پایگاه‌داده محلی
└── Makefile       فرمان‌های متداول توسعه
```

## اجرای محلی

1. پایگاه‌داده را با `docker compose up -d db` اجرا کنید.
2. در `backend/` فایل `.env.example` را به `.env` تبدیل و migrationها را اجرا کنید.
3. بک‌اند را با `cd backend && go run ./cmd/api` اجرا کنید.
4. فرانت‌اند را با `cd frontend && npm install && npm run dev` اجرا کنید.

سرویس به‌صورت پیش‌فرض فقط روی `http://127.0.0.1:8080` اجرا می‌شود. پاسخ‌های API فارسی هستند؛ برای دریافت پاسخ انگلیسی هدر `Accept-Language: en` ارسال کنید. زبان‌های پشتیبانی‌شده از مسیر `GET /v1/meta/locales` قابل دریافت‌اند.

رابط Next.js روی `http://127.0.0.1:3000/fa` اجرا می‌شود. مسیر انگلیسی `/en` است و ورود به `/` به فارسی هدایت می‌شود.

در محیط توسعه، کد OTP به‌جای ارسال SMS در log چاپ می‌شود. برای محیط production باید پیاده‌سازی `OTPNotifier` با درگاه پیامک جایگزین شود و `APP_ENV=production` باشد.

مدیر اولیه با متغیرهای `ADMIN_NAME`، `ADMIN_USERNAME` و `ADMIN_PASSWORD` هنگام شروع برنامه ایجاد می‌شود. رمز پیش‌فرض عمداً وجود ندارد.
