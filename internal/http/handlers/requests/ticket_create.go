package requests

type CreateTicketRequest struct {
	TicketID           string `json:"ticket_id"`
	Title              string `json:"title"`
	Type               string `json:"type"`
	Severity           string `json:"severity"`
	Source             string `json:"source"`
	IsEscalated        bool   `json:"is_escalated"`
	AutoResolved       bool   `json:"auto_resolved"`
	DetailedDesc       string `json:"detailed_description"`
	CreatedByUserID    int64  `json:"created_by_user_id"`
	HandledByUserID    *int64 `json:"handled_by_user_id,omitempty"`
}