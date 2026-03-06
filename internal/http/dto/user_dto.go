package dto

import "time"

type UserDTO struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	RoleKey     string     `json:"role_key"`
	Gender      *string    `json:"gender,omitempty"`
	Email       string     `json:"email"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	IsOnline    bool       `json:"is_online"`
	CreatedAt   time.Time  `json:"created_at"`
}