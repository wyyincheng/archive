package main

import (
	"archive/git"
	"archive/tools"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	appVersion  = "v0.0.12"
	configPath  = "/usr/local/share/YCLI/Archive"
	app         = cli.NewApp()
	config      = Config{}
	archiveInfo = Archive{
		Time:   time.Now().Unix(),
		Status: 0,
	}
	archivePath string
	logPath     string
)

func main() {
	loadConfig()
	buildCLI()

	err := app.Run(os.Args)
	if err != nil {
		saveArchive()
		fmt.Printf("App Error:%s\n", err)
	}
}

func buildCLI() {
	app.Name = "nwarchive"
	app.Usage = "archive appstore latest version which has been published."
	app.Action = func(c *cli.Context) error {

		//TODO: log cmd str

		if c.Bool("V") {
			fmt.Println(appVersion)
			return nil
		}

		target := c.String("into")
		vtag := c.String("t")
		if len(vtag) > 0 {
			if checkTagLegal(vtag) {
				readyArchive("archive_tag")
				archive(target, vtag)
				saveArchive()
				return nil
			}
			fmt.Printf("%s is not legal, check and input like: v1.0.0\n", vtag)
			return nil
		}

		if checkTagLegal(c.Args().First()) {
			readyArchive("archive_version")
			archive(target, c.Args().First())
			saveArchive()
			return nil
		}

		fmt.Println("Incorrect Usage. Shoe help :\n  archive -h")
		return nil
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "project",
			Aliases: []string{"p"},
			Value:   "niuwa-ios",
			Usage:   "Project you will archive.",
		},
		// &cli.StringFlag{
		// 	Name:    "into",
		// 	Value:   "master",
		// 	Aliases: []string{"i"},
		// 	Usage:   "archive version code into which branch.",
		// },
		&cli.StringFlag{
			Name:    "tag",
			Aliases: []string{"t"},
			Usage:   "project version tag you will archive.",
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
		// {
		// 	Name:  "lock",
		// 	Usage: "lock tag or branch",
		// 	Action: func(c *cli.Context) error {
		// 		fmt.Println("lock tag or branch")
		// 		return nil
		// 	},
		// },
		{
			Name:  "clean",
			Usage: "clean tags and branches after archive",
			Action: func(c *cli.Context) error {
				ignore := c.String("i")
				readyArchive("clean_all")

				excute("git fetch", false)

				var tracking git.Tracking = git.All
				if c.Bool("r") {
					tracking = git.Remote
				} else if c.Bool("l") {
					tracking = git.Local
				}

				clean := !c.Bool("s")
				if clean == false {
					needCleanBranch(tracking, ignore)
					needCleanTag(tracking, ignore)
					return nil
				}

				fmt.Println("\n🛠start clean tags.")
				cleanTag(tracking, ignore)
				fmt.Println("\n🛠start clean branches.")
				cleanBranch(tracking, ignore)
				saveArchive()
				return nil
			},
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "all",
					Aliases: []string{"a"},
					Value:   false,
					Usage:   "clean all branches which been merged",
				},
				&cli.BoolFlag{
					Name:    "remote",
					Aliases: []string{"r"},
					Value:   false,
					Usage:   "clean remote branches which been merged",
				},
				&cli.BoolFlag{
					Name:    "local",
					Aliases: []string{"l"},
					Value:   false,
					Usage:   "clean local branches which been merged",
				},
				&cli.BoolFlag{
					Name:    "suggest",
					Aliases: []string{"s"},
					Value:   false,
					Usage:   "show branches which been merged without clean",
				},
				&cli.StringFlag{
					Name:    "ignore",
					Aliases: []string{"i"},
					Usage:   "ignore branches which been merged without clean. eg: archive clean branch -i \"feature\\/v[0-9]+.[0-9]+.[0-9]+|master|feature/1.0.0/publish\"",
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:  "branch",
					Usage: "clean branches which been merged",
					Action: func(c *cli.Context) error {
						ignore := c.String("i")
						clean := !c.Bool("s")
						var tracking git.Tracking = git.All
						if c.Bool("r") {
							tracking = git.Remote
						} else if c.Bool("l") {
							tracking = git.Local
						}

						readyArchive("clean_branch")
						excute("git fetch", false)
						if clean == false {
							needCleanBranch(tracking, ignore)
							return nil
						}

						if c.Bool("a") {
							cleanBranch(git.All, ignore)
							saveArchive()
							return nil
						}
						if c.Bool("r") {
							cleanBranch(git.Remote, ignore)
							saveArchive()
							return nil
						}
						if c.Bool("l") {
							cleanBranch(git.Local, ignore)
							saveArchive()
							return nil
						}

						cleanBranch(git.All, ignore)
						saveArchive()
						return nil
					},
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "all",
							Aliases: []string{"a"},
							Value:   false,
							Usage:   "clean all branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "remote",
							Aliases: []string{"r"},
							Value:   false,
							Usage:   "clean remote branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "local",
							Aliases: []string{"l"},
							Value:   false,
							Usage:   "clean local branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "suggest",
							Aliases: []string{"s"},
							Value:   false,
							Usage:   "show branches which been merged without clean",
						},
						&cli.StringFlag{
							Name:    "ignore",
							Aliases: []string{"i"},
							Usage:   "ignore branches which been merged without clean. eg: archive clean branch -i \"feature\\/v[0-9]+.[0-9]+.[0-9]+|master|feature/1.0.0/publish\"",
						},
					},
				},
				{
					Name:  "tag",
					Usage: "clean tag which out of range rule",
					Action: func(c *cli.Context) error {

						ignore := c.String("i")
						clean := !c.Bool("s")
						var tracking git.Tracking = git.All
						if c.Bool("r") {
							tracking = git.Remote
						} else if c.Bool("l") {
							tracking = git.Local
						}

						readyArchive("clean_tag")
						excute("git fetch", false)
						if clean == false {
							needCleanTag(tracking, ignore)
							return nil
						}

						excute("git pull", true)
						if c.Bool("a") {
							cleanTag(git.All, ignore)
							saveArchive()
							return nil
						}
						if c.Bool("r") {
							cleanTag(git.Remote, ignore)
							saveArchive()
							return nil
						}
						if c.Bool("l") {
							cleanTag(git.Local, ignore)
							saveArchive()
							return nil
						}
						cleanTag(tracking, ignore)
						saveArchive()
						return nil
					},
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "all",
							Aliases: []string{"a"},
							Value:   false,
							Usage:   "clean all branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "remote",
							Aliases: []string{"r"},
							Value:   false,
							Usage:   "clean remote branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "local",
							Aliases: []string{"l"},
							Value:   false,
							Usage:   "clean local branches which been merged",
						},
						&cli.BoolFlag{
							Name:    "suggest",
							Aliases: []string{"s"},
							Value:   false,
							Usage:   "show branches which been merged without clean",
						},
						&cli.StringFlag{
							Name:    "ignore",
							Aliases: []string{"i"},
							Usage:   "ignore branches which been merged without clean. eg: archive clean branch -i \"feature\\/v[0-9]+.[0-9]+.[0-9]+|master|feature/1.0.0/publish\"",
						},
					},
				},
			},
		},
		{
			Name:  "backup",
			Usage: "backup branches or tags.",
			Action: func(c *cli.Context) error {
				ignore := c.String("i")
				buildLogger("backup")
				fmt.Println("\n🛠start backup tags.")
				backupTag(git.All, ignore)
				fmt.Println("\n🛠start clean branches.")
				backupBranch(git.All, ignore)
				saveArchive()
				return nil
			},
			Subcommands: []*cli.Command{
				{
					Name:  "branch",
					Usage: "backup branch",
					Action: func(c *cli.Context) error {
						ignore := c.String("i")
						buildLogger("backup branch")
						backupBranch(git.All, ignore)
						return nil
					},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "ignore",
							Aliases: []string{"i"},
							Usage:   "ignore branches. eg: archive backup branch -i \"feature\\/v[0-9]+.[0-9]+.[0-9]+|master|feature/1.0.0/publish\"",
						},
					},
				},
				{
					Name:  "tag",
					Usage: "backup tag",
					Action: func(c *cli.Context) error {
						ignore := c.String("i")
						buildLogger("backup tag")
						backupTag(git.All, ignore)
						return nil
					},
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    "ignore",
							Aliases: []string{"i"},
							Usage:   "ignore tags. eg: archive backup branch -i \"feature\\/v[0-9]+.[0-9]+.[0-9]+|master|feature/1.0.0/publish\"",
						},
					},
				},
			},
		},
		{
			Name:  "abort",
			Usage: "rollback archive which version you given",
			Action: func(c *cli.Context) error {
				fmt.Println("功能暂未开放")
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
				// target := c.String("into")
				vtag := c.String("t")
				test(c.String("i"), vtag)
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "ignore",
					Aliases: []string{"i"},
					Usage:   "ignore branches which been merged without clean",
				},
			},
		},
	}

	// sort.Sort(cli.FlagsByName(app.Flags))
	// sort.Sort(cli.CommandsByName(app.Commands))

}

func backupBranch(tracking git.Tracking, ignore string) {
	list := git.AllBranch(tracking, ignore)
	// fmt.Printf("backup branches:%v\n", list)
	infoJSON, _ := json.Marshal(list)
	// fmt.Printf("backup branches json:%s\n", infoJSON)
	backupPath := path.Join(config.WorkSpace, "backup", tools.String(archiveInfo.Time), "back_branch.json")
	write(infoJSON, backupPath)
}

func backupTag(tracking git.Tracking, ignore string) {
	list := git.AllTag(tracking, ignore)
	// fmt.Printf("backup tags:%v\n", list)
	infoJSON, _ := json.Marshal(list)
	// fmt.Printf("backup tags json:%s\n", infoJSON)
	backupPath := path.Join(config.WorkSpace, "backup", tools.String(archiveInfo.Time), "back_tag.json")
	write(infoJSON, backupPath)
}

func buildLogger(logName string) {
	// logPath = path.Join(config.WorkSpace, "Logs", logName+"_"+strconv.FormatInt(time.Now().Unix(), 10)+".log")
	// dirPath := path.Dir(logPath)
	// mkErr := os.MkdirAll(dirPath, os.ModePerm)
	// if mkErr != nil {
	// 	fmt.Printf("mkdir log folder err : '%s'\n", mkErr)
	// 	return
	// }
	// logfile, err := os.Create(logPath)
	// if err != nil {
	// 	fmt.Printf("create log file err : '%s'\n", mkErr)
	// 	return
	// }

	// logger = log.New(logfile, "", log.LstdFlags|log.Llongfile)
}

func readyArchive(logName string) {
	buildLogger(logName)
	archiveInfo.Log = logPath
	archiveInfo.Tag = tools.String(archiveInfo.Time)
	archiveInfo.User = strings.Trim(gitConfig("user.name"), "\n")
	archiveInfo.Email = strings.Trim(gitConfig("user.email"), "\n")
}

func archive(target string, vtag string) {

	/**
	1.检测命令
	2.同步代码
	3.切换分支


	*/
	checkCMD("git")
	if checkTagAvailable(vtag) == false {
		fmt.Printf("%s is not available, check and retry.\n", vtag)
		log.Fatalf("%s is not available, check and retry.\n", vtag)
		return
	}

	success := merge(target, vtag)
	if success {
		if checkTagLegal(vtag) == false {
			fmt.Printf("Auto merge success, but publish tag(%s) failure. You can use `archive clean` after push tag success.\n", vtag)
			fmt.Printf("Archive '%s' into '%s' success, see more info on:\nlog: '%s'\ninfo: '%s'\n", vtag, target, logPath, archivePath)
			updateVersion()
			return
		}
		publishTag(target, vtag)
		cleanTag(git.All, "")
		cleanBranch(git.All, "")
		fmt.Printf("Archive '%s' into '%s' success, see more info on:\nlog: '%s'\ninfo: '%s'\n", vtag, target, logPath, archivePath)
		updateVersion()
		return
	}
	fmt.Printf("Archive '%s' into '%s' failure, see more info on:\nlog: '%s'\ninfo: '%s'\n", vtag, target, logPath, archivePath)
	updateVersion()
}

func publishTag(branch string, vtag string) {
	excute("git checkout master", false)
	excute("git tag "+vtag, false)
	success, _ := excute("git push origin "+vtag, false)
	if success == false {
		excute("git tag -d "+vtag, false)
	}
	_, info := excute("git show "+vtag, true)
	fmt.Printf("auto publih tag '%s':\n'%s'\n", vtag, info)
}

func checkTagAvailable(vtag string) bool {
	_, info := excute("git cat-file -t "+vtag, true)
	return strings.Contains(info, "fatal: Not a valid object name")
}

func checkTagLegal(vtag string) bool {
	r := regexp.MustCompile("v([0-9]+\\.[0-9]+\\.[0-9])")
	match := r.MatchString(vtag)
	return match
}

func merge(target string, vtag string) bool {
	/**
	1.分支检测，target、from
	2.分支切换 target
	3.代码同步
	4.记录并merge
	5.同步

	//TODO: merge到master后通知vtag之后的版本分支同步master
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
		archiveInfo.Commit = fetchLatestCommit("branch", target, git.Local)
		mergeSuccess, _ := excute("git merge --no-ff "+branch, true)
		if mergeSuccess {
			excute("git push", false)
			archiveInfo.Branches = append(archiveInfo.Branches, &git.Branch{
				Name:     branch,
				Tracking: git.Remote,
				State:    git.Merged,
				Commit:   fetchLatestCommit("branch", branch, git.Remote),
			})
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

func fetchLatestCommit(sort string, info string, tracking git.Tracking) string {
	if sort == "branch" {

		var success bool
		var result string

		if tracking == git.Remote {
			status, resp := excute("git branch -r -v", false)
			success = status
			result = resp
		} else if tracking == git.Local {
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
					reg := regexp.MustCompile(`[\w]+`)
					cmt := reg.FindString(infos)
					fmt.Printf("'%s' '%s' '%s' fetch latest commit : '%s' \n", sort, info, tracking, cmt)
					return cmt
				}
			}
		}
	} else if sort == "tag" {
		success, commitInfo := excute("git show "+info, true)
		if success {
			reg := regexp.MustCompile(`commit [\w]+`)
			resutl := reg.FindString(commitInfo)
			commit := strings.Replace(resutl, "commit ", "", -1)
			return commit
		}
	}
	fmt.Printf("'%s' '%s' '%s' fetch latest commit failure \n", sort, info, tracking)
	return ""
}

func abort(action string, commit string) {
	if action == "branch" {

	} else if action == "tag" {

	} else if action == "merge" {
		excute("git merge --abort", false)
	}
}

func needCleanTag(tracking git.Tracking, ignore string) {
	//TODO: config 忽略版本tag
	if len(ignore) == 0 {
		ignore = "v([0-9]+\\.[0-9]+\\.[0-9])"
	} else {
		ignore = ignore + "|v([0-9]+\\.[0-9]+\\.[0-9])"
	}
	tags := git.AllTag(tracking, ignore)

	if len(tags) > 0 {
		fmt.Println("\n\nThese tag was suggested clean:")
	}
	for _, tag := range tags {
		//TODO: tools 格式化输出，长度补齐
		fmt.Printf("  %s %s %s \n", tag.Tracking, tag.Name, tag.Commit.Commit)
	}
}

func cleanTag(tracking git.Tracking, ignore string) {
	if len(ignore) == 0 {
		ignore = "v([0-9]+\\.[0-9]+\\.[0-9])"
	} else {
		ignore = ignore + "|v([0-9]+\\.[0-9]+\\.[0-9])"
	}
	tags := git.AllTag(tracking, ignore)
	tags = git.DelteTags(tags)
	archiveInfo.Tags = tags
}

func cleanBranch(tracking git.Tracking, ignore string) {
	list := git.MergedBranch(tracking, ignore)
	list = git.DelteBranches(list)
	archiveInfo.Branches = list
}

// 分支清理后再扫描一遍，看下有没有时间过久的分支，提示用户清理掉
func needCleanBranch(tracking git.Tracking, ignore string) {

	merged := git.MergedBranch(tracking, ignore)
	oldest := git.OldestBrnach(tracking, ignore, 14)

	if len(merged) > 0 {
		fmt.Println("\n\nThese merged branches was suggested clean:")
	}
	for _, branch := range merged {
		//TODO: tools 格式化输出，长度补齐
		if branch.State == git.Merged {
			fmt.Printf("  %s %s %s \n", branch.Tracking, branch.Name, branch.Commit)
		} else {
			fmt.Printf("needCleanBranch error logict: unkonw state")
		}
	}

	if len(oldest) > 0 {
		fmt.Println("\n\nThese oldest branches which two weeks not updated was suggested clean:")
	}
	for _, branch := range oldest {
		if branch.State == git.Oldest {
			fmt.Printf("  %s %s %s %s\n", branch.Tracking, branch.Name, branch.Commit, branch.LastDate)
		} else {
			fmt.Printf("needCleanBranch error logict: unkonw state")
		}
	}
}

// 检测时间是是否过期，与当前时间比对， ct：比对时间， gap：间隔，单位小时
func checkOutDate(ct int64, gap float64) bool {
	tm := time.Unix(ct, 0)
	diff := time.Now().Sub(tm).Hours()
	return diff > gap
}

func lock() {

}

//private methods

//tools
func checkCMD(cmd string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		saveArchive()
		log.Fatal(err)
	}
}

func excute(cmdStr string, silent bool) (bool, string) {
	// fmt.Printf("cmd run: '%s'\n", cmdStr)
	fmt.Println(cmdStr)
	branches := strings.Split(cmdStr, " ")
	cmd := exec.Command(branches[0], branches[1:]...)
	// cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		fmt.Println(errStr)
		if silent == false {
			saveArchive()
			//静默处理：正常返回处理结果，不结束程序
			log.Fatal(err)
		}
		return false, errStr
	}
	fmt.Println(outStr)
	//TODO: log 、 notification
	return true, outStr
}

func saveArchive() {
	fmt.Printf("save archive:%v\n", archiveInfo)
	infoJSON, _ := json.Marshal(archiveInfo)
	fmt.Printf("save archive json:%s\n", infoJSON)
	archivePath = path.Join(config.WorkSpace, "backup", archiveInfo.Tag+".json")
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
			saveArchive()
			log.Fatal(mkErr)
			// fmt.Fatal(mkErr)
			return
		}
		writeErr := ioutil.WriteFile(filePath, json, os.ModePerm)
		if writeErr != nil {
			saveArchive()
			log.Fatal(writeErr)
			// fmt.Fatal(writeErr)
			return
		}
		fmt.Printf("✅ save file success：\n", filePath)
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
		fmt.Printf("'%s' no exist.\n", configFile)
		config.WorkSpace = configPath
		config.LatestCheck = time.Now()
		config.UpdateVersion = Day
		config.Version = appVersion
		saveConfig(config)
		fmt.Printf("Default archive config constructor success! You can update it on path '%s'\n", configFile)
		return
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		//Log load config failure
		saveArchive()
		log.Fatal(err)
	}
	var localConfig Config
	err = json.Unmarshal(data, &localConfig)
	if err != nil {
		//Log load config failure
		saveArchive()
		log.Fatal(err)
	}
	config = localConfig
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
		fmt.Printf("check cli version failure(0): '%s'", err)
	} else {
		var data map[string]interface{}
		jsonErr := json.NewDecoder(resp.Body).Decode(&data)
		if jsonErr != nil {
			fmt.Printf("check cli version failure(1): '%s'", jsonErr)
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
	// list := git.MergedBranch(git.All, "v([0-9]+\\.[0-9]+\\.[0-9])")
	// list := git.OldestBrnach(git.Local, "master", 14)
	// list := git.MergedBranch(git.Remote, "master")
	// list = git.DelteBranch(list)
	// for _, branch := range list {
	// fmt.Printf("name:%s ,last:%s\n", branch.Name, branch.LastDate)
	// }

	// list := git.AllBranch(git.All, "v([0-9]+\\.[0-9]+\\.[0-9])")
	// for _, branch := range list {
	// 	fmt.Printf("name:%s ,last:%s\n", branch.Name, branch.LastDate)
	// }

	// list := git.AllTag(git.All, "vd([0-9]+\\.[0-9]+\\.[0-9])")
	// for _, tag := range list {
	// 	fmt.Printf("name:%s ,last:%s\n", tag.Name, tag.LastDate)
	// }

	needCleanTag(git.All, "")

	// commit := git.CommitForID("49701485c554926543bbeac506ddd33ac3849d06")
	// fmt.Printf("commit: %s", commit.Commit)
}

// String value for traking
func String(traking git.Tracking) string {
	switch traking {
	case git.All:
		return "All"
	case git.Local:
		return "Local"
	case git.Remote:
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
	BranchClean   Suggestion
	TagClean      Suggestion
}

//Archive 归档信息
type Archive struct {
	ENV      Config
	Tag      string
	Branch   string
	Commit   string
	User     string
	Email    string
	Branches []*git.Branch
	Tags     []*git.Tag
	Time     int64
	Status   int //0 默认状态，1 已还原，必要时可被删除
	Log      string
}

// Frequency check update frequency
type Frequency string

// 版本更新检测频率
const (
	Day   Frequency = "Day"
	Week  Frequency = "Week"
	Month Frequency = "Month"
)

// Suggestion clean suggestion
type Suggestion string

// 版本更新检测频率
const (
	Prompt Suggestion = "Prompt"
	Clean  Suggestion = "Clean"
)