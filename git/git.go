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
	Merged  State = "Merged"
	Delete  State = "Delete"
	Suggest State = "Suggest"
	Abort   State = "Abort"
	Error   State = "Error"
	Ignore  State = "Ignore"
	Success State = "Success"
	Oldest  State = "Oldest"
)
