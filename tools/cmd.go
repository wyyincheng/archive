package tools

import (
	"bytes"
	"os/exec"
	"strings"
)

func init() {

}

// Excute 执行cmd指令并返回执行状态和结果信息
func Excute(cmd string) (bool, string) {
	// fmt.Printf("cmd run: '%s'\n", cmdStr)
	// logger.Println(cmdStr)
	branches := strings.Split(cmd, " ")
	CMD := exec.Command(branches[0], branches[1:]...)
	// cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	CMD.Stdout = &stdout
	CMD.Stderr = &stderr
	err := CMD.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		// logger.Println(errStr)
		// if silent == false {
		// 	saveArchive()
		// 	//静默处理：正常返回处理结果，不结束程序
		// 	logger.Fatal(err)
		// }
		return false, errStr
	}
	// logger.Println(outStr)
	//TODO: log 、 notification
	return true, outStr
}
