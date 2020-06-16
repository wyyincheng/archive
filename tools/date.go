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
func String(unix int64) string {
	tm := time.Unix(unix, 0)
	return tm.Format("2006-01-02 15:04:05")
}
