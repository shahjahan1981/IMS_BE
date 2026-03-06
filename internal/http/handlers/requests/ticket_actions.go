package requests

type TicketIDRequest struct {
	ID int64 `json:"id"`
}

type AssignTicketRequest struct {
	ID              int64 `json:"id"`
	HandledByUserID int64 `json:"handled_by_user_id"`
}