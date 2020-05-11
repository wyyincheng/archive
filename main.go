package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var app = cli.NewApp()

var archiveInfo = Archive{
	CLI:     app.Version,
	Version: "v9.7.0",
	Time:    time.Now().Unix(),
	Status:  0,
}

func main() {

	buildCLI()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func buildCLI() {
	app.Name = "archive"
	app.Usage = "archive appstore latest version which has been published."
	app.Action = func(c *cli.Context) error {
		// fmt.Println("start archive")
		// fmt.Println("into:", c.String("into"))
		// fmt.Println("version:", c.String("v"))
		// fmt.Println("branch", c.String("b"))
		target := c.String("into")
		version := c.String("v")
		archive(target, version)
		return nil
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "project",
			Aliases: []string{"p"},
			Value:   "niuwa-ios",
			Usage:   "Project you will archive.",
		},
		&cli.StringFlag{
			Name:    "into",
			Value:   "master",
			Aliases: []string{"i"},
			Usage:   "archive version code into which branch.",
		},
		&cli.StringFlag{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "project version you will archive.",
		},
		&cli.StringFlag{
			Name:    "branch",
			Aliases: []string{"b"},
			Usage:   "project branch you will archive.",
		},
	}
	app.Commands = []*cli.Command{
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
	}

	// sort.Sort(cli.FlagsByName(app.Flags))
	// sort.Sort(cli.CommandsByName(app.Commands))

}

func archive(target string, version string) {
	/**
	1.检测命令
	2.同步代码
	3.切换分支


	*/
	checkCMD("git")
	archiveInfo.User = gitConfig("user.name")
	archiveInfo.Email = gitConfig("user.email")
	sync()
	merge("master", version)
}

func sync() {
	success, fetchResult := excute("git fetch")
	if success {

		println(fetchResult)
	}
}

func merge(target string, version string) {
	/**
	1.分支检测，target、from
	2.分支切换 target
	3.代码同步
	4.记录并merge
	5.同步
	*/
	success, branch := search(version)
	if success {

		archiveInfo.Version = version
		archiveInfo.Branch = branch
		// -f use checkout -f
		// ohter checkout "Your branch is up to date"

		excute("git checkout -f")
		excute("git checkout " + target)
		excute("git pull")
		archiveInfo.Commit = fetchLatestCommit("branch", target)
		branchInfo := Branch{
			Name:   branch,
			Commit: fetchLatestCommit("branch", branch),
		}
		mergeSuccess, _ := excute("git merge --no-ff " + branch)
		if mergeSuccess {
			excute("git push")
			fmt.Println(branchInfo)
			fmt.Println(fetchLatestCommit("branch", branch))
			// archiveInfo.branches = []Branch{
			// 	&branchInfo
			// }
		} else {
			print("auto merge faulure. ")
			// abort(mergeResult)
		}
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

func gitConfig(key string) string {
	_, config := excute("git config --get " + key)
	return config
}

func fetchLatestCommit(sort string, info string) string {
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

func write(info Archive) {
	infoJSON, _ := json.Marshal(info)
	if infoJSON != nil {
		pwd, _ := os.Getwd()
		filePath := path.Join(pwd, "backup")
		os.MkdirAll(filePath, os.ModePerm)
		ioutil.WriteFile(path.Join(filePath, info.Version+".json"), infoJSON, os.ModePerm)
	}
}

func test() {
	info := Archive{
		CLI:     "1.0.0",
		Version: "v9.7.0",
		User:    "yc",
		Time:    time.Now().Unix(),
		Status:  0,
	}

	infoJSON, _ := json.Marshal(info)
	if infoJSON != nil {
		pwd, _ := os.Getwd()
		filePath := path.Join(pwd, "backup")
		os.MkdirAll(filePath, os.ModePerm)
		ioutil.WriteFile(path.Join(filePath, info.Version+".json"), infoJSON, os.ModePerm)
	}
}

//Archive 归档信息
type Archive struct {
	CLI      string
	Version  string
	Branch   string
	Commit   string
	User     string
	Email    string
	branches []Branch
	tags     []Tag
	Time     int64
	Status   int //0 默认状态，1 已还原，必要时可被删除
}

//Branch 分支信息
type Branch struct {
	Name   string
	Commit string
}

//Tag tag信息
type Tag struct {
	Name   string
	Commit string
}
