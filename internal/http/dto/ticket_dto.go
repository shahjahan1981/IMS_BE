package dto

import "time"

type CreateTicketRequest struct {
	TicketID            string `json:"ticket_id"`
	Title               string `json:"title"`
	Type                string `json:"type"`
	Severity            string `json:"severity"`
	Source              string `json:"source"`
	IsEscalated         bool   `json:"is_escalated"`
	AutoResolved        bool   `json:"auto_resolved"`
	DetailedDescription string `json:"detailed_description"`
	CreatedByUserID     int64  `json:"created_by_user_id"`
	HandledByUserID     int64  `json:"handled_by_user_id"`
	TicketTimeIn        string `json:"ticket_time_in"`
	FirstResponseAt     string `json:"first_response_at"`
	TicketTimeOut       string `json:"ticket_time_out"`
}

type ResolveTicketRequest struct {
	TicketID string `json:"ticket_id"`
}

type TicketResponse struct {
	ID                  int64     `json:"id"`
	TicketID            string    `json:"ticket_id"`
	Title               string    `json:"title"`
	Type                string    `json:"type"`
	Severity            string    `json:"severity"`
	Source              string    `json:"source"`
	IsEscalated         bool      `json:"is_escalated"`
	AutoResolved        bool      `json:"auto_resolved"`
	DetailedDescription string    `json:"detailed_description"`
	CreatedByUserID     int64     `json:"created_by_user_id"`
	HandledByUserID     int64     `json:"handled_by_user_id"`
	TicketTimeIn        time.Time `json:"ticket_time_in"`
	FirstResponseAt     time.Time `json:"first_response_at"`
	TicketTimeOut       time.Time `json:"ticket_time_out"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}