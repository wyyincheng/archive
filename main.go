package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:  "archive",
		Usage: "archive appstore latest version which has been published.",
		Action: func(c *cli.Context) error {
			fmt.Println("start archive")
			archive(os.Args[0])
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
				Name:    "lock",
				Aliases: []string{"l"},
				Usage:   "lock tag or branch",
				Action: func(c *cli.Context) error {
					fmt.Println("lock tag or branch")
					return nil
				},
			},
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "clean tags and branches after archive",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "abort",
				Aliases: []string{"a"},
				Usage:   "rollback archive which version you given",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Usage:   "test cmd",
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
	merge("master", version)
}

func sync() {
	fetchResult := excute("git", "fetch")
	println(fetchResult)
}

func merge(target string, from string) {
	/**
	1.分支检测，target、from
	2.分支切换 target
	3.代码同步
	4.记录并merge
	5.同步
	*/
	excute("git", "checkout", "-f")
	excute("git", "checkout", target)
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

func excute(name string, arg ...string) string {
	fmt.Printf("cmd run: '%s' '%s'", name, arg)
	cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	if err != nil {
		fmt.Printf("'%s' failed : '%s'", name, errStr)
		log.Fatalf("cmd('%s').Run() failed with %s\n", name, err)
		// os.Exit(600)
	}
	// fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
	//TODO: log 、 notification
	return outStr
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
