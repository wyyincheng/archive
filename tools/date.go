package tools

import "time"

func init() {

}

// CheckOutDay 检测时间是是否过期，与当前时间比对， ct：比对时间， day：间隔，单位天
func CheckOutDay(ct int64, day int) bool {
	tm := time.Unix(ct, 0)
	diff := time.Now().Sub(tm).Hours()
	return diff > float64(day*24)
}

// String Jan _2 15:04:05 2006
func String(ct int64) string {
	tm := time.Unix(ct, 0)
	return tm.Format("Jan _2 15:04:05 2006")
}
