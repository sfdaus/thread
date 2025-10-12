package utils

const (
	MaxFileSize         = 5 * 1024 * 1024 // 5 MB per file
	MaxTotalAttachments = 5

	ThreadFileName = "THREAD-ATTACHMENT"
	ThreadFilePath = "threads/"

	// Notification Type
	UPVOTE_THREAD_NOTIFICATION_TYPE = "UPVOTE_THREAD"

	// Notification Reference Type
	UPVOTE_THREAD_NOTIFICATION_REFERENCE_TYPE = "THREAD"

	// Action URL Notification Constant
	UPVOTE_THREAD_NOTIFICATION_ACTION_URL = "/thread/"
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

var ThreadStatus = map[string]string{
	"submitted": "SUBMITTED",
	"drafted":   "DRAFTED",
	"rejected":  "REJECTED",
}

var ThreadNotificationPriority = map[string]string{
	"UPVOTE_THREAD": "medium",
}

var NotificationIdempotencyKey = map[string]string{
	"UPVOTE_THREAD": "notif:upvote",
}

var ThreadNotificationTitle = map[string]string{
	"THREAD_UPVOTE_TITLE": `menyukai postinganmu yang berjudul "[TITLE]"`,
}

var ThreadNotificationMessage = map[string]string{}
