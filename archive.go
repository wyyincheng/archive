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
		test(target, version)
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
				target := c.String("into")
				version := c.String("v")
				test(target, version)
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
	merge(target, version)
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

		excute("git checkout -f", false)
		excute("git fetch", false)
		excute("git checkout "+target, false)
		excute("git pull", false)
		archiveInfo.Commit = fetchLatestCommit("branch", target)
		mergeSuccess, _ := excute("git merge --no-ff "+branch, true)
		if mergeSuccess {
			excute("git push", false)
			archiveInfo.branches = []Branch{
				{
					Name:   branch,
					Commit: fetchLatestCommit("branch", branch),
				},
			}
			write(archiveInfo)
		} else {
			abort("merge", "")
		}
	}
}

func search(branch string) (bool, string) {
	success, searchResult := excute("git branch -r", false)
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
	_, config := excute("git config --get "+key, false)
	return config
}

func fetchLatestCommit(sort string, info string) string {
	if sort == "branch" {
		fmt.Println("fetch latest commit branch: " + info)
		success, result := excute("git branch -r -v", false)
		fmt.Println("fetch latest commit result : " + result)
		if success {
			commitInfos := strings.Split(result, "\n")
			for _, commit := range commitInfos {
				trimStr := strings.Trim(commit, " ")
				fmt.Println("fetch latest commit trimStr : " + trimStr)
				fmt.Println("fetch latest commit info : " + trimStr)
				if strings.HasPrefix(trimStr, info) {
					fmt.Println("fetch latest commit HasPrefix : " + info)
					infos := strings.Replace(trimStr, info+" ", "", 1)
					return strings.Split(infos, " ")[0]
				}
			}
		}
	} else if sort == "tag" {

	}
	fmt.Println("fetch latest commit failure")
	return ""
}

func abort(action string, commit string) {
	if action == "branch" {

	} else if action == "tag" {

	} else if action == "merge" {
		excute("git merge --abort", false)
	}
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

func excute(cmdStr string, silent bool) (bool, string) {
	// fmt.Printf("cmd run: '%s'\n", cmdStr)
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
		if silent == false {
			//静默处理：正常返回处理结果，不结束程序
			log.Fatalf("cmd('%s').Run() failed with %s\n", cmdStr, err)
		}
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

func test(target string, version string) {
	_, branch := search(version)
	commit := fetchLatestCommit("branch", branch)
	fmt.Println("commit: " + commit)
	archiveInfo.branches = []Branch{
		{
			Name:   branch,
			Commit: commit,
		},
	}
	fmt.Println(archiveInfo)
	write(archiveInfo)
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
