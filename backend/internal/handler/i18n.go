package handler

import (
	"net/http"
	"strings"
)

var translations = map[string]map[string]string{
	"fa": {
		"verification_code_sent":   "کد تأیید ارسال شد.",
		"phone_updated":            "شماره موبایل با موفقیت تغییر کرد.",
		"athlete_joined_gym":       "ورزشکار با موفقیت در باشگاه ثبت‌نام شد.",
		"internal_error":           "خطایی در سرور رخ داده است.",
		"invalid_input":            "اطلاعات واردشده معتبر نیست.",
		"invalid_otp":              "کد تأیید نامعتبر یا منقضی شده است.",
		"forbidden":                "اجازه انجام این عملیات را ندارید.",
		"invalid_credentials":      "نام کاربری یا رمز عبور نادرست است.",
		"protected_admin":          "مدیر پیش‌فرض با نام کاربری admin قابل حذف نیست.",
		"not_found":                "مورد درخواستی پیدا نشد.",
		"conflict":                 "این اطلاعات قبلاً ثبت شده است.",
		"gym_limit_reached":        "تعداد باشگاه‌های قابل ثبت به حد مجاز رسیده است.",
		"gym_image_limit_reached":  "حداکثر سه تصویر برای هر باشگاه قابل ثبت است.",
		"gym_full":                 "ظرفیت این باشگاه تکمیل شده است.",
		"unauthorized":             "برای ادامه باید وارد حساب کاربری شوید.",
		"insufficient_funds":       "موجودی کیف پول برای این برداشت کافی نیست.",
		"withdrawal_limit":         "مبلغ برداشت خارج از حدود تعیین‌شده است.",
		"invalid_state":            "وضعیت این درخواست اجازه این عملیات را نمی‌دهد.",
		"message_recipient":        "اجازه ارسال پیام به این گیرنده را ندارید.",
		"backup_in_progress":       "یک عملیات پشتیبان‌گیری دیگر در حال اجرا است.",
		"invalid_backup":           "فایل پشتیبان معتبر یا سازگار با این نصب کلاب‌ماکس نیست.",
		"backup_restored":          "پایگاه داده و فایل‌های آپلودشده با موفقیت بازیابی شدند.",
		"backup_database_restored": "پایگاه داده بازیابی شد؛ فایل‌های آپلودشده فعلی بدون تغییر باقی ماندند.",
	},
	"en": {
		"verification_code_sent":   "Verification code sent.",
		"phone_updated":            "Phone number updated successfully.",
		"athlete_joined_gym":       "Athlete joined the gym successfully.",
		"internal_error":           "An internal server error occurred.",
		"invalid_input":            "The supplied data is invalid.",
		"invalid_otp":              "The verification code is invalid or expired.",
		"forbidden":                "You are not allowed to perform this action.",
		"invalid_credentials":      "The username or password is incorrect.",
		"protected_admin":          "The default admin account cannot be deleted.",
		"not_found":                "The requested resource was not found.",
		"conflict":                 "The resource already exists.",
		"gym_limit_reached":        "The gym registration limit has been reached.",
		"gym_image_limit_reached":  "A gym can have at most three images.",
		"gym_full":                 "This gym has reached its capacity.",
		"unauthorized":             "Authentication is required.",
		"insufficient_funds":       "Wallet balance is insufficient for this withdrawal.",
		"withdrawal_limit":         "The withdrawal amount is outside the configured limits.",
		"invalid_state":            "This request cannot be changed from its current state.",
		"message_recipient":        "You are not allowed to message this recipient.",
		"backup_in_progress":       "Another backup operation is already running.",
		"invalid_backup":           "The backup is invalid or incompatible with this ClubMax installation.",
		"backup_restored":          "The database and uploaded files were restored successfully.",
		"backup_database_restored": "The database was restored; current uploaded files were left unchanged.",
	},
}

func locale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "fa"
		if strings.HasPrefix(strings.ToLower(r.Header.Get("Accept-Language")), "en") {
			lang = "en"
		}
		w.Header().Set("Content-Language", lang)
		w.Header().Add("Vary", "Accept-Language")
		next.ServeHTTP(w, r)
	})
}

func message(w http.ResponseWriter, key string) string {
	lang := w.Header().Get("Content-Language")
	if text, ok := translations[lang][key]; ok {
		return text
	}
	return translations["fa"][key]
}
