package git

import (
	"archive/tools"
	"fmt"
	"regexp"
	"strings"
)

func init() {

}

// func checkTagAvailable(vtag string) bool {
// 	_, info := excute("git cat-file -t "+vtag, true)
// 	return strings.Contains(info, "fatal: Not a valid object name")
// }

// func checkTagLegal(vtag string) bool {
// 	r := regexp.MustCompile("v([0-9]+\\.[0-9]+\\.[0-9])")
// 	match := r.MatchString(vtag)
// 	return match
// }

// DelteTags 删除一组tag
func DelteTags(list []*Tag) []*Tag {
	var deltedList []*Tag
	for _, tag := range list {
		deleted := DeleteTag(tag)
		deltedList = append(deltedList, deleted)
	}
	return deltedList
}

// DeleteTag 删除指定tag
func DeleteTag(tag *Tag) *Tag {
	if tag.State != Ignore {
		success, info := tools.Excute("git tag -d " + tag.Name)
		if success {
			pSuccess, result := tools.Excute("git push origin :" + tag.Name)
			if pSuccess {
				fmt.Printf(" ✅ delete tag success(%s %s) : \n", tag.Tracking, tag.Name)
				tag.State = Deleted
			} else {
				tag.State = Error
				tag.Desc = "Delete Failure::" + result
				fmt.Printf(" ❌ delete tag failure(%s %s) : %s\n", tag.Tracking, tag.Name, result)
			}
		} else {
			tag.State = Error
			tag.Desc = "Delete Failure:" + info
			fmt.Printf(" ❌ delete tag failure(%s %s) : %s\n", tag.Tracking, tag.Name, info)
		}
	}
	return tag
}

// AllTag 按正则过滤后的所有标签
func AllTag(tracking Tracking, ignore string) []*Tag {
	var result string
	if tracking == All {
		localArray := AllTag(Local, ignore)
		remoteArray := AllTag(Remote, ignore)
		return append(localArray, remoteArray...)
	} else if tracking == Local {
		success, resp := tools.Excute("git tag -l")
		if success == false {
			return nil
		}
		result = resp
	} else if tracking == Remote {
		success, resp := tools.Excute("git ls-remote --tags")
		if success == false {
			return nil
		}
		result = resp
	}
	return splitTag(result, tracking, ignore, Default)
}

func splitTag(result string, tracking Tracking, ignore string, state State) []*Tag {
	//追加保护tag config.tagRule
	// ignore = ignore + "|" + config.tagRule
	var resultTags []*Tag
	resultArray := strings.Split(result, "\n")
	for _, info := range resultArray {
		var tagName string
		var commit *Commit
		if tracking == Remote {
			if strings.HasPrefix(info, "From ") == false && len(info) > 0 {
				list := strings.Split(info, "refs/tags/")
				commitID := tools.TrimBoth(list[0])
				commit = CommitForID(commitID)
				tagName = list[len(list)-1]
			} else {
				continue
			}
		} else if tracking == Local {
			tagName = info
			success, log := tools.Excute("git show " + tagName)
			if success {
				commit = CommitForLog(log)
			}
		}

		tag := ignoreTagMatch(tagName, tracking, ignore)
		if tag.State == Ignore {
			// 		// fmt.Printf("splitBranch ignore branch(%s %s %s) : \n", tracking, branch, commit)
			// 		// logger.Printf("ignore clean branch(%s %s %s) : \n", tracking, branch, commit)
			continue
		}
		tag.State = state
		tag.Commit = commit
		if tag != nil && len(tag.Name) > 0 {
			resultTags = append(resultTags, tag)
		}
	}
	return resultTags
}

// ignore正则匹配：是否符合忽略要求
func ignoreTagMatch(name string, tracking Tracking, ignore string) *Tag {
	var state State
	if tracking == All {
		// saveArchive()
		// logger.Fatalf("check branch error: (%s %s)\n", tracking, branch)
		return nil
	}

	r := regexp.MustCompile(ignore)
	match := r.MatchString(name)
	if match == true {
		state = Ignore
	}

	return &Tag{
		Name:     name,
		State:    state,
		Tracking: tracking,
	}
}

//Tag 标签信息
type Tag struct {
	Name     string
	Commit   *Commit
	Tracking Tracking
	State    State
	Desc     string
	LastDate string
}
