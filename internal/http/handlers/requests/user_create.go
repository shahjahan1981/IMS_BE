package requests

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`

	// accept both firstname / first_name
	FirstName string `json:"first_name,omitempty"`
	Firstname string `json:"firstname,omitempty"`

	// accept both lastname / last_name
	LastName string `json:"last_name,omitempty"`
	Lastname string `json:"lastname,omitempty"`

	// accept both rolekey / role_key
	RoleKey string `json:"role_key,omitempty"`
	Rolekey string `json:"rolekey,omitempty"`

	// gender
	Gender string `json:"gender,omitempty"`

	// accept both phonenumber / phone_number
	PhoneNumber string `json:"phone_number,omitempty"`
	Phonenumber string `json:"phonenumber,omitempty"`
}