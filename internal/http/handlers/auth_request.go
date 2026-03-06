package handlers

type RegisterRequest struct {
	// New fields
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	RoleKey     string `json:"role_key"`
	Gender      string `json:"gender"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`

	// Backward compatible old field
	Name string `json:"name"`
}