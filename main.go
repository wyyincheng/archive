package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "archive",
		Usage: "archive appstore latest version which has been published.",
		Action: func(c *cli.Context) error {
			fmt.Println("start archive")
			archive(os.Args[1])
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "project,p",
				Value: "niuwa-ios",
				Usage: "Project you will archive.",
			},
			&cli.StringFlag{
				Name:  "version,v",
				Usage: "project version you will archive.",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "lock",
				Usage: "lock tag or branch",
				Action: func(c *cli.Context) error {
					fmt.Println("lock tag or branch")
					return nil
				},
			},
			{
				Name:  "clean",
				Usage: "clean tags and branches after archive",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "abort",
				Usage: "rollback archive which version you given",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "test",
				Usage: "test cmd",
				Action: func(c *cli.Context) error {
					test()
					return nil
				},
			},
		},
	}

	// sort.Sort(cli.FlagsByName(app.Flags))
	// sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func archive(version string) {
	/**
	1.检测命令
	2.同步代码
	3.切换分支


	*/
	checkCMD("git")
	sync()
	merge("develop", version)
}

func sync() {
	success, fetchResult := excute("git fetch")
	if success {

		println(fetchResult)
	}
}

func merge(target string, from string) {
	/**
	1.分支检测，target、from
	2.分支切换 target
	3.代码同步
	4.记录并merge
	5.同步
	*/
	success, branch := search(from)
	if success {

		// -f use checkout -f
		// ohter checkout "Your branch is up to date"

		excute("git checkout -f")
		excute("git checkout " + target)
		excute("git pull")
		beforCommit1 := backup("branch", branch)
		beforCommit2 := backup("branch", "master")
		mergeSuccess, _ := excute("git merge --no-ff " + branch)
		afterCommit3 := backup("branch", branch)
		afterCommit4 := backup("branch", "master")
		if mergeSuccess {
			excute("git push")
		} else {
			print("auto merge faulure. ")
			// abort(mergeResult)
		}
		fmt.Printf("merge result: '%s'\n'%s'\n'%s'\n'%s'\n", beforCommit1, beforCommit2, afterCommit3, afterCommit4)
	}
}

func search(branch string) (bool, string) {
	success, searchResult := excute("git branch -r")
	if success == false {
		return false, ""
	}
	// result := excute("git", "branch", "-r", "|", "grep", branch)
	branches := strings.Split(searchResult, "\n")
	for _, info := range branches {
		result := strings.Replace(info, " ", "", -1)
		if strings.HasSuffix(result, branch) {
			return true, result
		}
	}
	//deal error
	return false, ""
}

func backup(sort string, info string) string {
	if sort == "branch" {
		success, result := excute("git branch -r -v")
		if success {
			commitInfos := strings.Split(result, "\n")
			for _, commit := range commitInfos {
				trimStr := strings.Trim(commit, " ")
				if strings.HasPrefix(trimStr, info+" ") {
					infos := strings.Replace(trimStr, info+" ", "", 1)
					return strings.Split(infos, " ")[0]
				}
			}
		}
	} else if sort == "tag" {

	}
	return ""
}

func abort() {

}

func clean() {

}

func lock() {

}

//private methods

//tools
func checkCMD(cmd string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		fmt.Printf("didn't find '%s' cmd", cmd)
		os.Exit(404)
	}
}

func excute(cmdStr string) (bool, string) {
	fmt.Printf("cmd run: '%s'\n", cmdStr)
	branches := strings.Split(cmdStr, " ")
	cmd := exec.Command(branches[0], branches[1:]...)
	// cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		fmt.Printf("'%s' failed : '%s'", cmdStr, errStr)
		log.Fatalf("cmd('%s').Run() failed with %s\n", cmdStr, err)
		// os.Exit(600)
		return false, errStr
	}
	// fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	//TODO: log 、 notification
	return true, outStr
}

func test() {
	// cmd := exec.Command("git", "clone", "https://github.com/windzhu0514/cmd", "/Users/yc/Develop/Golang/GoShell/archive")
	cmd := exec.Command("git", "fetch")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
}
