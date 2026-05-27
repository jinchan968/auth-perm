package constant

const (
	TypeTemplate = "template"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusTemplate  = "template"
)

var validWorkflowStatuses = map[string]bool{
	StatusDraft:     true,
	StatusPublished: true,
	StatusTemplate:  true,
}

func IsValidWorkflowStatus(s string) bool {
	return validWorkflowStatuses[s]
}

var validStatuses = map[string]bool{
	StatusPending:   true,
	StatusRunning:   true,
	StatusSuccess:   true,
	StatusFailed:    true,
	StatusCancelled: true,
}

func IsValidRunStatus(s string) bool {
	return validStatuses[s]
}
