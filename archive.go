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

var config = Config{}
var archiveInfo = Archive{
	CLI:    app.Version,
	Time:   time.Now().Unix(),
	Status: 0,
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
			Subcommands: []*cli.Command{
				{
					Name:  "branch",
					Usage: "clean branches which been merged",
					Action: func(c *cli.Context) error {

						return nil
					},
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "all",
							Aliases: []string{"a"},
							Value:   true,
							Usage:   "clean all branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "remote",
							Aliases: []string{"r"},
							Value:   true,
							Usage:   "clean remote branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "local",
							Aliases: []string{"l"},
							Value:   true,
							Usage:   "clean local branches which been merged",
						},
					},
				},
				{
					Name:  "tag",
					Usage: "clean tag which out of range rule",
					Action: func(c *cli.Context) error {

						return nil
					},
				},
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
	checkVersion(version)
	archiveInfo.User = gitConfig("user.name")
	archiveInfo.Email = gitConfig("user.email")
	success := merge(target, version)
	if success {
		cleanBranch(All)
	}
}

func checkVersion(version string) (bool, string, string) {
	//tag 可用
	//branch 存在

	var branch, tag string
	var success bool

	return success, branch, tag
}

func merge(target string, version string) bool {
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
		archiveInfo.Commit = fetchLatestCommit("branch", target, Local)
		mergeSuccess, _ := excute("git merge --no-ff "+branch, true)
		if mergeSuccess {
			excute("git push", false)
			archiveInfo.branches = []Branch{
				{
					Name:     branch,
					Tracking: Remote,
					State:    Merged,
					Commit:   fetchLatestCommit("branch", branch, Remote),
				},
			}
			saveArchive(archiveInfo)
			return true
		}
		abort("merge", "")
	}
	return false
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

func fetchLatestCommit(sort string, info string, tracking Tracking) string {
	if sort == "branch" {
		fmt.Println("fetch latest commit branch: " + info)

		var success bool
		var result string

		if tracking == Remote {
			status, resp := excute("git branch -r -v", false)
			success = status
			result = resp
		} else if tracking == Local {
			status, resp := excute("git branch -v", false)
			success = status
			result = resp
		}

		fmt.Println("fetch latest commit result : " + result)
		if success {
			commitInfos := strings.Split(result, "\n")
			for _, commit := range commitInfos {
				trimStr := strings.Trim(strings.Trim(commit, "*"), " ")
				fmt.Println("fetch latest commit trimStr : " + trimStr)
				fmt.Println("fetch latest commit info : " + trimStr)
				if strings.HasPrefix(trimStr, info) {
					fmt.Println("fetch latest commit HasPrefix : " + info)
					infos := strings.Replace(trimStr, info+" ", "", 1)
					cmt := strings.Split(infos, " ")[0]
					return cmt
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

func cleanBranch(traking Tracking) {
	//指定分支，所有分支，本地分支，远程分支

	// excute("git checkout -f", false)
	// excute("git checkout master", false)

	var result string

	if traking == All {
		cleanBranch(Local)
		cleanBranch(Remote)
		return
	} else if traking == Local {
		_, resp := excute("git branch --merged", false)
		result = resp
	} else if traking == Remote {
		_, resp := excute("git branch -r --merged", false)
		result = resp
	}

	resultArray := strings.Split(result, "\n")
	branches := archiveInfo.branches
	for _, info := range resultArray {
		trimStr := strings.Trim(info, " ")
		branchInfo := strings.Replace(trimStr, "*", "", -1)
		branch := strings.Trim(branchInfo, " ")
		if branch == "master" || branch == "origin/master" || len(branch) == 0 {
			continue
		}
		// if config.DefaultBranch.contains(branch) {
		// continue
		// }
		branches = append(branches, Branch{
			Name:     branch,
			Tracking: traking,
			State:    Delete,
			Commit:   fetchLatestCommit("branch", branch, traking),
		})
	}
	archiveInfo.branches = branches
	fmt.Println(archiveInfo)
	fmt.Println("cleanBranch")
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

func saveArchive(info Archive) {
	infoJSON, _ := json.Marshal(info)
	filePath := path.Join(config.WorkSpace, "backup", info.Version+".json")
	write(infoJSON, filePath)
}

func write(json []byte, filePath string) {
	if json != nil {
		os.MkdirAll(filePath, os.ModePerm)
		ioutil.WriteFile(filePath, json, os.ModePerm)
	}
}

func readConfig() {

}

func test(target string, version string) {
	cleanBranch(All)
}

//Config 配置信息
type Config struct {
	Name          string
	Email         string
	WorkSpace     string
	DefaultBranch []string
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
	Name     string
	Commit   string
	Tracking Tracking
	State    State
}

//Tag tag信息
type Tag struct {
	Name   string
	Commit string
}

// Tracking type
type Tracking int

// tracking type
const (
	All Tracking = iota
	Local
	Remote
)

// Branch state
type State int

// 分支状态
const (
	Merged State = iota
	Delete
	Abort
)
