package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	appVersion  = "v0.0.2"
	configPath  = "/usr/local/share/YCLI/Archive"
	app         = cli.NewApp()
	config      = Config{}
	archiveInfo = Archive{
		Time:   time.Now().Unix(),
		Status: 0,
	}
	archivePath string
	logger      *log.Logger
	logPath     string
)

func main() {
	loadConfig()
	buildLogger()
	buildCLI()

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}
}

func buildCLI() {
	app.Name = "archive"
	app.Usage = "archive appstore latest version which has been published."
	app.Action = func(c *cli.Context) error {

		if c.Bool("V") {
			fmt.Println(appVersion)
			return nil
		}

		// fmt.Println("start archive")
		// fmt.Println("into:", c.String("into"))
		// fmt.Println("version:", c.String("v"))
		// fmt.Println("branch", c.String("b"))
		target := c.String("into")
		tag := c.String("t")
		archive(target, tag)
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
		&cli.BoolFlag{
			Name:    "Version",
			Aliases: []string{"V"},
			Value:   false,
			Usage:   "show cli version.",
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
			Name:  "config",
			Usage: "setting archive config",
			Action: func(c *cli.Context) error {
				key := c.String("get")
				value := getConfig(key)
				fmt.Printf("Archive Config ('%s' : '%s')", key, value)
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "get",
					Usage: "Get config value for the key.",
				},
				&cli.StringFlag{
					Name:  "get-all",
					Usage: "Get all config value.",
				},
				&cli.StringFlag{
					Name:  "add",
					Usage: "Add config key and value.",
				},
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

func buildLogger() {
	logPath = path.Join(config.WorkSpace, "Logs", strconv.FormatInt(time.Now().Unix(), 10)+".log")
	dirPath := path.Dir(logPath)
	mkErr := os.MkdirAll(dirPath, os.ModePerm)
	if mkErr != nil {
		fmt.Printf("mkdir log folder err : '%s'\n", mkErr)
		return
	}
	logfile, err := os.Create(logPath)
	if err != nil {
		fmt.Printf("create log file err : '%s'\n", mkErr)
		return
	}

	logger = log.New(logfile, "", log.LstdFlags|log.Llongfile)
}

func archive(target string, vtag string) {

	/**
	1.检测命令
	2.同步代码
	3.切换分支


	*/
	checkCMD("git")
	checkVersion(vtag)
	archiveInfo.User = strings.Trim(gitConfig("user.name"), "\n")
	archiveInfo.Email = strings.Trim(gitConfig("user.email"), "\n")
	success := merge(target, vtag)
	if success {
		archiveInfo.Log = logPath
		cleanBranch(All)
		saveArchive(archiveInfo)
		fmt.Printf("Archive '%s' into '%s' success, see more info on:\nlog: '%s'\ninfo: '%s'\n", vtag, target, logPath, archivePath)
		updateVersion()
		return
	}
	fmt.Printf("Archive '%s' into '%s' failure, see more info on:\nlog: '%s'\ninfo: '%s'\n", vtag, target, logPath, archivePath)
	updateVersion()
}

func checkVersion(vtag string) (bool, string, string) {
	//tag 可用
	//branch 存在

	var branch, tag string
	var success bool

	return success, branch, tag
}

func merge(target string, vtag string) bool {
	/**
	1.分支检测，target、from
	2.分支切换 target
	3.代码同步
	4.记录并merge
	5.同步
	*/
	success, branch := search(vtag)
	if success {

		archiveInfo.Tag = vtag
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

		if success {
			commitInfos := strings.Split(result, "\n")
			for _, commit := range commitInfos {
				trimStr := strings.Trim(strings.Trim(commit, "*"), " ")
				if strings.HasPrefix(trimStr, info) {
					infos := strings.Replace(trimStr, info+" ", "", 1)
					cmt := strings.Split(infos, " ")[0]
					logger.Printf("'%s' '%s' '%s' fetch latest commit : '%s' \n", sort, info, tracking, cmt)
					return cmt
				}
			}
		}
	} else if sort == "tag" {

	}
	logger.Printf("'%s' '%s' '%s' fetch latest commit failure \n", sort, info, tracking)
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
}

func lock() {

}

//private methods

//tools
func checkCMD(cmd string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		logger.Fatal(err)
	}
}

func excute(cmdStr string, silent bool) (bool, string) {
	// fmt.Printf("cmd run: '%s'\n", cmdStr)
	logger.Println(cmdStr)
	branches := strings.Split(cmdStr, " ")
	cmd := exec.Command(branches[0], branches[1:]...)
	// cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		logger.Println(errStr)
		if silent == false {
			//静默处理：正常返回处理结果，不结束程序
			log.Fatal(err)
		}
		return false, errStr
	}
	logger.Println(outStr)
	//TODO: log 、 notification
	return true, outStr
}

func saveArchive(info Archive) {
	infoJSON, _ := json.Marshal(info)
	archivePath = path.Join(config.WorkSpace, "backup", info.Tag+".json")
	write(infoJSON, archivePath)
}

func saveConfig(config Config) {
	infoJSON, _ := json.Marshal(config)
	filePath := path.Join(config.WorkSpace, "Config.json")
	write(infoJSON, filePath)
}

func write(json []byte, filePath string) {
	if json != nil {
		dirPath := path.Dir(filePath)
		mkErr := os.MkdirAll(dirPath, os.ModePerm)
		if mkErr != nil {
			logger.Fatal(mkErr)
			return
		}
		writeErr := ioutil.WriteFile(filePath, json, os.ModePerm)
		if writeErr != nil {
			logger.Fatal(writeErr)
			return
		}
	}
}

func getConfig(key string) string {
	loadConfig()
	return config.WorkSpace
}

func updateArchiveInfo() {
	archiveInfo.ENV = config
}

func loadConfig() {
	defer updateArchiveInfo()
	//checkFile
	configFile := path.Join(configPath, "Config.json")
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		//初始化
		logger.Printf("'%s' no exist.\n", configFile)
		config.WorkSpace = configPath
		config.LatestCheck = time.Now()
		config.UpdateVersion = Day
		config.Version = appVersion
		saveConfig(config)
		logger.Printf("Default archive config constructor success! You can update it on path '%s'\n", configFile)
		return
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		//Log load config failure
		logger.Fatal(err)
	}
	var localConfig Config
	err = json.Unmarshal(data, &localConfig)
	if err != nil {
		//Log load config failure
		logger.Fatal(err)
	}
	config = localConfig
}

func localTime() string {
	local, err := time.LoadLocation("Local")
	if err != nil {
		logger.Printf("format lcoal time failure: '%s'", err)
		return time.Now().String()
	}
	return time.Now().In(local).Format("202005010-15:04:05")
}

func updateVersion() {

	diff := time.Now().Sub(config.LatestCheck).Hours()
	needCheck := false
	switch config.UpdateVersion {
	case Day:
		needCheck = diff > 24
	case Week:
		needCheck = diff > 7*24
	case Month:
		needCheck = diff > 30*24
	default:
		needCheck = true
	}

	config.LatestCheck = time.Now()
	if needCheck == false {
		return
	}

	resp, err := http.Get("https://api.github.com/repos/wyyincheng/archive/releases/latest")
	if err != nil {
		logger.Printf("check cli version failure(0): '%s'", err)
	} else {
		var data map[string]interface{}
		jsonErr := json.NewDecoder(resp.Body).Decode(&data)
		if jsonErr != nil {
			logger.Printf("check cli version failure(1): '%s'", jsonErr)
			// resp.Body.Close()
			return
		}

		needUpdate := false
		latestList := strings.Split(strings.Replace(data["tag_name"].(string), "v", "", 1), ".")
		currentList := strings.Split(strings.Replace(appVersion, "v", "", 1), ".")
		for i := 0; i < len(latestList); i++ {
			lv, lerr := strconv.Atoi(latestList[i])
			cv, cerr := strconv.Atoi(currentList[i])
			if lerr == nil && cerr == nil {
				if lv > cv {
					needUpdate = true
					break
				}
			}
		}
		if needUpdate {
			fmt.Println("\n#######################################################################")
			fmt.Printf("# archive '%s' is available. You are on '%s'.\n", data["tag_name"], appVersion)
			fmt.Printf("# You should use the latest version.\n")
			fmt.Printf("# Please update using `brew upgrade wyyincheng/tap/archive`.\n")
			fmt.Println("#######################################################################")
		}
	}
	// resp.Body.Close()
}

func test(target string, vtag string) {
	updateVersion()
}

// String value for traking
func String(traking Tracking) string {
	switch traking {
	case All:
		return "All"
	case Local:
		return "Local"
	case Remote:
		return "Remote"
	default:
		return "Unkonw"
	}
}

//Config 配置信息
type Config struct {
	Version       string
	WorkSpace     string
	DefaultBranch []string
	UpdateVersion Frequency
	LatestCheck   time.Time
}

//Archive 归档信息
type Archive struct {
	ENV      Config
	Tag      string
	Branch   string
	Commit   string
	User     string
	Email    string
	branches []Branch
	tags     []Tag
	Time     int64
	Status   int //0 默认状态，1 已还原，必要时可被删除
	Log      string
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
type Tracking string

// tracking type
const (
	All    Tracking = "All"
	Local  Tracking = "Local"
	Remote Tracking = "Remote"
)

// State branch state
type State int

// 分支状态
const (
	Merged State = iota
	Delete
	Abort
)

// Frequency check update frequency
type Frequency int

// 版本更新检测频率
const (
	Day Frequency = iota
	Week
	Month
)
