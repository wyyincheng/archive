package git

import (
	"archive/tools"
	"fmt"
	"strings"
)

func init() {

}

// CommitForLog 根据log提取commit信息
func CommitForLog(log string) *Commit {
	//TODO:
	/**
	  1.detail log : 49701485c554926543bbeac506ddd33ac3849d06
	  	"commit [\w]+ ()\nAuthor:空格+*+ <邮箱>\nDate:空格+"
	  2.detail logs: git log
	  3.short  logs: git branch -v
	*/



	// if len(log) > 0 {
	// 	commitInfos := strings.Split(log, "\n")
	// 	for _, commit := range commitInfos {
	// 		trimStr := strings.Trim(strings.Trim(commit, "*"), " ")
	// 		var name = branch.Name
	// 		if branch.Tracking == Remote {
	// 			name = branch.Remote + "/" + branch.Name
	// 		}
	// 		if strings.HasPrefix(trimStr, name) {
	// 			infos := strings.Replace(trimStr, name+" ", "", 1)
	// 			reg := regexp.MustCompile(`[\w]+`)
	// 			cmt := reg.FindString(infos)
	// 			// logger.Printf("'%s' '%s' '%s' fetch latest commit : '%s' \n", sort, info, tracking, cmt)
	// 			branch.Commit = cmt
	// 		}
	// 	}
	// }

	return nil
}

// CommitForID 根据id提取commit信息
func CommitForID(id string) *Commit {
	if len(id) > 6 {
		//7+位id才算有效
		commit := &Commit{}
		commit.Auth = CommitInfo(id, Auth)
		commit.Email = CommitInfo(id, Email)
		commit.Commit = CommitInfo(id, ShortHash)
		commit.Date = CommitInfo(id, Date)
		commit.Desc = CommitInfo(id, Desc)
		commit.Ref = CommitInfo(id, Ref)
		return commit
	}
	return nil
}

// CommitInfo 通过commitId获取对应key值
func CommitInfo(commit string, key Log) string {
	_, info := tools.Excute("git log --pretty=format:“" + fmt.Sprintf("%s", key) + "” " + commit + " -1")
	info = strings.Replace(info, "“", "", -1)
	info = strings.Replace(info, "”", "", -1)
	return info
}

// func fetchLatestCommit(branch *Branch) *Branch {

// 	var success bool
// 	var result string

// 	if branch.Tracking == Remote {
// 		status, resp := tools.Excute("git branch -r -v")
// 		success = status
// 		result = resp
// 	} else if branch.Tracking == Local {
// 		status, resp := tools.Excute("git branch -v")
// 		success = status
// 		result = resp
// 	}

// 	if success == true && len(result) > 0 {
// 		commitInfos := strings.Split(result, "\n")
// 		for _, commit := range commitInfos {
// 			trimStr := strings.Trim(strings.Trim(commit, "*"), " ")
// 			var name = branch.Name
// 			if branch.Tracking == Remote {
// 				name = branch.Remote + "/" + branch.Name
// 			}
// 			if strings.HasPrefix(trimStr, name) {
// 				infos := strings.Replace(trimStr, name+" ", "", 1)
// 				reg := regexp.MustCompile(`[\w]+`)
// 				cmt := reg.FindString(infos)
// 				// logger.Printf("'%s' '%s' '%s' fetch latest commit : '%s' \n", sort, info, tracking, cmt)
// 				branch.Commit = cmt
// 			}
// 		}
// 	}

// 	// logger.Printf("'%s' '%s' '%s' fetch latest commit failure \n", sort, info, tracking)
// 	return branch
// }

// Commit commit信息
type Commit struct {
	Commit string
	Desc   string
	Auth   string
	Email  string
	Date   string
	Ref    string
}

// Log commit的log信息
type Log string

// 分支状态
const (
	Auth      Log = "%cn"
	Email     Log = "%ce"
	Hash      Log = "%H"
	ShortHash Log = "%h"
	Ref       Log = "%d"
	Date      Log = "%cd"
	Desc      Log = "%s"
)

// %an 作者（author）的名字
// %ae 作者的电子邮件地址
// %ad 作者修订日期（可以用 -date= 选项定制格式）
// %ar 作者修订日期，按多久以前的方式显示
// %cn 提交者(committer)的名字
// %ce 提交者的电子邮件地址
// %cd 提交日期
// %cr 提交日期，按多久以前的方式显示
// %s 提交说明
// %d: ref名称
