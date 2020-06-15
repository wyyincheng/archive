package git

func init()  {
	
}

func MergedBranch() []*Branch {

	return nil
}

//Branch 分支信息
type Branch struct {
	Name     string
	Commit   string
	Tracking Tracking
	State    State
	Desc     string
}