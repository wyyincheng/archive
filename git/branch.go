package git

import (
	"archive/tools"
	"regexp"
	"strings"
)

func init() {

}

// MergedBranch 按正则过滤后的已合并分支列表
func MergedBranch(tracking Tracking, ignore string) []*Branch {
	var mergedResult string
	if tracking == All {
		localArray := MergedBranch(Local, ignore)
		remoteArray := MergedBranch(Remote, ignore)
		return append(localArray, remoteArray...)
	} else if tracking == Local {
		success, resp := tools.Excute("git branch --merged")
		if success == false {
			return nil
		}
		mergedResult = resp
	} else if tracking == Remote {
		success, resp := tools.Excute("git branch -r --merged")
		if success == false {
			return nil
		}
		mergedResult = resp
	}
	return splitBranch(mergedResult, tracking, ignore, Merged)
}

func splitBranch(result string, tracking Tracking, ignore string, state State) []*Branch {
	// //追加默认分支、保护分支
	// ignore = ignore + "|master"
	var resultBranches []*Branch
	resultArray := strings.Split(result, "\n")
	for _, info := range resultArray {
		trimStr := strings.Trim(info, " ")
		branchInfo := strings.Replace(trimStr, "*", "", -1)
		name := strings.Trim(branchInfo, " ")
		if len(name) == 0 {
			continue
		}

		branch := ignoreMatch(name, tracking, ignore)
		if branch.State == Ignore {
			// 		// fmt.Printf("splitBranch ignore branch(%s %s %s) : \n", tracking, branch, commit)
			// 		// logger.Printf("ignore clean branch(%s %s %s) : \n", tracking, branch, commit)
			continue
		}
		branch = fetchLatestCommit(branch)
		if branch != nil {
			resultBranches = append(resultBranches, branch)
		}
	}
	return resultBranches
}

func fetchLatestCommit(branch *Branch) *Branch {

	var success bool
	var result string

	if branch.Tracking == Remote {
		status, resp := tools.Excute("git branch -r -v")
		success = status
		result = resp
	} else if branch.Tracking == Local {
		status, resp := tools.Excute("git branch -v")
		success = status
		result = resp
	}

	if success == true && len(result) > 0 {
		commitInfos := strings.Split(result, "\n")
		for _, commit := range commitInfos {
			trimStr := strings.Trim(strings.Trim(commit, "*"), " ")
			if strings.HasPrefix(trimStr, branch.Name) {
				infos := strings.Replace(trimStr, branch.Name+" ", "", 1)
				reg := regexp.MustCompile(`[\w]+`)
				cmt := reg.FindString(infos)
				// logger.Printf("'%s' '%s' '%s' fetch latest commit : '%s' \n", sort, info, tracking, cmt)
				branch.Commit = cmt
			}
		}
	}

	// logger.Printf("'%s' '%s' '%s' fetch latest commit failure \n", sort, info, tracking)
	return branch
}

// ignore正则匹配：是否符合忽略要求
func ignoreMatch(name string, tracking Tracking, ignore string) *Branch {
	var state State
	var remote string
	var branchName string
	if tracking == All {
		// saveArchive()
		// logger.Fatalf("check branch error: (%s %s)\n", tracking, branch)
		return nil
	} else if tracking == Local {
		branchName = name
		reg := regexp.MustCompile(ignore)
		resutl := reg.FindString(name)
		if resutl == name {
			state = Ignore
		}
	} else if tracking == Remote {
		reg := regexp.MustCompile(`[\w]+`)
		remote = reg.FindString(name)
		branchName = strings.Replace(name, remote+"/", "", 1)

		breg := regexp.MustCompile(ignore)
		resutl := breg.FindString(branchName)
		if resutl == branchName {
			state = Ignore
		}
	}
	return &Branch{
		State:    state,
		Remote:   remote,
		Tracking: tracking,
		Name:     branchName,
	}
}

//Branch 分支信息
type Branch struct {
	Name     string
	Commit   string
	Tracking Tracking
	State    State
	Desc     string
	Remote   string
}
