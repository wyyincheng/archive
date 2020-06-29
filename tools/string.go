package tools

import "regexp"

func init() {

}

// TrimBoth 使用正则去除两边空格
func TrimBoth(info string) string {
	reg := regexp.MustCompile(`[\w]+`)
	return reg.FindString(info)
}
