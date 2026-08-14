package service

import "time"

// persianYearMonth converts a Gregorian date to its Iranian Solar Hijri year/month.
func persianYearMonth(value time.Time) (int, int) {
	location, err := time.LoadLocation("Asia/Tehran")
	if err == nil {
		value = value.In(location)
	}
	gy, gm, gd := value.Date()
	gdm := [...]int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	jy := 0
	if gy > 1600 {
		jy = 979
		gy -= 1600
	} else {
		gy -= 621
	}
	gy2 := gy
	if gm > 2 {
		gy2++
	}
	days := 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 - 80 + gd + gdm[gm-1]
	jy += 33 * (days / 12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	if days < 186 {
		return jy, 1 + days/31
	}
	return jy, 7 + (days-186)/30
}
