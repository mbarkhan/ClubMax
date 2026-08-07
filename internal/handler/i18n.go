package handler

import (
	"net/http"
	"strings"
)

var translations = map[string]map[string]string{
	"fa": {
		"verification_code_sent": "کد تأیید ارسال شد.",
		"phone_updated":          "شماره موبایل با موفقیت تغییر کرد.",
		"athlete_joined_gym":     "ورزشکار با موفقیت در باشگاه ثبت‌نام شد.",
		"internal_error":         "خطایی در سرور رخ داده است.",
		"invalid_input":          "اطلاعات واردشده معتبر نیست.",
		"invalid_otp":            "کد تأیید نامعتبر یا منقضی شده است.",
		"forbidden":              "اجازه انجام این عملیات را ندارید.",
		"not_found":              "مورد درخواستی پیدا نشد.",
		"conflict":               "این اطلاعات قبلاً ثبت شده است.",
		"gym_limit_reached":      "تعداد باشگاه‌های قابل ثبت به حد مجاز رسیده است.",
		"unauthorized":           "برای ادامه باید وارد حساب کاربری شوید.",
	},
	"en": {
		"verification_code_sent": "Verification code sent.",
		"phone_updated":          "Phone number updated successfully.",
		"athlete_joined_gym":     "Athlete joined the gym successfully.",
		"internal_error":         "An internal server error occurred.",
		"invalid_input":          "The supplied data is invalid.",
		"invalid_otp":            "The verification code is invalid or expired.",
		"forbidden":              "You are not allowed to perform this action.",
		"not_found":              "The requested resource was not found.",
		"conflict":               "The resource already exists.",
		"gym_limit_reached":      "The gym registration limit has been reached.",
		"unauthorized":           "Authentication is required.",
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
