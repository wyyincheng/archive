package git

func init() {

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
type State string

// 分支状态
const (
	Default State = "Default"
	Merged  State = "Merged"
	Deleted State = "Deleted"
	Suggest State = "Suggest"
	Abort   State = "Abort"
	Error   State = "Error"
	Ignore  State = "Ignore"
	Success State = "Success"
	Oldest  State = "Oldest"
)
