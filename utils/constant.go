package utils

const (
	MaxFileSize         = 5 * 1024 * 1024 // 5 MB per file
	MaxTotalAttachments = 5
)

type ResponseStatus struct {
	Success string
	Failed  string
	Error   string
}

var Status = ResponseStatus{
	Success: "success",
	Failed:  "failed",
	Error:   "error",
}
