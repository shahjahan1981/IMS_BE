package dto

import "time"

type TicketDTO struct {
	ID                 int64      `json:"id"`
	TicketID           string     `json:"ticket_id"`
	Title              string     `json:"title"`
	Type               string     `json:"type"`
	Severity           string     `json:"severity"`
	Source             string     `json:"source"`
	IsEscalated        bool       `json:"is_escalated"`
	AutoResolved       bool       `json:"auto_resolved"`
	DetailedDesc       *string    `json:"detailed_description,omitempty"`
	CreatedByUserID    int64      `json:"created_by_user_id"`
	HandledByUserID    *int64     `json:"handled_by_user_id,omitempty"`
	TicketTimeIn       time.Time  `json:"ticket_time_in"`
	TicketTimeOut      *time.Time `json:"ticket_time_out,omitempty"`
	FirstResponseAt    *time.Time `json:"first_response_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}